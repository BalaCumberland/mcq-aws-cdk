package handlers

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/xuri/excelize/v2"
)

func HandleQuizUploadV2(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract query parameters
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

	// Parse Content-Type and extract boundary
	contentType := request.Headers["Content-Type"]
	if contentType == "" {
		contentType = request.Headers["content-type"]
	}
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
		return CreateErrorResponse(400, "Expected multipart/form-data content-type"), nil
	}

	// Decode base64 body if needed
	var bodyBytes []byte
	if request.IsBase64Encoded {
		bodyBytes, err = base64.StdEncoding.DecodeString(request.Body)
		if err != nil {
			return CreateErrorResponse(400, "Failed to decode base64 body"), nil
		}
	} else {
		bodyBytes = []byte(request.Body)
	}

	// Parse multipart form data
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

	// Process Excel and save
	quizData, err := processExcelV2(fileContent, category, duration, quizName)
	if err != nil {
		log.Printf("❌ Excel processing error: %v", err)
		return CreateErrorResponse(500, fmt.Sprintf("Failed to process Excel file: %v", err)), nil
	}

	log.Printf("📌 Uploading quiz: %s", quizData.QuizName)

	err = SaveQuizToDynamoDB(quizData)
	if err != nil {
		log.Printf("❌ Error saving quiz: %v", err)
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

func processExcelV2(fileBytes []byte, category string, duration int, quizName string) (QuizData, error) {
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
		correctAnswer := getCellValueV2(row, headerMap, "CorrectAnswer")
		allAnswersStr := getCellValueV2(row, headerMap, "AllAnswers")
		
		// Parse all answers from Excel
		var allAnswers []string
		if allAnswersStr != "" {
			// Split all answers by ~~ delimiter (with or without spaces)
			allAnswers = strings.Split(allAnswersStr, "~~")
			for i := range allAnswers {
				allAnswers[i] = strings.TrimSpace(allAnswers[i])
			}
		}
		
		questions = append(questions, Question{
			Question:      getCellValueV2(row, headerMap, "Question"),
			CorrectAnswer: correctAnswer,
			AllAnswers:    allAnswers,
			Explanation:   getCellValueV2(row, headerMap, "Explanation"),
		})
	}

	return QuizData{
		QuizName:  quizName,
		Duration:  duration,
		Category:  category,
		Questions: questions,
	}, nil
}

func getCellValueV2(row []string, headerMap map[string]int, key string) string {
	index, exists := headerMap[key]
	if !exists || index >= len(row) {
		return ""
	}
	return row[index]
}