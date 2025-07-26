package handlers

import (
	"encoding/json"
	"log"

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

	log.Printf("📌 Deleting quiz: %s (%s-%s-%s)", quizName, className, subjectName, topic)

	// Check if quiz exists with all filters
	getResult, err := dynamoClient.Scan(&dynamodb.ScanInput{
		TableName:        aws.String("quiz_questions"),
		FilterExpression: aws.String("quiz_name = :quizName AND class_name = :className AND subject_name = :subjectName AND topic = :topic"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":quizName":    {S: aws.String(quizName)},
			":className":   {S: aws.String(className)},
			":subjectName": {S: aws.String(subjectName)},
			":topic":       {S: aws.String(topic)},
		},
	})

	if err != nil {
		log.Printf("❌ Error checking quiz: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	if len(getResult.Items) == 0 {
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
		log.Printf("❌ Error deleting quiz: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	// Delete all attempt records for this specific quiz (matching all filters)
	scanResult, scanErr := dynamoClient.Scan(&dynamodb.ScanInput{
		TableName: aws.String("student_quiz_attempts_v2"),
		FilterExpression: aws.String("quiz_name = :quiz_name AND class_name = :className AND category = :subjectName"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":quiz_name":    {S: aws.String(quizName)},
			":className":    {S: aws.String(className)},
			":subjectName":  {S: aws.String(subjectName)},
		},
	})

	if scanErr == nil {
		for _, item := range scanResult.Items {
			if email := item["email"]; email != nil && email.S != nil {
				_, _ = dynamoClient.DeleteItem(&dynamodb.DeleteItemInput{
					TableName: aws.String("student_quiz_attempts_v2"),
					Key: map[string]*dynamodb.AttributeValue{
						"email": {S: email.S},
						"quiz_name": {S: aws.String(quizName)},
					},
				})
				log.Printf("🗑️ Deleted attempt record for %s", *email.S)
			}
		}
		log.Printf("🗑️ Deleted %d attempt records for quiz %s", len(scanResult.Items), quizName)
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