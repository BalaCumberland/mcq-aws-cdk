package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

type Answer struct {
	Qno     int      `json:"qno"`
	Options []string `json:"options"`
}

type SubmitRequest struct {
	Answers []Answer `json:"answers"`
}

type QuestionResult struct {
	Qno         int    `json:"qno"`
	Question    string `json:"question"`
	Status      string `json:"status"` // "correct", "wrong", "skipped"
	Explanation string `json:"explanation"`
}

type SubmitResponse struct {
	CorrectCount int              `json:"correctCount"`
	WrongCount   int              `json:"wrongCount"`
	SkippedCount int              `json:"skippedCount"`
	TotalCount   int              `json:"totalCount"`
	Percentage   float64          `json:"percentage"`
	Results      []QuestionResult `json:"results"`
}

func HandleQuizSubmit(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	quizName := request.QueryStringParameters["quizName"]
	if quizName == "" {
		return CreateErrorResponse(400, "Missing 'quizName' parameter"), nil
	}

	var submitReq SubmitRequest
	err := json.Unmarshal([]byte(request.Body), &submitReq)
	if err != nil {
		log.Printf("❌ Error parsing JSON: %v", err)
		return CreateErrorResponse(400, "Invalid JSON format"), nil
	}

	db, err := ConnectDB()
	if err != nil {
		log.Printf("❌ Database connection error: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}
	defer db.Close()

	// Get quiz questions
	var questionsJSON json.RawMessage
	err = db.QueryRow("SELECT questions FROM quiz_questions WHERE quiz_name = $1", quizName).Scan(&questionsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return CreateErrorResponse(404, "Quiz not found"), nil
		}
		log.Printf("❌ Database error: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	var questions []Question
	err = json.Unmarshal(questionsJSON, &questions)
	if err != nil {
		log.Printf("❌ Error parsing questions JSON: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
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

	for i, question := range questions {
		qno := i + 1
		answer, hasAnswer := answerMap[qno]
		
		var status string
		
		// Check if question was skipped
		if !hasAnswer || len(answer.Options) == 0 {
			status = "skipped"
			skippedCount++
		} else {
			// Parse, trim and lowercase correct answers (comma-separated)
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
			isCorrect := true
			if len(trimmedOptions) != len(correctAnswers) {
				isCorrect = false
			} else {
				// Check if all submitted options are in correct answers
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
		}
		
		results = append(results, QuestionResult{
			Qno:         qno,
			Question:    question.Question,
			Status:      status,
			Explanation: question.Explanation,
		})
	}

	var percentage float64
	if len(questions) > 0 {
		percentage = float64(correctCount) / float64(len(questions)) * 100
	}

	response := SubmitResponse{
		CorrectCount: correctCount,
		WrongCount:   wrongCount,
		SkippedCount: skippedCount,
		TotalCount:   len(questions),
		Percentage:   percentage,
		Results:      results,
	}

	responseJSON, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}