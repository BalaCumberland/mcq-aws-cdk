#!/bin/bash
cd lambdas/authorizer-lambda
mkdir deployment
cd deployment
cp -r ../index.js ../package.json ../package-lock.json ../node_modules ./
zip -r ../function.zip ./*
cd ..
aws lambda update-function-code --function-name firebase-authorizer --zip-file fileb://function.zip
rm -rf deployment
echo "Authorizer deployed successfully!"