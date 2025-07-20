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
	Qno           int      `json:"qno"`
	Question      string   `json:"question"`
	Status        string   `json:"status"` // "correct", "wrong", "skipped"
	StudentAnswer []string `json:"studentAnswer"`
	CorrectAnswer []string `json:"correctAnswer"`
	Explanation   string   `json:"explanation"`
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
		
		// Map correct answer letters to actual answer text
		correctAnswersForResponse := []string{}
		correctAnswerLetters := strings.Split(question.CorrectAnswer, ",")
		for _, letter := range correctAnswerLetters {
			letter = strings.ToUpper(strings.TrimSpace(letter))
			switch letter {
			case "A":
				if len(question.AllAnswers) > 0 {
					correctAnswersForResponse = append(correctAnswersForResponse, question.AllAnswers[0])
				}
			case "B":
				if len(question.AllAnswers) > 1 {
					correctAnswersForResponse = append(correctAnswersForResponse, question.AllAnswers[1])
				}
			case "C":
				if len(question.AllAnswers) > 2 {
					correctAnswersForResponse = append(correctAnswersForResponse, question.AllAnswers[2])
				}
			case "D":
				if len(question.AllAnswers) > 3 {
					correctAnswersForResponse = append(correctAnswersForResponse, question.AllAnswers[3])
				}
			default:
				// If not A/B/C/D, save as is
				correctAnswersForResponse = append(correctAnswersForResponse, letter)
			}
		}
		
		// Map student answer letters to actual answer text
		var studentAnswerForResponse []string
		if hasAnswer && len(answer.Options) > 0 {
			for _, option := range answer.Options {
				switch strings.ToUpper(strings.TrimSpace(option)) {
				case "A":
					if len(question.AllAnswers) > 0 {
						studentAnswerForResponse = append(studentAnswerForResponse, question.AllAnswers[0])
					}
				case "B":
					if len(question.AllAnswers) > 1 {
						studentAnswerForResponse = append(studentAnswerForResponse, question.AllAnswers[1])
					}
				case "C":
					if len(question.AllAnswers) > 2 {
						studentAnswerForResponse = append(studentAnswerForResponse, question.AllAnswers[2])
					}
				case "D":
					if len(question.AllAnswers) > 3 {
						studentAnswerForResponse = append(studentAnswerForResponse, question.AllAnswers[3])
					}
				default:
					// If not A/B/C/D, save as is
					studentAnswerForResponse = append(studentAnswerForResponse, option)
				}
			}
		} else {
			studentAnswerForResponse = []string{} // Empty array for skipped
		}
		
		results = append(results, QuestionResult{
			Qno:           qno,
			Question:      question.Question,
			Status:        status,
			StudentAnswer: studentAnswerForResponse,
			CorrectAnswer: correctAnswersForResponse,
			Explanation:   question.Explanation,
		})
	}

	var percentage float64
	if len(questions) > 0 {
		percentage = float64(correctCount) / float64(len(questions)) * 100
	}

	// Get user email and quiz category for progress tracking
	email, err := GetUserFromContext(request)
	if err != nil {
		log.Printf("❌ Error getting user from context: %v", err)
		// Continue without saving progress if no user context
	} else {
		// Get quiz category
		var category string
		err = db.QueryRow("SELECT category FROM quiz_questions WHERE quiz_name = $1", quizName).Scan(&category)
		if err == nil {
			// Save quiz attempt
			err = SaveQuizAttempt(email, quizName, category, correctCount, wrongCount, skippedCount, len(questions), percentage, results)
			if err != nil {
				log.Printf("❌ Error saving quiz attempt: %v", err)
			}
		}
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