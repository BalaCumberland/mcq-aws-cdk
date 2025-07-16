#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import { BucketsStack } from '../lib/buckets-stack';
import { ApiLambdaStack } from '../lib/api-lambda-stack';
import { DatabaseStack } from '../lib/database-stack';
import { AmplifyStack } from '../lib/amplify-stack';

const app = new cdk.App();

const bucketsStack = new BucketsStack(app, 'BucketsStack', {
  env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: 'us-east-1' },
});

const dbStack = new DatabaseStack(app, 'DatabaseStack', {
  env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: 'us-east-1' },
  sqlBackupBucket: bucketsStack.sqlBackupBucket
});

new ApiLambdaStack(app, 'ApiLambdaStack', {
  env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: 'us-east-1' },
  vpc: dbStack.vpc,
  lambdaSecurityGroup: dbStack.lambdaSecurityGroup,
  dbSecurityGroup: dbStack.dbSecurityGroup,
  sqlBackupBucket: bucketsStack.sqlBackupBucket
});

new AmplifyStack(app, 'AmplifyStack', {
  env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: 'us-east-1' }
});