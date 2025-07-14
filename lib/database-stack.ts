import * as cdk from 'aws-cdk-lib';
import * as rds from 'aws-cdk-lib/aws-rds';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as secretsmanager from 'aws-cdk-lib/aws-secretsmanager';
import { Construct } from 'constructs';

export class DatabaseStack extends cdk.Stack {
  public readonly vpc: ec2.Vpc;
  public readonly dbSecurityGroup: ec2.SecurityGroup;
  public readonly lambdaSecurityGroup: ec2.SecurityGroup;
  public readonly dbSecret: secretsmanager.Secret;

  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // VPC with explicit internet gateway
    this.vpc = new ec2.Vpc(this, 'McqVpc', {
      maxAzs: 2,
      natGateways: 0,
      enableDnsHostnames: true,
      enableDnsSupport: true,
      ipAddresses: ec2.IpAddresses.cidr('10.1.0.0/16'),
      subnetConfiguration: [
        {
          cidrMask: 24,
          name: 'Public',
          subnetType: ec2.SubnetType.PUBLIC
        },
        {
          cidrMask: 24,
          name: 'Database',
          subnetType: ec2.SubnetType.PRIVATE_ISOLATED
        }
      ]
    });

    // Security Group for Lambda
    this.lambdaSecurityGroup = new ec2.SecurityGroup(this, 'LambdaSecurityGroup', {
      vpc: this.vpc,
      description: 'Security group for Lambda functions'
    });

    // Security Group for Database
    this.dbSecurityGroup = new ec2.SecurityGroup(this, 'DbSecurityGroup', {
      vpc: this.vpc,
      description: 'Security group for PostgreSQL database'
    });

    // Allow Lambda SG to access RDS SG on port 5432
    this.dbSecurityGroup.addIngressRule(
      ec2.Peer.securityGroupId(this.lambdaSecurityGroup.securityGroupId),
      ec2.Port.tcp(5432),
      'Lambda access to database'
    );

    // Security Group for Bastion Host
    const bastionSecurityGroup = new ec2.SecurityGroup(this, 'BastionSecurityGroup', {
      vpc: this.vpc,
      description: 'Security group for bastion host'
    });

    // No SSH access needed - using Session Manager

    // Allow bastion to access database
    this.dbSecurityGroup.addIngressRule(
      ec2.Peer.securityGroupId(bastionSecurityGroup.securityGroupId),
      ec2.Port.tcp(5432),
      'Bastion access to database'
    );

    // Bastion Host with Session Manager
    const bastion = new ec2.Instance(this, 'BastionHost', {
      vpc: this.vpc,
      instanceType: ec2.InstanceType.of(ec2.InstanceClass.T3, ec2.InstanceSize.MICRO),
      machineImage: ec2.MachineImage.latestAmazonLinux2023(),
      securityGroup: bastionSecurityGroup,
      vpcSubnets: {
        subnetType: ec2.SubnetType.PUBLIC
      },
      role: new iam.Role(this, 'BastionRole', {
        assumedBy: new iam.ServicePrincipal('ec2.amazonaws.com'),
        managedPolicies: [
          iam.ManagedPolicy.fromAwsManagedPolicyName('AmazonSSMManagedInstanceCore')
        ],
        inlinePolicies: {
          S3Access: new iam.PolicyDocument({
            statements: [
              new iam.PolicyStatement({
                effect: iam.Effect.ALLOW,
                actions: ['s3:GetObject', 's3:ListBucket'],
                resources: [
                  'arn:aws:s3:::mcq-sqlbackup-536697228264',
                  'arn:aws:s3:::mcq-sqlbackup-536697228264/*'
                ]
              })
            ]
          })
        }
      })
    });

    // Database credentials
    this.dbSecret = new secretsmanager.Secret(this, 'DbSecret', {
      secretName: 'dbcredentials/postgres',
      generateSecretString: {
        secretStringTemplate: JSON.stringify({ username: 'postgres' }),
        generateStringKey: 'password',
        excludeCharacters: ' %+~`#$&*()|[]{}:;<>?!\'/"\\'
      }
    });

    // Firebase service account secret
    new secretsmanager.Secret(this, 'FirebaseSecret', {
      secretName: 'mcq-app/firebase-service-account',
      description: 'Firebase service account credentials for MCQ app'
    });

    // PostgreSQL Database (Free Tier)
    const database = new rds.DatabaseInstance(this, 'PostgresDb', {
      engine: rds.DatabaseInstanceEngine.POSTGRES,
      instanceType: ec2.InstanceType.of(ec2.InstanceClass.T3, ec2.InstanceSize.MICRO),
      vpc: this.vpc,
      securityGroups: [this.dbSecurityGroup],
      databaseName: 'mcqdb',
      instanceIdentifier: 'mcq-db',
      credentials: rds.Credentials.fromSecret(this.dbSecret, 'postgres'),
      allocatedStorage: 20,
      storageType: rds.StorageType.GP2,
      deleteAutomatedBackups: true,
      backupRetention: cdk.Duration.days(7),
      deletionProtection: false,
      publiclyAccessible: false,
      vpcSubnets: {
        subnetType: ec2.SubnetType.PRIVATE_ISOLATED
      }
    });

    // VPC Interface Endpoint for Secrets Manager
    this.vpc.addInterfaceEndpoint('SecretsManagerEndpoint', {
      service: ec2.InterfaceVpcEndpointAwsService.SECRETS_MANAGER,
      subnets: {
        subnetType: ec2.SubnetType.PUBLIC
      },
      securityGroups: [this.lambdaSecurityGroup]
    });

    // Outputs
    new cdk.CfnOutput(this, 'DatabaseEndpoint', {
      value: database.instanceEndpoint.hostname
    });

    new cdk.CfnOutput(this, 'DatabasePort', {
      value: database.instanceEndpoint.port.toString()
    });

    new cdk.CfnOutput(this, 'BastionHostPublicIp', {
      value: bastion.instancePublicIp
    });
  }
}