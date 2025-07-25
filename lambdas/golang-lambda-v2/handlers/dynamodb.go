package handlers

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var dynamoClient *dynamodb.DynamoDB

func init() {
	// Optimized HTTP client with connection pooling
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  false,
		},
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region:     aws.String("us-east-1"),
		MaxRetries: aws.Int(3),
		HTTPClient: httpClient,
	}))

	dynamoClient = dynamodb.New(sess)
}

// Quiz item structure
type QuizItem struct {
	QuizName    string      `json:"quiz_name" dynamodbav:"quiz_name"`
	Duration    interface{} `json:"duration" dynamodbav:"duration"`
	ClassName   string      `json:"class_name" dynamodbav:"class_name"`
	SubjectName string      `json:"subject_name" dynamodbav:"subject_name"`
	Topic       string      `json:"topic" dynamodbav:"topic"`
	Questions   []Question  `json:"questions" dynamodbav:"questions"`
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
	Email         string           `json:"email" dynamodbav:"email"`
	QuizName      string           `json:"quiz_name" dynamodbav:"quiz_name"`
	ClassName     string           `json:"class_name" dynamodbav:"class_name"`
	Category      string           `json:"category" dynamodbav:"category"`
	CorrectCount  int              `json:"correct_count" dynamodbav:"correct_count"`
	WrongCount    int              `json:"wrong_count" dynamodbav:"wrong_count"`
	SkippedCount  int              `json:"skipped_count" dynamodbav:"skipped_count"`
	TotalCount    int              `json:"total_count" dynamodbav:"total_count"`
	Percentage    interface{}      `json:"percentage" dynamodbav:"percentage"`
	AttemptNumber int              `json:"attempt_number" dynamodbav:"attempt_number"`
	AttemptedAt   string           `json:"attempted_at" dynamodbav:"attempted_at"`
	Results       []QuestionResult `json:"results" dynamodbav:"results"`
}

