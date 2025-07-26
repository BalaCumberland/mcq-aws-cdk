package handlers

import (
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func HandleQuizResultV2(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get UID from Firebase auth context
	uid, err := GetUserUIDFromContext(request)
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

	log.Printf("📌 Fetching result for: %s, Quiz: %s (%s-%s-%s)", uid, quizName, className, subjectName, topic)

	// Get quiz attempt using simple key lookup
	result, err := dynamoClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("student_quiz_attempts_v2"),
		Key: map[string]*dynamodb.AttributeValue{
			"uid":       {S: aws.String(uid)},
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