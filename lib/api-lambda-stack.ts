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
      runtime: lambda.Runtime.PROVIDED_AL2023,
      handler: 'bootstrap',
      architecture: lambda.Architecture.ARM_64,
      code: lambda.Code.fromAsset('authorizer-lambda'),
      timeout: cdk.Duration.seconds(30),
      memorySize: 256,
      environment: {
        FIREBASE_SECRET_ARN: props?.firebaseSecretArn || ''
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
      vpc: props?.vpc,
      vpcSubnets: {
        subnets: [
          ec2.Subnet.fromSubnetId(this, 'PrivateSubnet1', 'subnet-08a50ae7e49889508'),
          ec2.Subnet.fromSubnetId(this, 'PrivateSubnet2', 'subnet-0fc698411d47497d8')
        ]
      },
      securityGroups: props?.lambdaSecurityGroup ? [props.lambdaSecurityGroup] : undefined,
      environment: {
        DB_SECRET_ARN: props?.dbSecretArn || ''
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

    // Grant authorizer Lambda permission to read Firebase secrets
    authorizerLambda.addToRolePolicy(new iam.PolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: ['secretsmanager:GetSecretValue'],
      resources: [
        'arn:aws:secretsmanager:us-east-1:536697228264:secret:mcq-app/firebase-service-account-*'
      ]
    }));

    // Grant main Lambda permission to read DB secrets
    if (props?.dbSecretArn) {
      goLambda.addToRolePolicy(new iam.PolicyStatement({
        effect: iam.Effect.ALLOW,
        actions: ['secretsmanager:GetSecretValue'],
        resources: [
          props.dbSecretArn
        ]
      }));
    }

    // Output API URL
    new cdk.CfnOutput(this, 'ApiUrl', {
      value: api.url
    });
  }
}