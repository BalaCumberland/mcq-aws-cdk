package handlers

import (
	"encoding/json"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func HandleUnattemptedQuizzesV3(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	uid, err := GetUserFromContext(request)
	if err != nil {
		return CreateErrorResponse(401, "Unauthorized"), nil
	}

	student, err := GetStudentFromDynamoDB(uid)
	if err != nil || student == nil {
		return CreateErrorResponse(404, "Student not found"), nil
	}

	// Get all quizzes for student's class
	var enrolledSubjects []string
	for _, category := range VALID_CATEGORIES {
		if strings.HasPrefix(category, student.StudentClass) {
			enrolledSubjects = append(enrolledSubjects, category)
		}
	}

	// Get all attempted quizzes
	attemptedQuizzes := make(map[string]bool)
	queryResult, err := dynamoClient.Query(&dynamodb.QueryInput{
		TableName: aws.String("student_quiz_attempts_v3"),
		KeyConditionExpression: aws.String("uid = :uid"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":uid": {S: aws.String(uid)},
		},
		ProjectionExpression: aws.String("quiz_name"),
	})

	if err == nil {
		for _, item := range queryResult.Items {
			if quizName := item["quiz_name"]; quizName != nil && quizName.S != nil {
				attemptedQuizzes[*quizName.S] = true
			}
		}
	}

	// Get all available quizzes for enrolled subjects
	var unattemptedQuizzes []string
	for _, category := range enrolledSubjects {
		scanResult, err := dynamoClient.Scan(&dynamodb.ScanInput{
			TableName: aws.String("quiz_questions"),
			FilterExpression: aws.String("category = :category"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":category": {S: aws.String(category)},
			},
		})

		if err == nil {
			for _, item := range scanResult.Items {
				if quizName := item["quiz_name"]; quizName != nil && quizName.S != nil {
					unattemptedQuizzes = append(unattemptedQuizzes, *quizName.S)
				}
			}
		}
	}

	response := map[string]interface{}{
		"unattempted_quizzes": unattemptedQuizzes,
	}

	responseJSON, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}