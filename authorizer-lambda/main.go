package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"google.golang.org/api/option"
)

var firebaseAuth *auth.Client

func getFirebaseCredentials() ([]byte, error) {
	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	privateKey := os.Getenv("FIREBASE_PRIVATE_KEY")
	clientEmail := os.Getenv("FIREBASE_CLIENT_EMAIL")

	if projectID == "" || privateKey == "" || clientEmail == "" {
		return nil, fmt.Errorf("missing Firebase environment variables")
	}

	credsJSON := fmt.Sprintf(`{
		"type": "service_account",
		"project_id": "%s",
		"private_key": "%s",
		"client_email": "%s"
	}`, projectID, privateKey, clientEmail)

	return []byte(credsJSON), nil
}

func initFirebase() error {
	ctx := context.Background()

	credsJSON, err := getFirebaseCredentials()
	if err != nil {
		return fmt.Errorf("failed to get Firebase credentials: %v", err)
	}

	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsJSON(credsJSON))
	if err != nil {
		return fmt.Errorf("error initializing Firebase app: %v", err)
	}

	firebaseAuth, err = app.Auth(ctx)
	if err != nil {
		return fmt.Errorf("error initializing Firebase auth client: %v", err)
	}
	return nil
}

func handler(ctx context.Context, event events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	log.Println("🔐 Authorizer started")

	authHeader := event.AuthorizationToken
	if !strings.HasPrefix(authHeader, "Bearer ") {
		log.Println("❌ Missing Bearer prefix in token")
		return events.APIGatewayCustomAuthorizerResponse{}, fmt.Errorf("Unauthorized")
	}

	idToken := strings.TrimPrefix(authHeader, "Bearer ")

	// Lazy init Firebase client
	if firebaseAuth == nil {
		log.Println("🔄 Firebase not initialized, initializing now...")
		if err := initFirebase(); err != nil {
			log.Printf("❌ Firebase init failed: %v", err)
			return events.APIGatewayCustomAuthorizerResponse{}, fmt.Errorf("Unauthorized")
		}
	}

	authCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	token, err := firebaseAuth.VerifyIDToken(authCtx, idToken)
	if err != nil {
		log.Printf("❌ Firebase token verification failed: %v", err)
		return events.APIGatewayCustomAuthorizerResponse{}, fmt.Errorf("Unauthorized")
	}

	userEmail, _ := token.Claims["email"].(string)
	log.Printf("✅ Token verified for user: %s", userEmail)

	// Construct wildcard ARN: arn:aws:execute-api:<region>:<account>:<api-id>/<stage>/*/*
	parts := strings.Split(event.MethodArn, ":")
	if len(parts) < 6 {
		log.Println("❌ Malformed MethodArn")
		return events.APIGatewayCustomAuthorizerResponse{}, fmt.Errorf("Unauthorized")
	}

	arnPrefix := strings.Join(parts[:5], ":")     // arn:aws:execute-api:<region>:<account>:<api-id>
	resourceParts := strings.Split(parts[5], "/") // <api-id>/<stage>/<method>/...

	if len(resourceParts) < 2 {
		log.Println("❌ Malformed resource section in MethodArn")
		return events.APIGatewayCustomAuthorizerResponse{}, fmt.Errorf("Unauthorized")
	}

	apiID := resourceParts[0]
	stage := resourceParts[1]

	wildcardResource := fmt.Sprintf("%s:%s/%s/*/*", arnPrefix, apiID, stage)

	return events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: token.UID,
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   "Allow",
					Resource: []string{wildcardResource},
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
		log.Printf("⚠️ Firebase init failed (will retry on demand): %v", err)
	}
	lambda.Start(handler)
}
