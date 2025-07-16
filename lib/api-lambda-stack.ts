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
      code: lambda.Code.fromAsset('authorizer-lambda'),
      timeout: cdk.Duration.seconds(30),
      memorySize: 256,
      tracing: lambda.Tracing.ACTIVE,
      environment: {
        FIREBASE_PROJECT_ID: 'gothinkersteach',
        FIREBASE_PRIVATE_KEY: '',
        FIREBASE_CLIENT_EMAIL: ''
      }
    });

    // Main Lambda Function in VPC public subnet
    const goLambda = new lambda.Function(this, 'GolangUploadApi', {
      functionName: 'golang-upload-api',
      runtime: lambda.Runtime.PROVIDED_AL2023,
      handler: 'bootstrap',
      architecture: lambda.Architecture.ARM_64,
      code: lambda.Code.fromAsset('golang-lambda'),
      timeout: cdk.Duration.seconds(300),
      memorySize: 512,
      tracing: lambda.Tracing.ACTIVE,
      vpc: props?.vpc,
      vpcSubnets: {
        subnets: [
          ec2.Subnet.fromSubnetId(this, 'PrivateSubnet1', 'subnet-08a50ae7e49889508'),
          ec2.Subnet.fromSubnetId(this, 'PrivateSubnet2', 'subnet-0fc698411d47497d8')
        ]
      },
      securityGroups: props?.lambdaSecurityGroup ? [props.lambdaSecurityGroup] : undefined,
      environment: {
        DB_HOST: '',
        DB_PORT: '',
        DB_NAME: '',
        DB_USER: '',
        DB_PASSWORD: ''
      }
    });

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
    const goIntegration = new apigateway.LambdaIntegration(goLambda);

    // Register endpoint without authorization
    const registerResource = api.root.addResource('register');
    registerResource.addMethod('POST', goIntegration, {
      apiKeyRequired: false
    });

    // Students register endpoint without authorization
    const studentsResource = api.root.addResource('students');
    const studentsRegisterResource = studentsResource.addResource('register');
    studentsRegisterResource.addMethod('POST', goIntegration, {
      apiKeyRequired: false
    });

    // All other routes with Firebase authorization
    const proxyResource = api.root.addProxy({
      defaultIntegration: goIntegration,
      anyMethod: false
    });
    
    proxyResource.addMethod('ANY', goIntegration, {
      authorizer: authorizer,
      apiKeyRequired: false
    });

    // Restrict Lambda access to API Gateway only
    goLambda.addPermission('ApiGatewayInvoke', {
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