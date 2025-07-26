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
	case "/v2/upload/questions":
		return handlers.HandleQuizUploadV2(request)
	case "/v2/students/update":
		return handlers.HandleStudentUpdateV2(request)
	case "/v2/students/register":
		return handlers.HandleStudentRegisterV2(request)
	case "/v2/students/profile":
		return handlers.HandleStudentGetProfile(request)
	case "/v2/quiz/unattempted-quizzes":
		return handlers.HandleUnattemptedQuizzesV2(request)
	case "/v2/quiz":
		return handlers.HandleQuizGetByNameV2(request)
	case "/v2/quiz/submit":
		return handlers.HandleQuizSubmitV2(request)
	case "/v2/quiz/delete":
		return handlers.HandleQuizDeleteV2(request)
	case "/v2/quiz/list":
		return handlers.HandleQuizListV2(request)
	case "/v2/students/progress":
		return handlers.HandleStudentProgressV2(request)
	case "/v2/quiz/result":
		return handlers.HandleQuizResultV2(request)
	case "/v2/students/upgrade-class":
		return handlers.HandleStudentClassUpgradeV2(request)
	case "/v2/class/insert":
		return handlers.HandleClassInsert(request)
	case "/v2/class/delete":
		return handlers.HandleClassDelete(request)
	case "/v2/class/fetch":
		return handlers.HandleClassFetch(request)
	case "/v2/subject/insert":
		return handlers.HandleSubjectInsert(request)
	case "/v2/subject/delete":
		return handlers.HandleSubjectDelete(request)
	case "/v2/subject/fetch":
		return handlers.HandleSubjectFetch(request)
	case "/v2/topic/insert":
		return handlers.HandleTopicInsert(request)
	case "/v2/topic/delete":
		return handlers.HandleTopicDelete(request)
	case "/v2/topic/fetch":
		return handlers.HandleTopicFetch(request)
	case "/v2/students/lookup":
		return handlers.HandleStudentLookup(request)
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
