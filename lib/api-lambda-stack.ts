import * as cdk from 'aws-cdk-lib';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import { Construct } from 'constructs';

interface ApiLambdaStackProps extends cdk.StackProps {
  vpc?: ec2.Vpc;
  dbSecurityGroup?: ec2.SecurityGroup;
  lambdaSecurityGroup?: ec2.SecurityGroup;
  dbSecretArn?: string;
  firebaseSecretArn?: string;
}

export class ApiLambdaStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: ApiLambdaStackProps) {
    super(scope, id, props);

    // Configure passed Lambda Security Group
    if (props?.lambdaSecurityGroup) {
      // HTTPS internet access
      props.lambdaSecurityGroup.addEgressRule(
        ec2.Peer.anyIpv4(),
        ec2.Port.tcp(443),
        'Allow HTTPS internet access'
      );
      
      // Database access
      if (props.dbSecurityGroup) {
        props.lambdaSecurityGroup.addEgressRule(
          ec2.Peer.securityGroupId(props.dbSecurityGroup.securityGroupId),
          ec2.Port.tcp(5432),
          'Allow PostgreSQL access'
        );
      }
    }

    // Golang Lambda Function in public subnet
    const goLambda = new lambda.Function(this, 'GolangUploadApi', {
      functionName: 'golang-upload-api',
      runtime: lambda.Runtime.PROVIDED_AL2023,
      handler: 'bootstrap',
      architecture: lambda.Architecture.ARM_64,
      code: lambda.Code.fromAsset('golang-lambda'),
      timeout: cdk.Duration.seconds(60),
      memorySize: 512,

      environment: {
        DB_SECRET_ARN: props?.dbSecretArn || '',
        FIREBASE_SECRET_ARN: props?.firebaseSecretArn || ''
      }
    });

    // API Gateway with CORS
    const api = new apigateway.RestApi(this, 'McqApi', {
      defaultCorsPreflightOptions: {
        allowOrigins: apigateway.Cors.ALL_ORIGINS,
        allowMethods: apigateway.Cors.ALL_METHODS,
        allowHeaders: ['Content-Type', 'Authorization']
      }
    });

    // API Gateway integrations
    const goIntegration = new apigateway.LambdaIntegration(goLambda);

    // No API key or usage plan - open access with Firebase auth only

    // API routes at root path - all methods
    api.root.addMethod('ANY', goIntegration, {
      apiKeyRequired: false
    });

    // Add proxy resource to handle all sub-paths
    const proxyResource = api.root.addProxy({
      defaultIntegration: goIntegration,
      anyMethod: true
    });

    // Restrict Lambda access to API Gateway only
    goLambda.addPermission('ApiGatewayInvoke', {
      principal: new iam.ServicePrincipal('apigateway.amazonaws.com'),
      sourceArn: api.arnForExecuteApi()
    });

    // Grant Lambda permission to read secrets
    if (props?.dbSecretArn) {
      goLambda.addToRolePolicy(new iam.PolicyStatement({
        effect: iam.Effect.ALLOW,
        actions: ['secretsmanager:GetSecretValue'],
        resources: [
          props.dbSecretArn,
          'arn:aws:secretsmanager:us-east-1:536697228264:secret:mcq-app/firebase-service-account-*'
        ]
      }));
    }

    // Output API URL
    new cdk.CfnOutput(this, 'ApiUrl', {
      value: api.url
    });
  }
}