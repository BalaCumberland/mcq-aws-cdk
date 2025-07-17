#!/bin/bash
cd lambdas/golang-lambda
env GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap main.go
zip function.zip bootstrap
aws lambda update-function-code --function-name golang-upload-api --zip-file fileb://function.zip
echo "Lambda deployed successfully!"