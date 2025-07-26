package handlers

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func HandleQuizSubmitV2(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	var submitReq SubmitRequest
	err := json.Unmarshal([]byte(request.Body), &submitReq)
	if err != nil {
		log.Printf("❌ Error parsing JSON: %v", err)
		return CreateErrorResponse(400, "Invalid JSON format"), nil
	}

	log.Printf("📌 Processing quiz submission: %s for %s-%s-%s", quizName, className, subjectName, topic)

	// Get quiz data
	quiz, err := GetQuizFromDynamoDB(quizName, className, subjectName, topic)
	if err != nil || quiz == nil {
		log.Printf("❌ Quiz not found: %v", err)
		return CreateErrorResponse(404, "Quiz not found"), nil
	}

	// Create answer map for quick lookup
	answerMap := make(map[int]Answer)
	for _, answer := range submitReq.Answers {
		answerMap[answer.Qno] = answer
	}

	// Process all questions in order
	var results []QuestionResult
	correctCount := 0
	wrongCount := 0
	skippedCount := 0

	for i, question := range quiz.Questions {
		qno := i + 1
		answer, hasAnswer := answerMap[qno]
		
		var status string
		var studentAnswer []string
		
		// Check if question was skipped
		if !hasAnswer || len(answer.Options) == 0 {
			status = "skipped"
			skippedCount++
			studentAnswer = []string{}
		} else {
			// Parse correct answers (comma-separated)
			correctAnswers := strings.Split(question.CorrectAnswer, ",")
			for j := range correctAnswers {
				correctAnswers[j] = strings.ToLower(strings.TrimSpace(correctAnswers[j]))
			}
			
			// Trim and lowercase submitted options
			trimmedOptions := make([]string, len(answer.Options))
			for j, option := range answer.Options {
				trimmedOptions[j] = strings.ToLower(strings.TrimSpace(option))
			}
			
			// Check if submitted options match correct answers
			isCorrect := len(trimmedOptions) == len(correctAnswers)
			if isCorrect {
				for _, option := range trimmedOptions {
					found := false
					for _, correct := range correctAnswers {
						if option == correct {
							found = true
							break
						}
					}
					if !found {
						isCorrect = false
						break
					}
				}
			}
			
			if isCorrect {
				status = "correct"
				correctCount++
			} else {
				status = "wrong"
				wrongCount++
			}
			
			// Map student answer letters to actual text
			for _, option := range answer.Options {
				switch strings.ToUpper(strings.TrimSpace(option)) {
				case "A":
					if len(question.AllAnswers) > 0 {
						studentAnswer = append(studentAnswer, question.AllAnswers[0])
					}
				case "B":
					if len(question.AllAnswers) > 1 {
						studentAnswer = append(studentAnswer, question.AllAnswers[1])
					}
				case "C":
					if len(question.AllAnswers) > 2 {
						studentAnswer = append(studentAnswer, question.AllAnswers[2])
					}
				case "D":
					if len(question.AllAnswers) > 3 {
						studentAnswer = append(studentAnswer, question.AllAnswers[3])
					}
				}
			}
		}
		
		// Map correct answer letters to actual text
		correctAnswerText := []string{}
		correctAnswerLetters := strings.Split(question.CorrectAnswer, ",")
		for _, letter := range correctAnswerLetters {
			letter = strings.ToUpper(strings.TrimSpace(letter))
			switch letter {
			case "A":
				if len(question.AllAnswers) > 0 {
					correctAnswerText = append(correctAnswerText, question.AllAnswers[0])
				}
			case "B":
				if len(question.AllAnswers) > 1 {
					correctAnswerText = append(correctAnswerText, question.AllAnswers[1])
				}
			case "C":
				if len(question.AllAnswers) > 2 {
					correctAnswerText = append(correctAnswerText, question.AllAnswers[2])
				}
			case "D":
				if len(question.AllAnswers) > 3 {
					correctAnswerText = append(correctAnswerText, question.AllAnswers[3])
				}
			}
		}
		
		results = append(results, QuestionResult{
			Qno:           qno,
			Question:      question.Question,
			Status:        status,
			StudentAnswer: studentAnswer,
			CorrectAnswer: correctAnswerText,
			Explanation:   question.Explanation,
		})
	}

	totalCount := len(quiz.Questions)
	percentage := float64(correctCount) / float64(totalCount) * 100

	// Get user UID from context
	uid, err := GetUserUIDFromContext(request)
	if err != nil {
		return CreateErrorResponse(401, "Unauthorized"), nil
	}

	// Get existing attempt to increment attempt number
	existingResult, _ := dynamoClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("student_quiz_attempts_v2"),
		Key: map[string]*dynamodb.AttributeValue{
			"uid":       {S: aws.String(uid)},
			"quiz_name": {S: aws.String(quizName)},
		},
	})

	attemptNumber := 1
	if existingResult.Item != nil {
		// Delete existing record first
		_, _ = dynamoClient.DeleteItem(&dynamodb.DeleteItemInput{
			TableName: aws.String("student_quiz_attempts_v2"),
			Key: map[string]*dynamodb.AttributeValue{
				"uid":       {S: aws.String(uid)},
				"quiz_name": {S: aws.String(quizName)},
			},
		})
		
		// Increment attempt number
		if attemptNum := existingResult.Item["attempt_number"]; attemptNum != nil && attemptNum.N != nil {
			if num, err := strconv.Atoi(*attemptNum.N); err == nil {
				attemptNumber = num + 1
			}
		}
	}

	// Save new attempt with incremented count
	attempt := AttemptItem{
		UID:           uid,
		QuizName:      quizName,
		ClassName:     quiz.ClassName,
		Category:      quiz.SubjectName,
		CorrectCount:  correctCount,
		WrongCount:    wrongCount,
		SkippedCount:  skippedCount,
		TotalCount:    totalCount,
		Percentage:    percentage,
		AttemptNumber: attemptNumber,
		AttemptedAt:   time.Now().Format("2006-01-02T15:04:05Z"),
		Results:       results,
	}

	err = SaveAttemptToDynamoDB(attempt)
	if err != nil {
		log.Printf("❌ Error saving attempt: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	response := map[string]interface{}{
		"correctCount": correctCount,
		"wrongCount":   wrongCount,
		"skippedCount": skippedCount,
		"totalCount":   totalCount,
		"percentage":   percentage,
		"results":      results,
	}

	responseJSON, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}