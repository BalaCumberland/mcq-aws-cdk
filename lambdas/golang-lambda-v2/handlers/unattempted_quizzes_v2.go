package handlers

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func HandleUnattemptedQuizzesV2(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	email := request.QueryStringParameters["email"]
	className := request.QueryStringParameters["className"]
	subjectName := request.QueryStringParameters["subjectName"]
	topic := request.QueryStringParameters["topic"] // Optional

	if email == "" {
		return CreateErrorResponse(400, "Missing 'email' parameter"), nil
	}
	if className == "" {
		return CreateErrorResponse(400, "Missing 'className' parameter"), nil
	}
	if subjectName == "" {
		return CreateErrorResponse(400, "Missing 'subjectName' parameter"), nil
	}

	email = strings.ToLower(email)
	log.Printf("📌 Fetching unattempted quizzes for: %s, Class: %s, Subject: %s, Topic: %s", email, className, subjectName, topic)

	// Build filter expression based on parameters
	var filterExpression string
	var expressionAttributeValues map[string]*dynamodb.AttributeValue

	if topic != "" {
		// Filter by class, subject, and topic
		filterExpression = "class_name = :className AND subject_name = :subjectName AND topic = :topic"
		expressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":className":   {S: aws.String(className)},
			":subjectName": {S: aws.String(subjectName)},
			":topic":       {S: aws.String(topic)},
		}
	} else {
		// Filter by class and subject only
		filterExpression = "class_name = :className AND subject_name = :subjectName"
		expressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":className":   {S: aws.String(className)},
			":subjectName": {S: aws.String(subjectName)},
		}
	}

	// Get all quizzes matching criteria
	result, err := dynamoClient.Scan(&dynamodb.ScanInput{
		TableName:                 aws.String("quiz_questions"),
		FilterExpression:          aws.String(filterExpression),
		ExpressionAttributeValues: expressionAttributeValues,
	})

	if err != nil {
		log.Printf("❌ Error scanning quizzes: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	// Get attempted quizzes for student
	attemptResult, err := dynamoClient.Query(&dynamodb.QueryInput{
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

	// Create map of attempted quiz names
	attemptedQuizzes := make(map[string]bool)
	for _, item := range attemptResult.Items {
		if quizName := item["quiz_name"]; quizName != nil && quizName.S != nil {
			attemptedQuizzes[*quizName.S] = true
		}
	}

	// Return all quizzes (allow retakes)
	var unattemptedQuizzes []string
	for _, item := range result.Items {
		if quizName := item["quiz_name"]; quizName != nil && quizName.S != nil {
			unattemptedQuizzes = append(unattemptedQuizzes, *quizName.S)
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