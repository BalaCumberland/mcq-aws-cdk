#!/bin/bash
cd lambdas/golang-lambda-v2
env GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap main.go
zip function.zip bootstrap handlers/*.go
aws lambda update-function-code --function-name golang-upload-api-v2 --zip-file fileb://function.zip
rm bootstrap function.zip
echo "Lambda v2 deployed successfully!"