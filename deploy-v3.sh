#!/bin/bash
set -e

echo "🚀 Deploying Lambda V3..."

# Build TypeScript
npm run build

# Deploy CDK stack
npx cdk deploy ApiLambdaStackV3 --require-approval never

echo "✅ Lambda V3 deployed successfully!"