package handlers

import (
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
)

type ClassRequest struct {
	ClassName string `json:"className"`
}

type SubjectRequest struct {
	ClassName   string `json:"className"`
	SubjectName string `json:"subjectName"`
}

type TopicRequest struct {
	ClassName   string `json:"className"`
	SubjectName string `json:"subjectName"`
	Topic       string `json:"topic"`
}

// Class APIs
func HandleClassInsert(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req ClassRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return CreateErrorResponse(400, "Invalid request body"), nil
	}

	if err := InsertClass(req.ClassName); err != nil {
		log.Printf("Failed to insert class: %v", err)
		return CreateErrorResponse(500, "Failed to insert class"), nil
	}

	return CreateSuccessResponse("Class inserted successfully"), nil
}

func HandleClassDelete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req ClassRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return CreateErrorResponse(400, "Invalid request body"), nil
	}

	if err := DeleteClass(req.ClassName); err != nil {
		log.Printf("Failed to delete class: %v", err)
		return CreateErrorResponse(500, "Failed to delete class"), nil
	}

	return CreateSuccessResponse("Class deleted successfully"), nil
}

func HandleClassFetch(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	classes, err := FetchClasses()
	if err != nil {
		log.Printf("Failed to fetch classes: %v", err)
		return CreateErrorResponse(500, "Failed to fetch classes"), nil
	}

	response, _ := json.Marshal(classes)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(response),
	}, nil
}

// Subject APIs
func HandleSubjectInsert(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req SubjectRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return CreateErrorResponse(400, "Invalid request body"), nil
	}

	if err := InsertSubject(req.ClassName, req.SubjectName); err != nil {
		log.Printf("Failed to insert subject: %v", err)
		return CreateErrorResponse(500, "Failed to insert subject"), nil
	}

	return CreateSuccessResponse("Subject inserted successfully"), nil
}

func HandleSubjectDelete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req SubjectRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return CreateErrorResponse(400, "Invalid request body"), nil
	}

	if err := DeleteSubject(req.ClassName, req.SubjectName); err != nil {
		log.Printf("Failed to delete subject: %v", err)
		return CreateErrorResponse(500, "Failed to delete subject"), nil
	}

	return CreateSuccessResponse("Subject deleted successfully"), nil
}

func HandleSubjectFetch(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	className := request.QueryStringParameters["className"]
	if className == "" {
		return CreateErrorResponse(400, "className parameter required"), nil
	}

	subjects, err := FetchSubjects(className)
	if err != nil {
		log.Printf("Failed to fetch subjects: %v", err)
		return CreateErrorResponse(500, "Failed to fetch subjects"), nil
	}

	response, _ := json.Marshal(subjects)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(response),
	}, nil
}

// Topic APIs
func HandleTopicInsert(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req TopicRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return CreateErrorResponse(400, "Invalid request body"), nil
	}

	if err := InsertTopic(req.ClassName, req.SubjectName, req.Topic); err != nil {
		log.Printf("Failed to insert topic: %v", err)
		return CreateErrorResponse(500, "Failed to insert topic"), nil
	}

	return CreateSuccessResponse("Topic inserted successfully"), nil
}

func HandleTopicDelete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req TopicRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return CreateErrorResponse(400, "Invalid request body"), nil
	}

	if err := DeleteTopic(req.ClassName, req.SubjectName, req.Topic); err != nil {
		log.Printf("Failed to delete topic: %v", err)
		return CreateErrorResponse(500, "Failed to delete topic"), nil
	}

	return CreateSuccessResponse("Topic deleted successfully"), nil
}

func HandleTopicFetch(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	className := request.QueryStringParameters["className"]
	subjectName := request.QueryStringParameters["subjectName"]
	
	if className == "" || subjectName == "" {
		return CreateErrorResponse(400, "className and subjectName parameters required"), nil
	}

	topics, err := FetchTopics(className, subjectName)
	if err != nil {
		log.Printf("Failed to fetch topics: %v", err)
		return CreateErrorResponse(500, "Failed to fetch topics"), nil
	}

	response, _ := json.Marshal(topics)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(response),
	}, nil
}