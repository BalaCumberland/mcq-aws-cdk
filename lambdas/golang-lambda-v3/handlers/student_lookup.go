package handlers

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func findStudentUIDByEmail(email string) (string, error) {
	// Scan students table to find UID by email
	scanResult, err := dynamoClient.Scan(&dynamodb.ScanInput{
		TableName: aws.String("students_v3"),
		FilterExpression: aws.String("email = :email"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":email": {S: aws.String(email)},
		},
		ProjectionExpression: aws.String("uid"),
	})

	if err != nil {
		return "", err
	}

	if len(scanResult.Items) == 0 {
		return "", nil
	}

	var student struct {
		UID string `json:"uid" dynamodbav:"uid"`
	}
	err = dynamodbattribute.UnmarshalMap(scanResult.Items[0], &student)
	if err != nil {
		return "", err
	}

	return student.UID, nil
}

func findStudentUIDByPhone(phone string) (string, error) {
	// Scan students table to find UID by phone
	scanResult, err := dynamoClient.Scan(&dynamodb.ScanInput{
		TableName: aws.String("students_v3"),
		FilterExpression: aws.String("phone_number = :phone"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":phone": {S: aws.String(phone)},
		},
		ProjectionExpression: aws.String("uid"),
	})

	if err != nil {
		return "", err
	}

	if len(scanResult.Items) == 0 {
		return "", nil
	}

	var student struct {
		UID string `json:"uid" dynamodbav:"uid"`
	}
	err = dynamodbattribute.UnmarshalMap(scanResult.Items[0], &student)
	if err != nil {
		return "", err
	}

	return student.UID, nil
}