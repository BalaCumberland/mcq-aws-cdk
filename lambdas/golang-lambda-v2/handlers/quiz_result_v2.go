package handlers

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func HandleQuizResultV2(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get email from Firebase auth context
	email, err := GetUserFromContext(request)
	if err != nil {
		return CreateErrorResponse(401, "Unauthorized"), nil
	}

	quizName := request.QueryStringParameters["quizName"]
	className := request.QueryStringParameters["className"]
	subjectName := request.QueryStringParameters["subjectName"]
	topic := request.QueryStringParameters["topic"]
	
	if quizName == "" {
		return CreateErrorResponse(400, "Missing 'quizName' parameter"), nil
	}
	if className == "" {
		return CreateErrorResponse(400, "Missing 'className' parameter"), nil
	}
	if subjectName == "" {
		return CreateErrorResponse(400, "Missing 'subjectName' parameter"), nil
	}
	if topic == "" {
		return CreateErrorResponse(400, "Missing 'topic' parameter"), nil
	}

	email = strings.ToLower(email)
	log.Printf("📌 Fetching result for: %s, Quiz: %s (%s-%s-%s)", email, quizName, className, subjectName, topic)

	// Get quiz attempt using simple key lookup
	result, err := dynamoClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("student_quiz_attempts"),
		Key: map[string]*dynamodb.AttributeValue{
			"email":     {S: aws.String(email)},
			"quiz_name": {S: aws.String(quizName)},
		},
	})

	if err != nil {
		log.Printf("❌ Error fetching result: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	if result.Item == nil {
		return CreateErrorResponse(404, "Quiz result not found"), nil
	}

	var attempt AttemptItem
	err = dynamodbattribute.UnmarshalMap(result.Item, &attempt)


	if err != nil {
		log.Printf("❌ Error unmarshaling result: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	response := map[string]interface{}{
		"message":       "Result fetched successfully",
		"quizName":      attempt.QuizName,
		"correctCount":  attempt.CorrectCount,
		"wrongCount":    attempt.WrongCount,
		"skippedCount":  attempt.SkippedCount,
		"totalCount":    attempt.TotalCount,
		"percentage":    attempt.Percentage,
		"attemptNumber": attempt.AttemptNumber,
		"attemptedAt":   attempt.AttemptedAt,
		"results":       attempt.Results,
	}

	responseJSON, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}