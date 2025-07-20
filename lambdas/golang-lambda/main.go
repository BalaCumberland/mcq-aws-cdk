package main

import (
	"fmt"
	"log"

	"go-upload-excel/handlers"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func lambdaHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("🚀 Lambda function started")
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
	case "/upload/questions":
		return handlers.HandleQuizUpload(request)
	case "/students/update":
		return handlers.HandleStudentUpdate(request)
	case "/students/register":
		return handlers.HandleStudentRegister(request)
	case "/students/get-by-email":
		return handlers.HandleStudentGetByEmail(request)
	case "/quiz/unattempted-quizzes":
		return handlers.HandleUnattemptedQuizzes(request)
	case "/quiz/get-by-name":
		return handlers.HandleQuizGetByName(request)
	case "/quiz/submit":
		return handlers.HandleQuizSubmit(request)
	case "/quiz/delete":
		return handlers.HandleQuizDelete(request)
	case "/students/progress":
		return handlers.HandleStudentProgress(request)
	case "/quiz/result":
		return handlers.HandleQuizResult(request)
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
	log.Printf("🚀 Starting Lambda function...")
	lambda.Start(lambdaHandler)
}
