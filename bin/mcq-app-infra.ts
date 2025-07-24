#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import { BucketsStack } from '../lib/buckets-stack';
import { ApiLambdaStack } from '../lib/api-lambda-stack';
import { ApiLambdaStackV3 } from '../lib/api-lambda-stack-v3';
import { DatabaseStack } from '../lib/database-stack';
import { DynamoDbStack } from '../lib/dynamodb-stack';
import { AmplifyStack } from '../lib/amplify-stack';

const app = new cdk.App();

const bucketsStack = new BucketsStack(app, 'BucketsStack', {
  env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: 'us-east-1' },
});

const dbStack = new DatabaseStack(app, 'DatabaseStack', {
  env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: 'us-east-1' },
  sqlBackupBucket: bucketsStack.sqlBackupBucket
});

const dynamoStack = new DynamoDbStack(app, 'DynamoDbStack', {
  env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: 'us-east-1' }
});

new ApiLambdaStack(app, 'ApiLambdaStack', {
  env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: 'us-east-1' },
  vpc: dbStack.vpc,
  lambdaSecurityGroup: dbStack.lambdaSecurityGroup,
  dbSecurityGroup: dbStack.dbSecurityGroup,
  sqlBackupBucket: bucketsStack.sqlBackupBucket
});

new ApiLambdaStackV3(app, 'ApiLambdaStackV3', {
  env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: 'us-east-1' }
});

new AmplifyStack(app, 'AmplifyStack', {
  env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: 'us-east-1' }
});