package handlers

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func HandleQuizDeleteV3(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	_, err := CheckAdminRole(request)
	if err != nil {
		return CreateErrorResponse(403, err.Error()), nil
	}
	
	quizName := request.QueryStringParameters["quizName"]
	if quizName == "" {
		return CreateErrorResponse(400, "Missing 'quizName' parameter"), nil
	}

	// Check if quiz exists first
	getResult, err := dynamoClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("quiz_questions"),
		Key: map[string]*dynamodb.AttributeValue{
			"quiz_name": {S: aws.String(quizName)},
		},
	})

	if err != nil {
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	if getResult.Item == nil {
		return CreateErrorResponse(404, "Quiz not found"), nil
	}

	// Delete quiz
	_, err = dynamoClient.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("quiz_questions"),
		Key: map[string]*dynamodb.AttributeValue{
			"quiz_name": {S: aws.String(quizName)},
		},
	})

	if err != nil {
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	// Delete all attempt records for this quiz
	scanResult, scanErr := dynamoClient.Scan(&dynamodb.ScanInput{
		TableName: aws.String("student_quiz_attempts_v3"),
		FilterExpression: aws.String("quiz_name = :quiz_name"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":quiz_name": {S: aws.String(quizName)},
		},
	})

	if scanErr == nil {
		for _, item := range scanResult.Items {
			if uid := item["uid"]; uid != nil && uid.S != nil {
				_, _ = dynamoClient.DeleteItem(&dynamodb.DeleteItemInput{
					TableName: aws.String("student_quiz_attempts_v3"),
					Key: map[string]*dynamodb.AttributeValue{
						"uid": {S: uid.S},
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