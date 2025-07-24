package main

import (
	"fmt"
	"log"

	"go-upload-excel-v3/handlers"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func lambdaHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("🚀 Lambda V3 function started")
	log.Printf("📌 Received request: Path = %s, Method = %s", request.Path, request.HTTPMethod)
	log.Printf("Path: %s, Resource: %s, Stage: %s", request.Path, request.Resource, request.RequestContext.Stage)

	if request.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    handlers.GetCORSHeaders(),
			Body:       `{"message":"CORS preflight response"}`,
		}, nil
	}

	switch request.Path {
	case "/v3/students/register":
		return handlers.HandleStudentRegisterV3(request)
	case "/v3/students/get":
		return handlers.HandleStudentGetV3(request)
	case "/v3/students/update":
		return handlers.HandleStudentUpdateV3(request)
	case "/v3/students/progress":
		return handlers.HandleStudentProgressV3(request)
	case "/v3/students/upgrade-class":
		return handlers.HandleStudentClassUpgradeV3(request)
	case "/v3/quiz/get-by-name":
		return handlers.HandleQuizGetByNameV3(request)
	case "/v3/quiz/submit":
		return handlers.HandleQuizSubmitV3(request)
	case "/v3/quiz/unattempted-quizzes":
		return handlers.HandleUnattemptedQuizzesV3(request)
	case "/v3/upload/questions":
		return handlers.HandleQuizUploadV3(request)
	case "/v3/quiz/delete":
		return handlers.HandleQuizDeleteV3(request)
	case "/v3/quiz/result":
		return handlers.HandleQuizResultV3(request)
	default:
		log.Printf("❌ Invalid API Path: %s", request.Path)
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Headers:    handlers.GetCORSHeaders(),
			Body:       fmt.Sprintf(`{"error":"Invalid API endpoint", "receivedPath": "%s"}`, request.Path),
		}, nil
	}
}

func main() {
	log.Printf("🚀 Starting Lambda V3 function...")
	lambda.Start(lambdaHandler)
}