package handlers

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

func HandleQuizGetByNameV3(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	_, err := GetUserFromContext(request)
	if err != nil {
		return CreateErrorResponse(401, "Unauthorized"), nil
	}

	quizName := request.QueryStringParameters["quizName"]
	if quizName == "" {
		return CreateErrorResponse(400, "Missing 'quizName' parameter"), nil
	}

	quiz, err := GetQuizFromDynamoDB(quizName)
	if err != nil {
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	if quiz == nil {
		return CreateErrorResponse(404, "Quiz not found"), nil
	}

	responseJSON, _ := json.Marshal(quiz)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}