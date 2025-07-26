#!/bin/bash
set -e

echo "🚀 Deploying Firebase Authorizer..."

# Navigate to authorizer directory
cd "$(dirname "$0")/../lambdas/authorizer-lambda"

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
    echo "📦 Installing dependencies..."
    npm install
fi

# Create deployment package
echo "📦 Creating deployment package..."
rm -rf deployment function.zip
mkdir deployment
cp index.js package.json deployment/
cp -r node_modules deployment/

# Create zip file
cd deployment
zip -r ../function.zip . > /dev/null
cd ..

# Deploy to AWS Lambda
echo "☁️ Updating Lambda function..."
aws lambda update-function-code \
    --function-name firebase-authorizer \
    --zip-file fileb://function.zip

# Cleanup
rm -rf deployment

echo "✅ Authorizer deployed successfully!"