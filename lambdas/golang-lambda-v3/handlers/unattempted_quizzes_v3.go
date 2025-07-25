package handlers

import (
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func HandleUnattemptedQuizzesV3(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	uid, err := GetUserFromContext(request)
	if err != nil {
		return CreateErrorResponse(401, "Unauthorized"), nil
	}

	category := request.QueryStringParameters["category"]
	log.Printf("📌 Category parameter: %s", category)
	if category == "" {
		return CreateErrorResponse(400, "Missing 'category' parameter"), nil
	}
	log.Printf("📌 Fetching unattempted quizzes for UID: %s, Category: %s", uid, category)

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

	// Get all quizzes in category
	scanResult, err := dynamoClient.Scan(&dynamodb.ScanInput{
		TableName: aws.String("quiz_questions"),
		FilterExpression: aws.String("category = :category"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":category": {S: aws.String(category)},
		},
	})

	if err != nil {
		log.Printf("❌ Error scanning quizzes: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	log.Printf("📊 Found %d items in scan result", len(scanResult.Items))

	// Return all quizzes (allow retakes)
	var unattemptedQuizzes []string
	for _, item := range scanResult.Items {
		if quizName := item["quiz_name"]; quizName != nil && quizName.S != nil {
			unattemptedQuizzes = append(unattemptedQuizzes, *quizName.S)
			log.Printf("📝 Quiz: %s", *quizName.S)
		}
	}

	log.Printf("🎯 Returning %d unattempted quizzes", len(unattemptedQuizzes))

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