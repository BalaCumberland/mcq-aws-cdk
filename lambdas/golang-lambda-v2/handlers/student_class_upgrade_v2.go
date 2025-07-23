package handlers

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

type ClassUpgradeRequest struct {
	NewClass string `json:"newClass"`
}

func HandleStudentClassUpgradeV2(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get authenticated user email
	userEmail, err := GetUserFromContext(request)
	if err != nil {
		log.Printf("❌ Failed to get user from context: %v", err)
		return CreateErrorResponse(401, "Unauthorized"), nil
	}

	var upgradeRequest ClassUpgradeRequest
	err = json.Unmarshal([]byte(request.Body), &upgradeRequest)
	if err != nil {
		log.Printf("❌ Error parsing JSON: %v", err)
		return CreateErrorResponse(400, "Invalid JSON format"), nil
	}

	if upgradeRequest.NewClass == "" {
		return CreateErrorResponse(400, "Missing 'newClass' parameter"), nil
	}

	email := strings.ToLower(userEmail)
	log.Printf("📌 Upgrading class for student: %s to %s", email, upgradeRequest.NewClass)

	// Get existing student
	student, err := GetStudentFromDynamoDB(email)
	if err != nil {
		log.Printf("❌ Error fetching student: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	if student == nil {
		return CreateErrorResponse(404, "Student not found"), nil
	}

	currentClass := student.StudentClass
	newClass := upgradeRequest.NewClass

	// Validate upgrade path
	if !isValidUpgrade(currentClass, newClass) {
		return CreateErrorResponse(400, "Invalid class upgrade path"), nil
	}

	// Update student class
	student.StudentClass = newClass

	// Save updated student
	err = SaveStudentToDynamoDB(*student)
	if err != nil {
		log.Printf("❌ Error updating student: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	response := map[string]interface{}{
		"message":     "Class upgraded successfully",
		"email":       email,
		"oldClass":    currentClass,
		"newClass":    newClass,
	}

	responseJSON, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}

func getUpgradableClasses(currentClass string) []string {
	validUpgrades := map[string][]string{
		"CLS6":       {"CLS7"},
		"CLS7":       {"CLS8"},
		"CLS8":       {"CLS9"},
		"CLS9":       {"CLS10"},
		"CLS10":      {"CLS11-MPC", "CLS11-BIPC"},
		"CLS11-MPC":  {"CLS12-MPC"},
		"CLS11-BIPC": {"CLS12-BIPC"},
	}

	allowedUpgrades, exists := validUpgrades[currentClass]
	if !exists {
		return []string{}
	}

	return allowedUpgrades
}

func isValidUpgrade(currentClass, newClass string) bool {
	upgradableClasses := getUpgradableClasses(currentClass)
	for _, allowed := range upgradableClasses {
		if newClass == allowed {
			return true
		}
	}
	return false
}