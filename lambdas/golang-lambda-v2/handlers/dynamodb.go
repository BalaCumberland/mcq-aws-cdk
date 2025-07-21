package handlers

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var dynamoClient *dynamodb.DynamoDB

func init() {
	sess := session.Must(session.NewSession())
	dynamoClient = dynamodb.New(sess)
}

// Quiz item structure
type QuizItem struct {
	QuizName  string      `json:"quiz_name" dynamodbav:"quiz_name"`
	Duration  interface{} `json:"duration" dynamodbav:"duration"`
	Category  string      `json:"category" dynamodbav:"category"`
	Questions []Question  `json:"questions" dynamodbav:"questions"`
}

// Student item structure
type StudentItem struct {
	Email        string      `json:"email" dynamodbav:"email"`
	Name         string      `json:"name" dynamodbav:"name"`
	StudentClass string      `json:"student_class" dynamodbav:"student_class"`
	PhoneNumber  string      `json:"phone_number" dynamodbav:"phone_number"`
	SubExpDate   interface{} `json:"sub_exp_date,omitempty" dynamodbav:"sub_exp_date,omitempty"`
	UpdatedBy    interface{} `json:"updated_by,omitempty" dynamodbav:"updated_by,omitempty"`
	Amount       interface{} `json:"amount,omitempty" dynamodbav:"amount,omitempty"`
	PaymentTime  interface{} `json:"payment_time,omitempty" dynamodbav:"payment_time,omitempty"`
	Role         interface{} `json:"role,omitempty" dynamodbav:"role,omitempty"`
}

// Quiz attempt item structure
type AttemptItem struct {
	Email        string           `json:"email" dynamodbav:"email"`
	QuizName     string           `json:"quiz_name" dynamodbav:"quiz_name"`
	Category     string           `json:"category" dynamodbav:"category"`
	CorrectCount int              `json:"correct_count" dynamodbav:"correct_count"`
	WrongCount   int              `json:"wrong_count" dynamodbav:"wrong_count"`
	SkippedCount int              `json:"skipped_count" dynamodbav:"skipped_count"`
	TotalCount   int              `json:"total_count" dynamodbav:"total_count"`
	Percentage   interface{}      `json:"percentage" dynamodbav:"percentage"`
	AttemptNumber int             `json:"attempt_number" dynamodbav:"attempt_number"`
	AttemptedAt  string           `json:"attempted_at" dynamodbav:"attempted_at"`
	Results      []QuestionResult `json:"results" dynamodbav:"results"`
}

// Save quiz to DynamoDB
func SaveQuizToDynamoDB(quiz QuizData) error {
	item := QuizItem{
		QuizName:  quiz.QuizName,
		Duration:  quiz.Duration,
		Category:  quiz.Category,
		Questions: quiz.Questions,
	}

	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = dynamoClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("quiz_questions"),
		Item:      av,
	})

	return err
}

// Get quiz from DynamoDB
func GetQuizFromDynamoDB(quizName string) (*QuizItem, error) {
	result, err := dynamoClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("quiz_questions"),
		Key: map[string]*dynamodb.AttributeValue{
			"quiz_name": {S: aws.String(quizName)},
		},
	})

	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, nil
	}

	var quiz QuizItem
	err = dynamodbattribute.UnmarshalMap(result.Item, &quiz)
	return &quiz, err
}

// Save student to DynamoDB
func SaveStudentToDynamoDB(student StudentItem) error {
	av, err := dynamodbattribute.MarshalMap(student)
	if err != nil {
		return err
	}

	_, err = dynamoClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("students"),
		Item:      av,
	})

	return err
}

// Get student from DynamoDB
func GetStudentFromDynamoDB(email string) (*StudentItem, error) {
	result, err := dynamoClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("students"),
		Key: map[string]*dynamodb.AttributeValue{
			"email": {S: aws.String(email)},
		},
	})

	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, nil
	}

	var student StudentItem
	err = dynamodbattribute.UnmarshalMap(result.Item, &student)
	return &student, err
}

// Save quiz attempt to DynamoDB
func SaveAttemptToDynamoDB(attempt AttemptItem) error {
	av, err := dynamodbattribute.MarshalMap(attempt)
	if err != nil {
		return err
	}

	_, err = dynamoClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("student_quiz_attempts"),
		Item:      av,
	})

	return err
}