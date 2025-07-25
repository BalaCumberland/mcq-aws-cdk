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
	ClassName    string  `json:"className"`
	SubjectName  string  `json:"subjectName"`
	Percentage   float64 `json:"percentage"`
	Attempted    int     `json:"attempted"`
	Unattempted  int     `json:"unattempted"`
}

type TestScore struct {
	QuizName       string  `json:"quizName"`
	ClassName      string  `json:"className"`
	SubjectName    string  `json:"subjectName"`
	Topic          string  `json:"topic"`
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
	SubjectSummary  []ProgressSummary            `json:"subjectSummary"`
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

	// Get enrolled subjects for this student class from class_subjects table
	subjects, err := FetchSubjects(student.StudentClass)
	if err != nil || len(subjects) == 0 {
		log.Printf("❌ No subjects found for class %s: %v", student.StudentClass, err)
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

	// Create maps to track subject stats
	attemptedQuizzes := make(map[string]map[string]bool) // subject -> quiz names
	percentageSum := make(map[string]float64)
	percentageCount := make(map[string]int)
	individualTests := make(map[string][]TestScore)

	// Process attempts
	log.Printf("📊 Found %d attempts in DynamoDB", len(result.Items))
	for _, item := range result.Items {
		var className, subjectName, topic, quizName, attemptedAt string
		var correctCount, wrongCount, skippedCount, totalCount, attemptNumber int
		var percentage float64
		
		if cn := item["class_name"]; cn != nil && cn.S != nil {
			className = *cn.S
		}
		if sn := item["category"]; sn != nil && sn.S != nil {
			subjectName = *sn.S
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

		// Only include student's class and enrolled subjects
		if className != student.StudentClass {
			continue
		}
		enrolled := false
		for _, subject := range subjects {
			if subjectName == subject {
				enrolled = true
				break
			}
		}
		if !enrolled {
			continue
		}

		// Update subject stats
		if attemptedQuizzes[subjectName] == nil {
			attemptedQuizzes[subjectName] = make(map[string]bool)
		}
		attemptedQuizzes[subjectName][quizName] = true
		percentageSum[subjectName] += percentage
		percentageCount[subjectName]++

		// Round percentage to 1 decimal place
		roundedPercentage := float64(int(percentage*10+0.5)) / 10
		
		// Add to individual tests
		test := TestScore{
			QuizName:      quizName,
			ClassName:     className,
			SubjectName:   subjectName,
			Topic:         topic,
			CorrectCount:  correctCount,
			WrongCount:    wrongCount,
			SkippedCount:  skippedCount,
			TotalCount:    totalCount,
			Percentage:    roundedPercentage,
			TotalAttempts: attemptNumber,
			LatestScore:   roundedPercentage,
			AttemptedAt:   attemptedAt,
		}
		individualTests[subjectName] = append(individualTests[subjectName], test)
	}

	// Create subject summary for all enrolled subjects
	var subjectSummary []ProgressSummary
	for _, subject := range subjects {
		// Get total quiz count for this class and subject
		totalQuizzes := 0
		quizResult, err := dynamoClient.Scan(&dynamodb.ScanInput{
			TableName: aws.String("quiz_questions"),
			FilterExpression: aws.String("class_name = :className AND subject_name = :subjectName"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":className":   {S: aws.String(student.StudentClass)},
				":subjectName": {S: aws.String(subject)},
			},
			Select: aws.String("COUNT"),
		})
		if err == nil {
			totalQuizzes = int(*quizResult.Count)
		}

		attempted := len(attemptedQuizzes[subject])
		unattempted := totalQuizzes - attempted
		
		// Calculate average percentage and round to 1 decimal
		var avgPercentage float64
		if percentageCount[subject] > 0 {
			avgPercentage = percentageSum[subject] / float64(percentageCount[subject])
			avgPercentage = float64(int(avgPercentage*10+0.5)) / 10 // Round to 1 decimal
		}

		subjectSummary = append(subjectSummary, ProgressSummary{
			ClassName:    student.StudentClass,
			SubjectName:  subject,
			Percentage:   avgPercentage,
			Attempted:    attempted,
			Unattempted:  unattempted,
		})
	}

	response := ProgressResponse{
		Email:           email,
		SubjectSummary:  subjectSummary,
		IndividualTests: individualTests,
	}

	responseJSON, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}