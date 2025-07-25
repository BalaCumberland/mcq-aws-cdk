import * as cdk from 'aws-cdk-lib';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import { Construct } from 'constructs';

export class DynamoDbStack extends cdk.Stack {
  public readonly quizTable: dynamodb.Table;
  public readonly studentTable: dynamodb.Table;
  public readonly studentInfoTable: dynamodb.Table;
  public readonly attemptsTable: dynamodb.Table;
  public readonly studentQuizzesTable: dynamodb.Table;
  public readonly classSubjectsTable: dynamodb.Table;

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

    // Students Info Table (with Firebase UID)
    this.studentInfoTable = new dynamodb.Table(this, 'StudentInfoTable', {
      tableName: 'students_info',
      partitionKey: { name: 'uid', type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.RETAIN
    });

    // Add GSI for email lookup
    this.studentInfoTable.addGlobalSecondaryIndex({
      indexName: 'email-index',
      partitionKey: { name: 'email', type: dynamodb.AttributeType.STRING }
    });

    // Student Quiz Attempts Table
    this.attemptsTable = new dynamodb.Table(this, 'AttemptsTable', {
      tableName: 'student_quiz_attempts_v2',
      partitionKey: { name: 'uid', type: dynamodb.AttributeType.STRING },
      sortKey: { name: 'quiz_name', type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.RETAIN
    });

    // Student Quizzes Table
    this.studentQuizzesTable = new dynamodb.Table(this, 'StudentQuizzesTable', {
      tableName: 'student_quizzes_v2',
      partitionKey: { name: 'uid', type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.RETAIN
    });

    // GSI for category queries
    this.attemptsTable.addGlobalSecondaryIndex({
      indexName: 'category-index',
      partitionKey: { name: 'category', type: dynamodb.AttributeType.STRING },
      sortKey: { name: 'attempted_at', type: dynamodb.AttributeType.STRING }
    });

    // Class Subjects Table
    this.classSubjectsTable = new dynamodb.Table(this, 'ClassSubjectsTable', {
      tableName: 'class_subjects',
      partitionKey: { name: 'class_name', type: dynamodb.AttributeType.STRING },
      sortKey: { name: 'subject_name', type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.RETAIN
    });
  }
}