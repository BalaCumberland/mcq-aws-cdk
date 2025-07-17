package handlers

import (
	"log"
	"net/url"

	"github.com/aws/aws-lambda-go/events"
)

func HandleQuizDelete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	quizName := request.QueryStringParameters["quizName"]
	if quizName == "" {
		return CreateErrorResponse(400, "Missing 'quizName' parameter"), nil
	}

	// URL decode the quiz name
	decodedQuizName, err := url.QueryUnescape(quizName)
	if err != nil {
		log.Printf("❌ Failed to decode quiz name: %v", err)
		decodedQuizName = quizName // fallback to original
	}

	// Get user email from JWT token
	email, err := GetUserFromContext(request)
	if err != nil {
		log.Printf("❌ Failed to get user from context: %v", err)
		return CreateErrorResponse(401, "Unauthorized"), nil
	}

	db, err := ConnectDB()
	if err != nil {
		log.Printf("❌ Database connection error: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}
	defer db.Close()

	// Check user role
	userRole, err := GetUserRole(db, email)
	if err != nil {
		log.Printf("❌ Failed to get user role: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	if userRole != "admin" && userRole != "super" {
		return CreateErrorResponse(403, "Access denied. Admin or Super role required"), nil
	}

	// Delete quiz
	result, err := db.Exec("DELETE FROM quiz_questions WHERE quiz_name = $1", decodedQuizName)
	if err != nil {
		log.Printf("❌ Database error: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return CreateErrorResponse(404, "Quiz not found"), nil
	}

	return CreateSuccessResponse("Quiz deleted successfully"), nil
}