import * as cdk from 'aws-cdk-lib';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as iam from 'aws-cdk-lib/aws-iam';
import { Construct } from 'constructs';

export class DynamoDbStack extends cdk.Stack {
  public readonly quizTable: dynamodb.Table;
  public readonly studentTable: dynamodb.Table;
  public readonly attemptsTable: dynamodb.Table;
  public readonly studentQuizzesTable: dynamodb.Table;

  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // Quiz Questions Table
    this.quizTable = new dynamodb.Table(this, 'QuizTable', {
      tableName: 'quiz_questions',
      partitionKey: { name: 'quiz_name', type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.RETAIN
    });

    // Students Table
    this.studentTable = new dynamodb.Table(this, 'StudentTable', {
      tableName: 'students',
      partitionKey: { name: 'email', type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.RETAIN
    });

    // Student Quiz Attempts Table
    this.attemptsTable = new dynamodb.Table(this, 'AttemptsTable', {
      tableName: 'student_quiz_attempts',
      partitionKey: { name: 'email', type: dynamodb.AttributeType.STRING },
      sortKey: { name: 'quiz_name', type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.RETAIN
    });

    // Student Quizzes Table
    this.studentQuizzesTable = new dynamodb.Table(this, 'StudentQuizzesTable', {
      tableName: 'student_quizzes',
      partitionKey: { name: 'email', type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.RETAIN
    });

    // GSI for category queries
    this.attemptsTable.addGlobalSecondaryIndex({
      indexName: 'category-index',
      partitionKey: { name: 'category', type: dynamodb.AttributeType.STRING },
      sortKey: { name: 'attempted_at', type: dynamodb.AttributeType.STRING }
    });
  }
}