package handlers

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type ProgressSummary struct {
	Category     string  `json:"category"`
	Percentage   float64 `json:"percentage"`
	Attempted    int     `json:"attempted"`
	Unattempted  int     `json:"unattempted"`
}

type TestScore struct {
	QuizName       string  `json:"quizName"`
	Category       string  `json:"category"`
	CorrectCount   int     `json:"correctCount"`
	WrongCount     int     `json:"wrongCount"`
	SkippedCount   int     `json:"skippedCount"`
	TotalCount     int     `json:"totalCount"`
	Percentage     float64 `json:"percentage"`
	TotalAttempts  int     `json:"totalAttempts"`
	LatestScore    float64 `json:"latestScore"`
	AttemptedAt    string  `json:"attemptedAt"`
}

type ProgressResponse struct {
	Email           string                       `json:"email"`
	CategorySummary []ProgressSummary            `json:"categorySummary"`
	IndividualTests map[string][]TestScore       `json:"individualTests"`
}

func HandleStudentProgressV2(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	email, err := GetUserFromContext(request)
	if err != nil {
		return CreateErrorResponse(401, "Unauthorized"), nil
	}

	// Get student's enrolled subjects
	student, err := GetStudentFromDynamoDB(email)
	if err != nil || student == nil {
		log.Printf("❌ Student not found: %v", err)
		return CreateErrorResponse(404, "Student not found"), nil
	}

	// Get enrolled subjects for this student class
	var enrolledSubjects []string
	for _, category := range VALID_CATEGORIES {
		if strings.HasPrefix(category, student.StudentClass) {
			enrolledSubjects = append(enrolledSubjects, category)
		}
	}

	if len(enrolledSubjects) == 0 {
		return CreateErrorResponse(404, "No subjects found for student class"), nil
	}

	// Get all attempts for student
	result, err := dynamoClient.Query(&dynamodb.QueryInput{
		TableName: aws.String("student_quiz_attempts"),
		KeyConditionExpression: aws.String("email = :email"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":email": {S: aws.String(email)},
		},
	})

	if err != nil {
		log.Printf("❌ Error querying attempts: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	// Create maps to track category stats
	attemptedQuizzes := make(map[string]map[string]bool) // category -> quiz names
	percentageSum := make(map[string]float64)
	percentageCount := make(map[string]int)
	individualTests := make(map[string][]TestScore)

	// Process attempts
	log.Printf("📊 Found %d attempts in DynamoDB", len(result.Items))
	for _, item := range result.Items {
		log.Printf("📊 Processing item: %+v", item)
		var category, quizName, attemptedAt string
		var correctCount, wrongCount, skippedCount, totalCount, attemptNumber int
		var percentage float64
		
		if cat := item["category"]; cat != nil && cat.S != nil {
			category = *cat.S
		}
		if qn := item["quiz_name"]; qn != nil && qn.S != nil {
			quizName = *qn.S
		}
		if cc := item["correct_count"]; cc != nil && cc.N != nil {
			correctCount, _ = strconv.Atoi(*cc.N)
		}
		if wc := item["wrong_count"]; wc != nil && wc.N != nil {
			wrongCount, _ = strconv.Atoi(*wc.N)
		}
		if sc := item["skipped_count"]; sc != nil && sc.N != nil {
			skippedCount, _ = strconv.Atoi(*sc.N)
		}
		if tc := item["total_count"]; tc != nil && tc.N != nil {
			totalCount, _ = strconv.Atoi(*tc.N)
		}
		if pct := item["percentage"]; pct != nil {
			if pct.N != nil {
				percentage, _ = strconv.ParseFloat(*pct.N, 64)
			} else if pct.S != nil {
				percentage, _ = strconv.ParseFloat(*pct.S, 64)
			}
			log.Printf("📊 Quiz: %s, Percentage: %f", quizName, percentage)
		}
		if an := item["attempt_number"]; an != nil && an.N != nil {
			attemptNumber, _ = strconv.Atoi(*an.N)
		}
		if at := item["attempted_at"]; at != nil && at.S != nil {
			attemptedAt = *at.S
		}

		// Only include enrolled subjects
		enrolled := false
		for _, subject := range enrolledSubjects {
			if category == subject {
				enrolled = true
				break
			}
		}
		if !enrolled {
			continue
		}

		// Update category stats - count each quiz attempt separately
		if attemptedQuizzes[category] == nil {
			attemptedQuizzes[category] = make(map[string]bool)
		}
		attemptedQuizzes[category][quizName] = true
		percentageSum[category] += percentage
		percentageCount[category]++
		log.Printf("📊 Added %s percentage: %f (total: %f, count: %d)", category, percentage, percentageSum[category], percentageCount[category])

		// Round percentage to 1 decimal place
		roundedPercentage := float64(int(percentage*10+0.5)) / 10
		
		// Add to individual tests
		test := TestScore{
			QuizName:      quizName,
			Category:      category,
			CorrectCount:  correctCount,
			WrongCount:    wrongCount,
			SkippedCount:  skippedCount,
			TotalCount:    totalCount,
			Percentage:    roundedPercentage,
			TotalAttempts: attemptNumber,
			LatestScore:   roundedPercentage,
			AttemptedAt:   attemptedAt,
		}
		individualTests[category] = append(individualTests[category], test)
	}

	// Create category summary for all enrolled subjects
	var categorySummary []ProgressSummary
	for _, category := range enrolledSubjects {
		// Get total quiz count for this category
		totalQuizzes := 0
		quizResult, err := dynamoClient.Scan(&dynamodb.ScanInput{
			TableName: aws.String("quiz_questions"),
			FilterExpression: aws.String("category = :category"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":category": {S: aws.String(category)},
			},
			Select: aws.String("COUNT"),
		})
		if err == nil {
			totalQuizzes = int(*quizResult.Count)
		}

		attempted := len(attemptedQuizzes[category])
		unattempted := totalQuizzes - attempted
		
		// Calculate average percentage and round to 1 decimal
		var avgPercentage float64
		if percentageCount[category] > 0 {
			avgPercentage = percentageSum[category] / float64(percentageCount[category])
			avgPercentage = float64(int(avgPercentage*10+0.5)) / 10 // Round to 1 decimal
		}

		categorySummary = append(categorySummary, ProgressSummary{
			Category:    category,
			Percentage:  avgPercentage,
			Attempted:   attempted,
			Unattempted: unattempted,
		})
	}

	response := ProgressResponse{
		Email:           email,
		CategorySummary: categorySummary,
		IndividualTests: individualTests,
	}

	responseJSON, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}