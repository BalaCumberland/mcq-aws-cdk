package handlers

import (
	"encoding/json"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func HandleQuizResultV3(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	uid, err := GetUserFromContext(request)
	if err != nil {
		return CreateErrorResponse(401, "Unauthorized"), nil
	}

	quizName := request.QueryStringParameters["quizName"]
	if quizName == "" {
		return CreateErrorResponse(400, "Missing 'quizName' parameter"), nil
	}

	// Get latest attempt for this quiz
	queryResult, err := dynamoClient.Query(&dynamodb.QueryInput{
		TableName: aws.String("student_quiz_attempts_v3"),
		KeyConditionExpression: aws.String("uid = :uid AND quiz_name = :quiz_name"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":uid": {S: aws.String(uid)},
			":quiz_name": {S: aws.String(quizName)},
		},
		ScanIndexForward: aws.Bool(false),
		Limit: aws.Int64(1),
	})

	if err != nil {
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	if len(queryResult.Items) == 0 {
		return CreateErrorResponse(404, "No attempt found for this quiz"), nil
	}

	var attempt AttemptItem
	err = dynamodbattribute.UnmarshalMap(queryResult.Items[0], &attempt)
	if err != nil {
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	// Round percentage to 1 decimal place
	if percentageFloat, ok := attempt.Percentage.(float64); ok {
		attempt.Percentage = float64(int(percentageFloat*10+0.5)) / 10
	} else if percentageStr, ok := attempt.Percentage.(string); ok {
		if pct, err := strconv.ParseFloat(percentageStr, 64); err == nil {
			attempt.Percentage = float64(int(pct*10+0.5)) / 10
		}
	}

	responseJSON, _ := json.Marshal(attempt)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}