// Save quiz to DynamoDB
func SaveQuizToDynamoDB(quiz QuizData) error {
	item := QuizItem{
		QuizName:    quiz.QuizName,
		Duration:    quiz.Duration,
		ClassName:   quiz.ClassName,
		SubjectName: quiz.SubjectName,
		Topic:       quiz.Topic,
		Questions:   quiz.Questions,
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

// Get quiz from DynamoDB with filters
func GetQuizFromDynamoDB(quizName, className, subjectName, topic string) (*QuizItem, error) {
	result, err := dynamoClient.Scan(&dynamodb.ScanInput{
		TableName:        aws.String("quiz_questions"),
		FilterExpression: aws.String("quiz_name = :quizName AND class_name = :className AND subject_name = :subjectName AND topic = :topic"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":quizName":    {S: aws.String(quizName)},
			":className":   {S: aws.String(className)},
			":subjectName": {S: aws.String(subjectName)},
			":topic":       {S: aws.String(topic)},
		},
	})

	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	var quiz QuizItem
	err = dynamodbattribute.UnmarshalMap(result.Items[0], &quiz)
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

// Class Subject item structure
type ClassSubjectItem struct {
	ClassName   string   `json:"class_name" dynamodbav:"class_name"`
	SubjectName string   `json:"subject_name" dynamodbav:"subject_name"`
	Topics      []string `json:"topics" dynamodbav:"topics"`
}

// Class operations
func InsertClass(className string) error {
	// Insert a placeholder item for the class
	item := ClassSubjectItem{
		ClassName:   className,
		SubjectName: "_CLASS_PLACEHOLDER",
		Topics:      []string{},
	}

	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = dynamoClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("class_subjects"),
		Item:      av,
	})
	return err
}

func DeleteClass(className string) error {
	// Query all items for this class
	result, err := dynamoClient.Query(&dynamodb.QueryInput{
		TableName:              aws.String("class_subjects"),
		KeyConditionExpression: aws.String("class_name = :className"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":className": {S: aws.String(className)},
		},
	})
	if err != nil {
		return err
	}

	// Delete all items
	for _, item := range result.Items {
		_, err = dynamoClient.DeleteItem(&dynamodb.DeleteItemInput{
			TableName: aws.String("class_subjects"),
			Key: map[string]*dynamodb.AttributeValue{
				"class_name":   item["class_name"],
				"subject_name": item["subject_name"],
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func FetchClasses() ([]string, error) {
	result, err := dynamoClient.Scan(&dynamodb.ScanInput{
		TableName:        aws.String("class_subjects"),
		FilterExpression: aws.String("subject_name = :placeholder"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":placeholder": {S: aws.String("_CLASS_PLACEHOLDER")},
		},
	})
	if err != nil {
		return nil, err
	}

	var classes []string
	for _, item := range result.Items {
		if className, ok := item["class_name"]; ok && className.S != nil {
			classes = append(classes, *className.S)
		}
	}
	return classes, nil
}

// Subject operations
func InsertSubject(className, subjectName string) error {
	item := ClassSubjectItem{
		ClassName:   className,
		SubjectName: subjectName,
		Topics:      []string{},
	}

	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = dynamoClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("class_subjects"),
		Item:      av,
	})
	return err
}

func DeleteSubject(className, subjectName string) error {
	_, err := dynamoClient.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("class_subjects"),
		Key: map[string]*dynamodb.AttributeValue{
			"class_name":   {S: aws.String(className)},
			"subject_name": {S: aws.String(subjectName)},
		},
	})
	return err
}

func FetchSubjects(className string) ([]string, error) {
	result, err := dynamoClient.Query(&dynamodb.QueryInput{
		TableName:              aws.String("class_subjects"),
		KeyConditionExpression: aws.String("class_name = :className"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":className": {S: aws.String(className)},
		},
	})
	if err != nil {
		return nil, err
	}

	var subjects []string
	for _, item := range result.Items {
		if subjectName, ok := item["subject_name"]; ok && subjectName.S != nil {
			// Filter out placeholder in code instead of DynamoDB
			if *subjectName.S != "_CLASS_PLACEHOLDER" {
				subjects = append(subjects, *subjectName.S)
			}
		}
	}
	if subjects == nil {
		subjects = []string{}
	}
	return subjects, nil
}

// Topic operations
func InsertTopic(className, subjectName, topic string) error {
	// Get current item
	result, err := dynamoClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("class_subjects"),
		Key: map[string]*dynamodb.AttributeValue{
			"class_name":   {S: aws.String(className)},
			"subject_name": {S: aws.String(subjectName)},
		},
	})
	if err != nil {
		return err
	}

	var topics []string
	if result.Item != nil {
		var item ClassSubjectItem
		err = dynamodbattribute.UnmarshalMap(result.Item, &item)
		if err != nil {
			return err
		}
		topics = item.Topics
	}

	// Check if topic already exists
	for _, t := range topics {
		if t == topic {
			return nil // Already exists
		}
	}

	// Add new topic
	topics = append(topics, topic)

	// Update item
	item := ClassSubjectItem{
		ClassName:   className,
		SubjectName: subjectName,
		Topics:      topics,
	}

	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = dynamoClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("class_subjects"),
		Item:      av,
	})
	return err
}

func DeleteTopic(className, subjectName, topic string) error {
	// Get current item
	result, err := dynamoClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("class_subjects"),
		Key: map[string]*dynamodb.AttributeValue{
			"class_name":   {S: aws.String(className)},
			"subject_name": {S: aws.String(subjectName)},
		},
	})
	if err != nil {
		return err
	}

	if result.Item == nil {
		return nil // Item doesn't exist
	}

	var item ClassSubjectItem
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		return err
	}

	// Remove topic
	var newTopics []string
	for _, t := range item.Topics {
		if t != topic {
			newTopics = append(newTopics, t)
		}
	}

	// Update item
	item.Topics = newTopics
	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = dynamoClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("class_subjects"),
		Item:      av,
	})
	return err
}

func FetchTopics(className, subjectName string) ([]string, error) {
	result, err := dynamoClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("class_subjects"),
		Key: map[string]*dynamodb.AttributeValue{
			"class_name":   {S: aws.String(className)},
			"subject_name": {S: aws.String(subjectName)},
		},
	})
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return []string{}, nil
	}

	var item ClassSubjectItem
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		return nil, err
	}

	return item.Topics, nil
}
