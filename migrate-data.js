const { Client } = require('pg');
const AWS = require('aws-sdk');

const dynamodb = new AWS.DynamoDB.DocumentClient({ region: 'us-east-1' });

const pgClient = new Client({
  host: 'mcq-db.cxseo0q6o4fc.us-east-1.rds.amazonaws.com',
  port: 5432,
  database: 'mcqdb',
  user: 'postgres',
  password: 'hy6HCu,aNANvIkX3jnqdBNxiPku^tR'
});

async function migrateData() {
  try {
    await pgClient.connect();
    console.log('✅ Connected to PostgreSQL');

    // Migrate quiz_questions
    console.log('📋 Migrating quiz questions...');
    const quizResult = await pgClient.query('SELECT * FROM quiz_questions');
    
    for (const row of quizResult.rows) {
      const item = {
        quiz_name: row.quiz_name,
        duration: row.duration,
        category: row.category,
        questions: row.questions
      };
      
      await dynamodb.put({
        TableName: 'quiz_questions',
        Item: item
      }).promise();
      
      console.log(`✅ Migrated quiz: ${row.quiz_name}`);
    }

    // Migrate students
    console.log('👥 Migrating students...');
    const studentResult = await pgClient.query('SELECT * FROM students');
    
    for (const row of studentResult.rows) {
      const item = {
        email: row.email,
        name: row.name,
        phone_number: row.phone_number,
        student_class: row.student_class,
        sub_exp_date: row.sub_exp_date,
        updated_by: row.updated_by,
        amount: row.amount,
        payment_time: row.payment_time,
        role: row.role
      };
      
      await dynamodb.put({
        TableName: 'students',
        Item: item
      }).promise();
      
      console.log(`✅ Migrated student: ${row.email}`);
    }

    // Migrate student_quiz_attempts
    console.log('📊 Migrating quiz attempts...');
    const attemptResult = await pgClient.query('SELECT * FROM student_quiz_attempts');
    
    for (const row of attemptResult.rows) {
      const item = {
        email: row.email,
        quiz_name: row.quiz_name,
        category: row.category,
        correct_count: row.correct_count,
        wrong_count: row.wrong_count,
        skipped_count: row.skipped_count,
        total_count: row.total_count,
        percentage: row.percentage,
        attempt_number: row.attempt_number,
        attempted_at: row.attempted_at.toISOString(),
        results: row.results
      };
      
      await dynamodb.put({
        TableName: 'student_quiz_attempts',
        Item: item
      }).promise();
      
      console.log(`✅ Migrated attempt: ${row.email} - ${row.quiz_name}`);
    }

    console.log('🎉 Migration completed successfully!');
    
  } catch (error) {
    console.error('❌ Migration failed:', error);
  } finally {
    await pgClient.end();
  }
}

migrateData();