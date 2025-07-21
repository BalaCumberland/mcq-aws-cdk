package handlers

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func HandleQuizDeleteV2(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Check admin permissions
	_, err := CheckAdminRole(request)
	if err != nil {
		log.Printf("❌ Permission denied: %v", err)
		return CreateErrorResponse(403, err.Error()), nil
	}
	
	quizName := request.QueryStringParameters["quizName"]
	if quizName == "" {
		return CreateErrorResponse(400, "Missing 'quizName' parameter"), nil
	}

	log.Printf("📌 Deleting quiz: %s", quizName)

	// Delete quiz
	_, err = dynamoClient.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("quiz_questions"),
		Key: map[string]*dynamodb.AttributeValue{
			"quiz_name": {S: aws.String(quizName)},
		},
	})

	if err != nil {
		log.Printf("❌ Error deleting quiz: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	// Delete all attempt records for this quiz
	scanResult, scanErr := dynamoClient.Scan(&dynamodb.ScanInput{
		TableName: aws.String("student_quiz_attempts"),
		FilterExpression: aws.String("quiz_name = :quiz_name"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":quiz_name": {S: aws.String(quizName)},
		},
	})

	if scanErr == nil {
		for _, item := range scanResult.Items {
			if email := item["email"]; email != nil && email.S != nil {
				_, _ = dynamoClient.DeleteItem(&dynamodb.DeleteItemInput{
					TableName: aws.String("student_quiz_attempts"),
					Key: map[string]*dynamodb.AttributeValue{
						"email": {S: email.S},
						"quiz_name": {S: aws.String(quizName)},
					},
				})
			}
		}
	}

	response := map[string]interface{}{
		"message":  "Quiz deleted successfully",
		"quizName": quizName,
	}

	responseJSON, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}