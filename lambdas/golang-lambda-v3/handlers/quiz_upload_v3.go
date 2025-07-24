package handlers

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/xuri/excelize/v2"
)

func HandleQuizUploadV3(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	_, err := CheckAdminRole(request)
	if err != nil {
		return CreateErrorResponse(403, err.Error()), nil
	}

	queryParams := request.QueryStringParameters
	category := queryParams["category"]
	durationStr := queryParams["duration"]
	quizName := queryParams["quizName"]

	if category == "" || durationStr == "" || quizName == "" {
		return CreateErrorResponse(400, "Missing required query parameters"), nil
	}

	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		return CreateErrorResponse(400, "Invalid duration format"), nil
	}

	contentType := request.Headers["Content-Type"]
	if contentType == "" {
		contentType = request.Headers["content-type"]
	}
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
		return CreateErrorResponse(400, "Expected multipart/form-data content-type"), nil
	}

	var bodyBytes []byte
	if request.IsBase64Encoded {
		bodyBytes, err = base64.StdEncoding.DecodeString(request.Body)
		if err != nil {
			return CreateErrorResponse(400, "Failed to decode base64 body"), nil
		}
	} else {
		bodyBytes = []byte(request.Body)
	}

	reader := multipart.NewReader(bytes.NewReader(bodyBytes), params["boundary"])

	var fileContent []byte
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return CreateErrorResponse(400, "Failed to parse multipart file"), nil
		}
		if part.FormName() == "file" {
			fileContent, err = io.ReadAll(part)
			if err != nil {
				return CreateErrorResponse(400, "Failed to read file content"), nil
			}
			break
		}
	}

	if len(fileContent) == 0 {
		return CreateErrorResponse(400, "File content is empty or missing"), nil
	}

	quizData, err := processExcelV3(fileContent, category, duration, quizName)
	if err != nil {
		return CreateErrorResponse(500, fmt.Sprintf("Failed to process Excel file: %v", err)), nil
	}

	err = SaveQuizToDynamoDB(quizData)
	if err != nil {
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	responseJSON := fmt.Sprintf(`{"message":"%s","quizName":"%s","category":"%s","duration":%v,"questionCount":%d}`, 
		"Quiz uploaded successfully", quizData.QuizName, quizData.Category, quizData.Duration, len(quizData.Questions))
	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Headers:    GetCORSHeaders(),
		Body:       responseJSON,
	}, nil
}

func processExcelV3(fileBytes []byte, category string, duration int, quizName string) (QuizData, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return QuizData{}, err
	}

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return QuizData{}, err
	}

	if len(rows) < 2 {
		return QuizData{}, errors.New("insufficient data in the file")
	}

	headerMap := make(map[string]int)
	for i, header := range rows[0] {
		headerMap[header] = i
	}

	requiredHeaders := []string{"Question", "CorrectAnswer", "AllAnswers", "Explanation"}
	for _, header := range requiredHeaders {
		if _, exists := headerMap[header]; !exists {
			return QuizData{}, fmt.Errorf("missing required column: %s", header)
		}
	}

	var questions []Question
	for _, row := range rows[1:] {
		correctAnswer := getCellValueV3(row, headerMap, "CorrectAnswer")
		allAnswersStr := getCellValueV3(row, headerMap, "AllAnswers")
		
		var allAnswers []string
		if allAnswersStr != "" {
			allAnswers = strings.Split(allAnswersStr, "~~")
			for i := range allAnswers {
				allAnswers[i] = strings.TrimSpace(allAnswers[i])
			}
		}
		
		questions = append(questions, Question{
			Question:      getCellValueV3(row, headerMap, "Question"),
			CorrectAnswer: correctAnswer,
			AllAnswers:    allAnswers,
			Explanation:   getCellValueV3(row, headerMap, "Explanation"),
		})
	}

	return QuizData{
		QuizName:  quizName,
		Duration:  duration,
		Category:  category,
		Questions: questions,
	}, nil
}

func getCellValueV3(row []string, headerMap map[string]int, key string) string {
	index, exists := headerMap[key]
	if !exists || index >= len(row) {
		return ""
	}
	return row[index]
}

func SaveQuizToDynamoDB(quiz QuizData) error {
	item := QuizItem{
		QuizName:  quiz.QuizName,
		Duration:  quiz.Duration,
		Category:  quiz.Category,
		Questions: quiz.Questions,
	}

	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = dynamoClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("quiz_questions"),
		Item:      av,
	})

	return err
}