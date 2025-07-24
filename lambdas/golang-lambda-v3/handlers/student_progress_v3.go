package handlers

import (
	"encoding/json"
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
	UID             string                       `json:"uid"`
	CategorySummary []ProgressSummary            `json:"categorySummary"`
	IndividualTests map[string][]TestScore       `json:"individualTests"`
}

func HandleStudentProgressV3(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	uid, err := GetUserFromContext(request)
	if err != nil {
		return CreateErrorResponse(401, "Unauthorized"), nil
	}

	student, err := GetStudentFromDynamoDB(uid)
	if err != nil || student == nil {
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
		TableName: aws.String("student_quiz_attempts_v3"),
		KeyConditionExpression: aws.String("uid = :uid"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":uid": {S: aws.String(uid)},
		},
	})

	if err != nil {
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	// Create maps to track category stats
	attemptedQuizzes := make(map[string]map[string]bool)
	percentageSum := make(map[string]float64)
	percentageCount := make(map[string]int)
	individualTests := make(map[string][]TestScore)

	// Process attempts
	for _, item := range result.Items {
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

		// Update category stats
		if attemptedQuizzes[category] == nil {
			attemptedQuizzes[category] = make(map[string]bool)
		}
		attemptedQuizzes[category][quizName] = true
		percentageSum[category] += percentage
		percentageCount[category]++

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
		UID:             uid,
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