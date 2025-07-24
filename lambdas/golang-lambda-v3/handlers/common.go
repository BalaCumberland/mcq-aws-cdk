package handlers

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

type QuizData struct {
	QuizName  string     `json:"quizName"`
	Duration  interface{} `json:"duration"`
	Category  string     `json:"category"`
	Questions []Question `json:"questions"`
}

type Question struct {
	Explanation   string   `json:"explanation"`
	Question      string   `json:"question"`
	CorrectAnswer string   `json:"correctAnswer"`
	AllAnswers    []string `json:"allAnswers"`
}

type StudentUpdateRequest struct {
	Email        string  `json:"email,omitempty"`
	PhoneNumber  string  `json:"phoneNumber,omitempty"`
	Name         string  `json:"name,omitempty"`
	StudentClass string  `json:"studentClass,omitempty"`
	Amount       float64 `json:"amount,omitempty"`
	UpdatedBy    string  `json:"updatedBy,omitempty"`
	SubExpDate   string  `json:"subExpDate,omitempty"`
	PaymentTime  string  `json:"paymentTime,omitempty"`
	Role         string  `json:"role,omitempty"`
}

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
	Status        string   `json:"status"`
	StudentAnswer []string `json:"studentAnswer"`
	CorrectAnswer []string `json:"correctAnswer"`
	Explanation   string   `json:"explanation"`
}

type StudentRegisterRequest struct {
	Name         string `json:"name"`
	PhoneNumber  string `json:"phoneNumber"`
	StudentClass string `json:"studentClass"`
}

var VALID_CATEGORIES = []string{
	"CLS6-TELUGU", "CLS6-HINDI", "CLS6-ENGLISH", "CLS6-MATHS", "CLS6-SCIENCE", "CLS6-SOCIAL",
	"CLS7-TELUGU", "CLS7-HINDI", "CLS7-ENGLISH", "CLS7-MATHS", "CLS7-SCIENCE", "CLS7-SOCIAL",
	"CLS8-TELUGU", "CLS8-HINDI", "CLS8-ENGLISH", "CLS8-MATHS", "CLS8-SCIENCE", "CLS8-SOCIAL",
	"CLS9-TELUGU", "CLS9-HINDI", "CLS9-ENGLISH", "CLS9-MATHS", "CLS9-SCIENCE", "CLS9-SOCIAL",
	"CLS10-TELUGU", "CLS10-HINDI", "CLS10-ENGLISH", "CLS10-MATHS", "CLS10-SCIENCE", "CLS10-SOCIAL",
	"CLS10-BRIDGE", "CLS10-POLYTECHNIC", "CLS10-FORMULAS",
	"CLS11-MPC-PHYSICS", "CLS11-MPC-MATHS1A", "CLS11-MPC-MATHS1B", "CLS11-MPC-CHEMISTRY",
	"CLS11-MPC-EAMCET", "CLS11-MPC-JEEMAINS", "CLS11-MPC-JEEADV",
	"CLS12-MPC-PHYSICS", "CLS12-MPC-MATHS2A", "CLS12-MPC-MATHS2B", "CLS12-MPC-CHEMISTRY",
	"CLS12-MPC-EAMCET", "CLS12-MPC-JEEMAINS", "CLS12-MPC-JEEADV",
	"CLS11-BIPC-PHYSICS", "CLS11-BIPC-BOTANY", "CLS11-BIPC-ZOOLOGY", "CLS11-BIPC-CHEMISTRY",
	"CLS11-BIPC-EAPCET", "CLS11-BIPC-NEET",
	"CLS12-BIPC-PHYSICS", "CLS12-BIPC-BOTANY", "CLS12-BIPC-ZOOLOGY", "CLS12-BIPC-CHEMISTRY",
	"CLS12-BIPC-EAPCET", "CLS12-BIPC-NEET",
}

func GetUserFromContext(request events.APIGatewayProxyRequest) (string, error) {
	if request.RequestContext.Authorizer == nil {
		return "", fmt.Errorf("no authorizer context")
	}
	uid, ok := request.RequestContext.Authorizer["uid"].(string)
	if !ok || uid == "" {
		return "", fmt.Errorf("missing user uid from authorizer")
	}
	return uid, nil
}

func GetTargetUIDFromContext(request events.APIGatewayProxyRequest) (string, error) {
	if request.RequestContext.Authorizer == nil {
		return "", fmt.Errorf("no authorizer context")
	}
	targetUID, ok := request.RequestContext.Authorizer["targetUID"].(string)
	if !ok || targetUID == "" {
		return "", fmt.Errorf("missing target uid from authorizer")
	}
	return targetUID, nil
}

func GetEmailFromContext(request events.APIGatewayProxyRequest) string {
	if request.RequestContext.Authorizer == nil {
		return ""
	}
	email, _ := request.RequestContext.Authorizer["email"].(string)
	return email
}

func GetPhoneFromContext(request events.APIGatewayProxyRequest) string {
	if request.RequestContext.Authorizer == nil {
		return ""
	}
	phoneNumber, _ := request.RequestContext.Authorizer["phoneNumber"].(string)
	return phoneNumber
}

func CheckAdminRole(request events.APIGatewayProxyRequest) (string, error) {
	userUID, err := GetUserFromContext(request)
	if err != nil {
		return "", err
	}
	
	userStudent, _ := GetStudentFromDynamoDB(userUID)
	userRole := "student"
	if userStudent != nil && userStudent.Role != nil {
		if roleStr, ok := userStudent.Role.(string); ok {
			userRole = roleStr
		}
	}
	
	if userRole != "admin" && userRole != "super" {
		return "", fmt.Errorf("only 'admin' or 'super' role allowed")
	}
	
	return userRole, nil
}

func GetCORSHeaders() map[string]string {
	return map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "OPTIONS, POST, PUT, GET",
		"Access-Control-Allow-Headers": "Content-Type, Authorization",
	}
}

func CreateSuccessResponse(message string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       fmt.Sprintf(`{"message":"%s"}`, message),
	}
}

func CreateErrorResponse(statusCode int, errorMessage string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    GetCORSHeaders(),
		Body:       fmt.Sprintf(`{"error":"%s"}`, errorMessage),
	}
}