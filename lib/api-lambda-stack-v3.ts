import * as cdk from 'aws-cdk-lib';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as iam from 'aws-cdk-lib/aws-iam';
import { Construct } from 'constructs';

export class ApiLambdaStackV3 extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // Create DynamoDB tables for V3
    const studentsTableV3 = new dynamodb.Table(this, 'StudentsTableV3', {
      tableName: 'students_v3',
      partitionKey: { name: 'uid', type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.RETAIN,
    });

    const attemptsTableV3 = new dynamodb.Table(this, 'AttemptsTableV3', {
      tableName: 'student_quiz_attempts_v3',
      partitionKey: { name: 'uid', type: dynamodb.AttributeType.STRING },
      sortKey: { name: 'quiz_name', type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.RETAIN,
    });

    // Create V3 Authorizer Lambda
    const authorizerV3 = new lambda.Function(this, 'AuthorizerV3', {
      runtime: lambda.Runtime.NODEJS_18_X,
      handler: 'index.handler',
      code: lambda.Code.fromAsset('lambdas/authorizer-lambda-v3'),
      environment: {
        FIREBASE_PROJECT_ID: process.env.FIREBASE_PROJECT_ID || '',
        FIREBASE_PRIVATE_KEY_ID: process.env.FIREBASE_PRIVATE_KEY_ID || '',
        FIREBASE_PRIVATE_KEY: process.env.FIREBASE_PRIVATE_KEY || '',
        FIREBASE_CLIENT_EMAIL: process.env.FIREBASE_CLIENT_EMAIL || '',
        FIREBASE_CLIENT_ID: process.env.FIREBASE_CLIENT_ID || '',
        FIREBASE_CLIENT_CERT_URL: process.env.FIREBASE_CLIENT_CERT_URL || '',
      },
      timeout: cdk.Duration.seconds(30),
    });

    // Create V3 Go Lambda
    const goLambdaV3 = new lambda.Function(this, 'GolangUploadApiV3', {
      runtime: lambda.Runtime.PROVIDED_AL2023,
      handler: 'bootstrap',
      code: lambda.Code.fromAsset('lambdas/golang-lambda-v3'),
      timeout: cdk.Duration.seconds(300),
      memorySize: 512,
      architecture: lambda.Architecture.ARM_64,
      environment: {
        AWS_LAMBDA_EXEC_WRAPPER: '/opt/otel-instrument',
      },
      tracing: lambda.Tracing.ACTIVE,
    });

    // Grant DynamoDB permissions
    studentsTableV3.grantReadWriteData(goLambdaV3);
    attemptsTableV3.grantReadWriteData(goLambdaV3);
    
    // Grant access to existing quiz_questions table
    goLambdaV3.addToRolePolicy(new iam.PolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: [
        'dynamodb:GetItem',
        'dynamodb:PutItem',
        'dynamodb:UpdateItem',
        'dynamodb:DeleteItem',
        'dynamodb:Query',
        'dynamodb:Scan'
      ],
      resources: ['arn:aws:dynamodb:us-east-1:*:table/quiz_questions']
    }));

    // Create API Gateway
    const api = new apigateway.RestApi(this, 'McqApiV3', {
      restApiName: 'MCQ Service V3',
      description: 'This service serves MCQ application V3.',
      defaultCorsPreflightOptions: {
        allowOrigins: apigateway.Cors.ALL_ORIGINS,
        allowMethods: apigateway.Cors.ALL_METHODS,
        allowHeaders: ['Content-Type', 'X-Amz-Date', 'Authorization', 'X-Api-Key'],
      },
    });

    // Create authorizer with no caching
    const authorizer = new apigateway.TokenAuthorizer(this, 'FirebaseAuthorizerV3NoCache', {
      handler: authorizerV3,
      resultsCacheTtl: cdk.Duration.seconds(0),
    });

    // Add single resource-based permission for all API Gateway methods
    goLambdaV3.addPermission('ApiGatewayInvoke', {
      principal: new iam.ServicePrincipal('apigateway.amazonaws.com'),
      sourceArn: `${api.arnForExecuteApi()}/*/*`
    });

    // Create Lambda integration
    const goV3Integration = new apigateway.LambdaIntegration(goLambdaV3);

    // V3 endpoints
    const v3Resource = api.root.addResource('v3');
    const v3StudentsResource = v3Resource.addResource('students');
    const v3QuizResource = v3Resource.addResource('quiz');
    const v3UploadResource = v3Resource.addResource('upload');
    
    // V3 Students endpoints
    v3StudentsResource.addResource('register').addMethod('POST', goV3Integration, { apiKeyRequired: false });
    v3StudentsResource.addResource('get').addMethod('GET', goV3Integration, { authorizer: authorizer, apiKeyRequired: false });
    v3StudentsResource.addResource('lookup').addMethod('GET', goV3Integration, { authorizer: authorizer, apiKeyRequired: false });
    v3StudentsResource.addResource('update').addMethod('POST', goV3Integration, { authorizer: authorizer, apiKeyRequired: false });
    v3StudentsResource.addResource('progress').addMethod('GET', goV3Integration, { authorizer: authorizer, apiKeyRequired: false });
    v3StudentsResource.addResource('class-upgrade').addMethod('POST', goV3Integration, { authorizer: authorizer, apiKeyRequired: false });

    // V3 Quiz endpoints
    v3QuizResource.addResource('get-by-name').addMethod('GET', goV3Integration, { authorizer: authorizer, apiKeyRequired: false });
    v3QuizResource.addResource('submit').addMethod('POST', goV3Integration, { authorizer: authorizer, apiKeyRequired: false });
    v3QuizResource.addResource('unattempted-quizzes').addMethod('GET', goV3Integration, { authorizer: authorizer, apiKeyRequired: false });
    v3QuizResource.addResource('delete').addMethod('DELETE', goV3Integration, { authorizer: authorizer, apiKeyRequired: false });
    v3QuizResource.addResource('result').addMethod('GET', goV3Integration, { authorizer: authorizer, apiKeyRequired: false });

    // V3 Upload endpoints
    v3UploadResource.addResource('questions').addMethod('POST', goV3Integration, { authorizer: authorizer, apiKeyRequired: false });



    // Output API URL
    new cdk.CfnOutput(this, 'ApiUrlV3', {
      value: api.url,
      description: 'URL of the API Gateway V3',
    });
  }
}