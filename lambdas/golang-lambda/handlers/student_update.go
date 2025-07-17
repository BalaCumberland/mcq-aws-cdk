package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func HandleStudentUpdate(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userEmail, err := GetUserFromContext(request)
	if err != nil {
		log.Printf("❌ Failed to get user from context: %v", err)
		return CreateErrorResponse(401, "Unauthorized"), nil
	}
	log.Printf("🔐 Authenticated user: %s", userEmail)

	var studentUpdate StudentUpdateRequest
	err = json.Unmarshal([]byte(request.Body), &studentUpdate)
	if err != nil {
		log.Println("❌ Error parsing JSON:", err)
		return CreateErrorResponse(400, "Invalid JSON format"), nil
	}

	if studentUpdate.Email == "" {
		return CreateErrorResponse(400, "Missing 'email' parameter"), nil
	}

	db, err := ConnectDB()
	if err != nil {
		log.Println("❌ Database connection error:", err)
		return CreateErrorResponse(500, "Database connection failed"), nil
	}
	defer db.Close()

	userRole, err := GetUserRole(db, userEmail)
	if err != nil {
		log.Printf("❌ Failed to get user role: %v", err)
		return CreateErrorResponse(500, "Failed to verify user permissions"), nil
	}

	isSubscriptionUpdate := studentUpdate.Amount != nil
	if isSubscriptionUpdate && userRole != "super" {
		return CreateErrorResponse(403, "Only 'super' role can update subscription"), nil
	}
	if !isSubscriptionUpdate && userRole != "admin" && userRole != "super" {
		return CreateErrorResponse(403, "Only 'admin' or 'super' role can update student fields"), nil
	}

	rowsAffected, err := updateStudent(db, studentUpdate)
	if err != nil {
		log.Println("❌ Error updating student:", err)
		return CreateErrorResponse(500, "Internal server error"), nil
	}

	if rowsAffected == 0 {
		return CreateErrorResponse(404, "No student found with the provided email"), nil
	}

	return CreateSuccessResponse("Student updated successfully"), nil
}

func updateStudent(db *sql.DB, student StudentUpdateRequest) (int64, error) {
	normalizedEmail := strings.ToLower(student.Email)
	log.Printf("🔍 Updating student: Email = %s", normalizedEmail)

	var existingSubExpDate sql.NullString
	err := db.QueryRow("SELECT sub_exp_date FROM students WHERE LOWER(email) = $1", normalizedEmail).Scan(&existingSubExpDate)
	if err != nil {
		log.Printf("❌ Failed to fetch existing sub_exp_date for email %s: %v", normalizedEmail, err)
		return 0, fmt.Errorf("failed to fetch existing sub_exp_date: %w", err)
	}

	log.Printf("📅 Existing sub_exp_date: %v", existingSubExpDate.String)

	today := time.Now().Format("2006-01-02")

	tx, err := db.Begin()
	if err != nil {
		log.Printf("❌ Failed to begin transaction for email %s: %v", normalizedEmail, err)
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := "UPDATE students SET "
	params := []interface{}{normalizedEmail}
	paramIndex := 2
	updateFields := []string{}

	if student.Name != nil && *student.Name != "" {
		log.Printf("📝 Updating name: %s", *student.Name)
		updateFields = append(updateFields, fmt.Sprintf("name = $%d", paramIndex))
		params = append(params, *student.Name)
		paramIndex++
	}

	if student.PhoneNumber != nil && *student.PhoneNumber != "" {
		log.Printf("📞 Updating phone number: %s", *student.PhoneNumber)
		updateFields = append(updateFields, fmt.Sprintf("phone_number = $%d", paramIndex))
		params = append(params, *student.PhoneNumber)
		paramIndex++
	}

	if student.StudentClass != nil && *student.StudentClass != "" {
		log.Printf("🏫 Updating student class: %s", *student.StudentClass)
		updateFields = append(updateFields, fmt.Sprintf("student_class = $%d", paramIndex))
		params = append(params, *student.StudentClass)
		paramIndex++
	}

	if student.Amount != nil {
		log.Printf("💰 Updating amount: %f", *student.Amount)
		updateFields = append(updateFields, fmt.Sprintf("amount = $%d", paramIndex))
		params = append(params, *student.Amount)
		paramIndex++

		if *student.Amount > 0 {
			log.Printf("⏳ Updating payment_time to NOW() since amount > 0")
			updateFields = append(updateFields, "payment_time = NOW()")

			var newSubExpDate string
			if existingSubExpDate.Valid && existingSubExpDate.String >= today {
				log.Printf("📅 Extending sub_exp_date by 1 year from %s", existingSubExpDate.String)
				newSubExpDate = fmt.Sprintf("DATE '%s' + INTERVAL '1 year'", existingSubExpDate.String)
			} else {
				log.Printf("📅 Setting new sub_exp_date as today + 1 year")
				newSubExpDate = fmt.Sprintf("DATE '%s' + INTERVAL '1 year'", today)
			}

			updateFields = append(updateFields, fmt.Sprintf("sub_exp_date = %s", newSubExpDate))

			if student.UpdatedBy != nil && *student.UpdatedBy != "" {
				log.Printf("👤 Updated by: %s", *student.UpdatedBy)
				updateFields = append(updateFields, fmt.Sprintf("updated_by = $%d", paramIndex))
				params = append(params, *student.UpdatedBy)
				paramIndex++
			}
		} else {
			log.Printf("💰 Amount is 0, skipping sub_exp_date & payment_time update")
		}
	}

	if len(updateFields) == 0 {
		log.Printf("⚠️ No valid fields to update for email: %s", normalizedEmail)
		return 0, fmt.Errorf("no valid fields to update")
	}

	query += fmt.Sprintf("%s WHERE LOWER(email) = $1", strings.Join(updateFields, ", "))

	log.Printf("📡 Executing query: %s", query)

	result, err := tx.Exec(query, params...)
	if err != nil {
		log.Printf("❌ Failed to execute update for email %s: %v", normalizedEmail, err)
		return 0, fmt.Errorf("failed to execute update: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("❌ Failed to commit transaction for email %s: %v", normalizedEmail, err)
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("❌ Failed to fetch affected rows for email %s: %v", normalizedEmail, err)
		return 0, fmt.Errorf("failed to fetch affected rows: %w", err)
	}

	log.Printf("✅ Successfully updated %d row(s) for email %s", rowsAffected, normalizedEmail)
	return rowsAffected, nil
}