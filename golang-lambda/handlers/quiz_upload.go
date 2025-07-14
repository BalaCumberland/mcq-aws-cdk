package handlers

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/xuri/excelize/v2"
)

func HandleQuizUpload(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	fileContent, err := base64.StdEncoding.DecodeString(request.Body)
	if err != nil {
		return CreateErrorResponse(400, "Invalid file encoding"), nil
	}

	quizData, err := processExcel(fileContent, category, duration, quizName)
	if err != nil {
		return CreateErrorResponse(500, "Failed to process Excel file"), nil
	}

	err = SaveToPostgres(quizData)
	if err != nil {
		return CreateErrorResponse(500, "Failed to save to database"), nil
	}

	return CreateSuccessResponse("Quiz uploaded successfully"), nil
}

func processExcel(fileBytes []byte, category string, duration int, quizName string) (QuizData, error) {
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

	requiredHeaders := []string{"Question", "CorrectAnswer", "IncorrectAnswers", "Explanation"}
	for _, header := range requiredHeaders {
		if _, exists := headerMap[header]; !exists {
			return QuizData{}, fmt.Errorf("missing required column: %s", header)
		}
	}

	var questions []Question
	for _, row := range rows[1:] {
		questions = append(questions, Question{
			Question:         getCellValue(row, headerMap, "Question"),
			CorrectAnswer:    getCellValue(row, headerMap, "CorrectAnswer"),
			IncorrectAnswers: getCellValue(row, headerMap, "IncorrectAnswers"),
			Explanation:      getCellValue(row, headerMap, "Explanation"),
		})
	}

	return QuizData{QuizName: quizName, Duration: duration, Category: category, Questions: questions}, nil
}

func getCellValue(row []string, headerMap map[string]int, key string) string {
	index, exists := headerMap[key]
	if !exists || index >= len(row) {
		return ""
	}
	return row[index]
}