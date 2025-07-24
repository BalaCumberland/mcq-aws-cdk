package handlers

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func HandleQuizSubmitV3(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	uid, err := GetUserFromContext(request)
	if err != nil {
		return CreateErrorResponse(401, "Unauthorized"), nil
	}

	quizName := request.QueryStringParameters["quizName"]
	if quizName == "" {
		return CreateErrorResponse(400, "Missing 'quizName' parameter"), nil
	}

	var submitRequest SubmitRequest
	err = json.Unmarshal([]byte(request.Body), &submitRequest)
	if err != nil {
		return CreateErrorResponse(400, "Invalid JSON format"), nil
	}

	quiz, err := GetQuizFromDynamoDB(quizName)
	if err != nil || quiz == nil {
		return CreateErrorResponse(404, "Quiz not found"), nil
	}

	// Get next attempt number
	queryResult, _ := dynamoClient.Query(&dynamodb.QueryInput{
		TableName: aws.String("student_quiz_attempts_v3"),
		KeyConditionExpression: aws.String("uid = :uid AND quiz_name = :quiz_name"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":uid": {S: aws.String(uid)},
			":quiz_name": {S: aws.String(quizName)},
		},
		ScanIndexForward: aws.Bool(false),
		Limit: aws.Int64(1),
	})

	attemptNumber := 1
	if len(queryResult.Items) > 0 {
		if an := queryResult.Items[0]["attempt_number"]; an != nil && an.N != nil {
			if num, err := strconv.Atoi(*an.N); err == nil {
				attemptNumber = num + 1
			}
		}
	}

	// Calculate results
	var results []QuestionResult
	correctCount := 0
	wrongCount := 0
	skippedCount := 0

	for i, question := range quiz.Questions {
		qno := i + 1
		var studentAnswer []string
		
		for _, answer := range submitRequest.Answers {
			if answer.Qno == qno {
				studentAnswer = answer.Options
				break
			}
		}

		status := "skipped"
		if len(studentAnswer) > 0 {
			if len(studentAnswer) == 1 && studentAnswer[0] == question.CorrectAnswer {
				status = "correct"
				correctCount++
			} else {
				status = "wrong"
				wrongCount++
			}
		} else {
			skippedCount++
		}

		results = append(results, QuestionResult{
			Qno: qno,
			Question: question.Question,
			Status: status,
			StudentAnswer: studentAnswer,
			CorrectAnswer: []string{question.CorrectAnswer},
			Explanation: question.Explanation,
		})
	}

	totalCount := len(quiz.Questions)
	percentage := float64(correctCount) / float64(totalCount) * 100
	roundedPercentage := float64(int(percentage*10+0.5)) / 10

	// Save attempt
	attempt := AttemptItem{
		UID: uid,
		QuizName: quizName,
		Category: quiz.Category,
		CorrectCount: correctCount,
		WrongCount: wrongCount,
		SkippedCount: skippedCount,
		TotalCount: totalCount,
		Percentage: roundedPercentage,
		AttemptNumber: attemptNumber,
		AttemptedAt: time.Now().Format(time.RFC3339),
		Results: results,
	}

	err = SaveAttemptToDynamoDB(attempt)
	if err != nil {
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	response := map[string]interface{}{
		"correctCount": correctCount,
		"wrongCount": wrongCount,
		"skippedCount": skippedCount,
		"totalCount": totalCount,
		"percentage": roundedPercentage,
		"results": results,
	}

	responseJSON, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: GetCORSHeaders(),
		Body: string(responseJSON),
	}, nil
}