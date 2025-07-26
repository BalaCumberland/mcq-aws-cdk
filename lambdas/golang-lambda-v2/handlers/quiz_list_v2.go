package handlers

import (
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type QuizListItem struct {
	QuizName    string      `json:"quizName" dynamodbav:"quiz_name"`
	ClassName   string      `json:"className" dynamodbav:"class_name"`
	SubjectName string      `json:"subjectName" dynamodbav:"subject_name"`
	Topic       string      `json:"topic" dynamodbav:"topic"`
	Duration    interface{} `json:"duration" dynamodbav:"duration"`
}

func HandleQuizListV2(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Check admin permissions
	_, err := CheckAdminRole(request)
	if err != nil {
		return CreateErrorResponse(403, err.Error()), nil
	}

	className := request.QueryStringParameters["className"]
	subjectName := request.QueryStringParameters["subjectName"]
	topic := request.QueryStringParameters["topic"]

	if className == "" {
		return CreateErrorResponse(400, "Missing 'className' parameter"), nil
	}
	if subjectName == "" {
		return CreateErrorResponse(400, "Missing 'subjectName' parameter"), nil
	}
	if topic == "" {
		return CreateErrorResponse(400, "Missing 'topic' parameter"), nil
	}

	log.Printf("📌 Listing quizzes for: %s-%s-%s", className, subjectName, topic)

	// Scan for matching quizzes
	result, err := dynamoClient.Scan(&dynamodb.ScanInput{
		TableName:        aws.String("quiz_questions"),
		FilterExpression: aws.String("class_name = :className AND subject_name = :subjectName AND topic = :topic"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":className":   {S: aws.String(className)},
			":subjectName": {S: aws.String(subjectName)},
			":topic":       {S: aws.String(topic)},
		},
	})

	if err != nil {
		log.Printf("❌ Error scanning quizzes: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	var quizzes []QuizListItem
	for _, item := range result.Items {
		var quiz QuizListItem
		err = dynamodbattribute.UnmarshalMap(item, &quiz)
		if err != nil {
			log.Printf("❌ Error unmarshaling quiz: %v", err)
			continue
		}
		quizzes = append(quizzes, quiz)
	}

	response := map[string]interface{}{
		"quizzes": quizzes,
		"count":   len(quizzes),
	}

	responseJSON, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}