import * as cdk from 'aws-cdk-lib';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as s3 from 'aws-cdk-lib/aws-s3';
import { Construct } from 'constructs';

interface ApiLambdaStackProps extends cdk.StackProps {
  vpc?: ec2.Vpc;
  dbSecurityGroup?: ec2.SecurityGroup;
  lambdaSecurityGroup?: ec2.SecurityGroup;
  sqlBackupBucket?: s3.Bucket;
}

export class ApiLambdaStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: ApiLambdaStackProps) {
    super(scope, id, props);

    // Configure passed Lambda Security Group
    if (props?.lambdaSecurityGroup && props.dbSecurityGroup) {
      // Database access only
      props.lambdaSecurityGroup.addEgressRule(
        ec2.Peer.securityGroupId(props.dbSecurityGroup.securityGroupId),
        ec2.Port.tcp(5432),
        'Allow PostgreSQL access'
      );
    }

    // Firebase Authorizer Lambda (outside VPC)
    const authorizerLambda = new lambda.Function(this, 'FirebaseAuthorizer', {
      functionName: 'firebase-authorizer',
      runtime: lambda.Runtime.NODEJS_20_X,
      handler: 'index.handler',
      architecture: lambda.Architecture.ARM_64,
      code: lambda.Code.fromAsset('lambdas/authorizer-lambda'),
      timeout: cdk.Duration.seconds(30),
      memorySize: 256,
      tracing: lambda.Tracing.ACTIVE,
      environment: {
        FIREBASE_PROJECT_ID: 'gothinkersteach',
        FIREBASE_PRIVATE_KEY: '',
        FIREBASE_CLIENT_EMAIL: ''
      }
    });

    // V1 Lambda removed - keeping code files

    // V2 Lambda Function with DynamoDB
    const goLambdaV2 = new lambda.Function(this, 'GolangUploadApiV2', {
      functionName: 'golang-upload-api-v2',
      runtime: lambda.Runtime.PROVIDED_AL2023,
      handler: 'bootstrap',
      architecture: lambda.Architecture.ARM_64,
      code: lambda.Code.fromAsset('lambdas/golang-lambda-v2'),
      timeout: cdk.Duration.seconds(300),
      memorySize: 512,
      tracing: lambda.Tracing.ACTIVE,
      vpc: props?.vpc,
      vpcSubnets: {
        subnets: [
          ec2.Subnet.fromSubnetId(this, 'PrivateSubnet1V2', 'subnet-08a50ae7e49889508'),
          ec2.Subnet.fromSubnetId(this, 'PrivateSubnet2V2', 'subnet-0fc698411d47497d8')
        ]
      },
      securityGroups: props?.lambdaSecurityGroup ? [props.lambdaSecurityGroup] : undefined
    });

    // Migration Lambda removed

    // Add DynamoDB permissions to V2 Lambda
    goLambdaV2.addToRolePolicy(new iam.PolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: [
        'dynamodb:GetItem',
        'dynamodb:PutItem',
        'dynamodb:UpdateItem',
        'dynamodb:DeleteItem',
        'dynamodb:Query',
        'dynamodb:Scan'
      ],
      resources: [
        'arn:aws:dynamodb:*:*:table/quiz_questions',
        'arn:aws:dynamodb:*:*:table/students', 
        'arn:aws:dynamodb:*:*:table/students_info',
        'arn:aws:dynamodb:*:*:table/students_info/index/*',
        'arn:aws:dynamodb:*:*:table/student_quiz_attempts',
        'arn:aws:dynamodb:*:*:table/student_quiz_attempts/index/*',
        'arn:aws:dynamodb:*:*:table/student_quizzes',
        'arn:aws:dynamodb:*:*:table/class_subjects'
      ]
    }));

    // CloudWatch role for API Gateway logging
    const apiGatewayCloudWatchRole = new iam.Role(this, 'ApiGatewayCloudWatchRole', {
      assumedBy: new iam.ServicePrincipal('apigateway.amazonaws.com'),
      managedPolicies: [
        iam.ManagedPolicy.fromAwsManagedPolicyName('service-role/AmazonAPIGatewayPushToCloudWatchLogs')
      ]
    });

    // Set CloudWatch role for API Gateway account
    new apigateway.CfnAccount(this, 'ApiGatewayAccount', {
      cloudWatchRoleArn: apiGatewayCloudWatchRole.roleArn
    });

    // API Gateway with CORS
    const api = new apigateway.RestApi(this, 'McqApi', {
      binaryMediaTypes: ['multipart/form-data'],
      defaultCorsPreflightOptions: {
        allowOrigins: apigateway.Cors.ALL_ORIGINS,
        allowMethods: apigateway.Cors.ALL_METHODS,
        allowHeaders: ['Content-Type', 'Authorization']
      },
      deployOptions: {
        loggingLevel: apigateway.MethodLoggingLevel.INFO,
        dataTraceEnabled: true,
        metricsEnabled: true,
        tracingEnabled: true
      }
    });

    // API Gateway authorizer
    const authorizer = new apigateway.TokenAuthorizer(this, 'ApiAuthorizer', {
      handler: authorizerLambda,
      identitySource: 'method.request.header.Authorization'
    });

    // API Gateway integrations
    const goV2Integration = new apigateway.LambdaIntegration(goLambdaV2);

    // Root resources
    const studentsResource = api.root.addResource('students');
    const quizResource = api.root.addResource('quiz');

    // V2 routes with DynamoDB
    const v2Resource = api.root.addResource('v2');
    const v2ProxyResource = v2Resource.addProxy({
      defaultIntegration: goV2Integration,
      anyMethod: false
    });
    
    v2ProxyResource.addMethod('ANY', goV2Integration, {
      authorizer: authorizer,
      apiKeyRequired: false
    });

    // V2 register endpoint without authorization
    const v2StudentsResource = v2Resource.addResource('students');
    const v2StudentsRegisterResource = v2StudentsResource.addResource('register');
    v2StudentsRegisterResource.addMethod('POST', goV2Integration, {
      apiKeyRequired: false
    });

    // Root proxy removed

    // Restrict Lambda access to API Gateway only
    goLambdaV2.addPermission('ApiGatewayInvokeV2', {
      principal: new iam.ServicePrincipal('apigateway.amazonaws.com'),
      sourceArn: api.arnForExecuteApi()
    });

    // No Secrets Manager permissions needed - using environment variables

    // Output API URL
    new cdk.CfnOutput(this, 'ApiUrl', {
      value: api.url
    });
  }
}