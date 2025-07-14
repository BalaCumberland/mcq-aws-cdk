package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"google.golang.org/api/option"
)

var firebaseAuth *auth.Client

func getFirebaseCredentials() ([]byte, error) {
	sess := session.Must(session.NewSession())
	svc := secretsmanager.New(sess)

	result, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String("mcq-app/firebase-service-account"),
	})
	if err != nil {
		return nil, err
	}

	return []byte(*result.SecretString), nil
}

func initFirebase() error {
	ctx := context.Background()
	
	credsJSON, err := getFirebaseCredentials()
	if err != nil {
		return fmt.Errorf("failed to get Firebase credentials: %v", err)
	}

	conf := &firebase.Config{}
	app, err := firebase.NewApp(ctx, conf, option.WithCredentialsJSON(credsJSON))
	if err != nil {
		return fmt.Errorf("error initializing firebase app: %v", err)
	}
	
	firebaseAuth, err = app.Auth(ctx)
	if err != nil {
		return fmt.Errorf("error initializing firebase auth client: %v", err)
	}
	return nil
}

func handler(ctx context.Context, event events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	log.Printf("🔐 Authorizer started")

	authHeader := event.AuthorizationToken
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return events.APIGatewayCustomAuthorizerResponse{}, fmt.Errorf("Unauthorized")
	}

	idToken := strings.TrimPrefix(authHeader, "Bearer ")

	authCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	token, err := firebaseAuth.VerifyIDToken(authCtx, idToken)
	if err != nil {
		return events.APIGatewayCustomAuthorizerResponse{}, fmt.Errorf("Unauthorized")
	}

	userEmail := token.Claims["email"].(string)
	log.Printf("✅ Token verified for user: %s", userEmail)

	return events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: token.UID,
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   "Allow",
					Resource: []string{event.MethodArn},
				},
			},
		},
		Context: map[string]interface{}{
			"email": userEmail,
			"uid":   token.UID,
		},
	}, nil
}

func main() {
	if err := initFirebase(); err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}
	lambda.Start(handler)
}