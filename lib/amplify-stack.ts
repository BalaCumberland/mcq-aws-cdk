import * as cdk from 'aws-cdk-lib';
import * as amplify from 'aws-cdk-lib/aws-amplify';
import { Construct } from 'constructs';

export class AmplifyStack extends cdk.Stack {
  public readonly appId: string;

  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const amplifyApp = new amplify.CfnApp(this, 'McqReactApp', {
      name: 'mcq-react-app',
      repository: 'https://github.com/BalaCumberland/mcq-app',
      platform: 'WEB',
      accessToken: process.env.GITHUB_TOKEN || 'YOUR_GITHUB_TOKEN',
      buildSpec: `
version: 1
frontend:
  phases:
    preBuild:
      commands:
        - npm ci
    build:
      commands:
        - npm run build
  artifacts:
    baseDirectory: dist
    files:
      - '**/*'
  cache:
    paths:
      - node_modules/**/*
      `,
      environmentVariables: [
        {
          name: 'AMPLIFY_DIFF_DEPLOY',
          value: 'false'
        }
      ]
    });

    const mainBranch = new amplify.CfnBranch(this, 'MainBranch', {
      appId: amplifyApp.attrAppId,
      branchName: 'main',
      enableAutoBuild: true,
      framework: 'React'
    });

    this.appId = amplifyApp.attrAppId;

    new cdk.CfnOutput(this, 'AmplifyAppUrl', {
      value: `https://main.${amplifyApp.attrAppId}.amplifyapp.com`,
      description: 'Amplify App URL'
    });

    new cdk.CfnOutput(this, 'AmplifyAppId', {
      value: amplifyApp.attrAppId,
      description: 'Amplify App ID'
    });
  }
}