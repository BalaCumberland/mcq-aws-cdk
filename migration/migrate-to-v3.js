const AWS = require('aws-sdk');
const admin = require('firebase-admin');

// Initialize Firebase Admin SDK
admin.initializeApp({
  credential: admin.credential.cert({
    type: "service_account",
    project_id: process.env.FIREBASE_PROJECT_ID,
    private_key_id: process.env.FIREBASE_PRIVATE_KEY_ID,
    private_key: process.env.FIREBASE_PRIVATE_KEY.replace(/\\n/g, '\n'),
    client_email: process.env.FIREBASE_CLIENT_EMAIL,
    client_id: process.env.FIREBASE_CLIENT_ID,
    auth_uri: "https://accounts.google.com/o/oauth2/auth",
    token_uri: "https://oauth2.googleapis.com/token",
    auth_provider_x509_cert_url: "https://www.googleapis.com/oauth2/v1/certs",
    client_x509_cert_url: process.env.FIREBASE_CLIENT_CERT_URL
  })
});

const dynamodb = new AWS.DynamoDB.DocumentClient({ region: 'us-east-1' });

async function getUidFromEmail(email) {
  try {
    const userRecord = await admin.auth().getUserByEmail(email);
    return userRecord.uid;
  } catch (error) {
    console.error(`❌ Error getting UID for ${email}:`, error.message);
    return null;
  }
}

async function migrateStudents() {
  console.log('🔄 Starting student migration...');
  
  // Scan all students from v2 table
  const studentsResult = await dynamodb.scan({
    TableName: 'students'
  }).promise();

  let migrated = 0;
  let failed = 0;

  for (const student of studentsResult.Items) {
    const uid = await getUidFromEmail(student.email);
    
    if (!uid) {
      console.log(`❌ Skipping ${student.email} - no Firebase UID found`);
      failed++;
      continue;
    }

    // Create v3 student record
    const v3Student = {
      uid: uid,
      name: student.name,
      student_class: student.student_class,
      phone_number: student.phone_number,
      sub_exp_date: student.sub_exp_date,
      updated_by: student.updated_by,
      amount: student.amount,
      payment_time: student.payment_time,
      role: student.role
    };

    try {
      await dynamodb.put({
        TableName: 'students_v3',
        Item: v3Student
      }).promise();
      
      console.log(`✅ Migrated student: ${student.email} -> ${uid}`);
      migrated++;
    } catch (error) {
      console.error(`❌ Failed to migrate ${student.email}:`, error.message);
      failed++;
    }
  }

  console.log(`📊 Students migration complete: ${migrated} migrated, ${failed} failed`);
  return { migrated, failed };
}

async function migrateQuizAttempts() {
  console.log('🔄 Starting quiz attempts migration...');
  
  // Scan all attempts from v2 table
  const attemptsResult = await dynamodb.scan({
    TableName: 'student_quiz_attempts'
  }).promise();

  let migrated = 0;
  let failed = 0;

  for (const attempt of attemptsResult.Items) {
    const uid = await getUidFromEmail(attempt.email);
    
    if (!uid) {
      console.log(`❌ Skipping attempt for ${attempt.email} - no Firebase UID found`);
      failed++;
      continue;
    }

    // Create v3 attempt record
    const v3Attempt = {
      uid: uid,
      quiz_name: attempt.quiz_name,
      category: attempt.category,
      correct_count: attempt.correct_count,
      wrong_count: attempt.wrong_count,
      skipped_count: attempt.skipped_count,
      total_count: attempt.total_count,
      percentage: attempt.percentage,
      attempt_number: attempt.attempt_number,
      attempted_at: attempt.attempted_at,
      results: attempt.results
    };

    try {
      await dynamodb.put({
        TableName: 'student_quiz_attempts_v3',
        Item: v3Attempt
      }).promise();
      
      console.log(`✅ Migrated attempt: ${attempt.email}/${attempt.quiz_name} -> ${uid}`);
      migrated++;
    } catch (error) {
      console.error(`❌ Failed to migrate attempt ${attempt.email}/${attempt.quiz_name}:`, error.message);
      failed++;
    }
  }

  console.log(`📊 Quiz attempts migration complete: ${migrated} migrated, ${failed} failed`);
  return { migrated, failed };
}

async function main() {
  try {
    console.log('🚀 Starting migration from v2 to v3...');
    
    const studentStats = await migrateStudents();
    const attemptStats = await migrateQuizAttempts();
    
    console.log('\n📈 Migration Summary:');
    console.log(`Students: ${studentStats.migrated} migrated, ${studentStats.failed} failed`);
    console.log(`Quiz Attempts: ${attemptStats.migrated} migrated, ${attemptStats.failed} failed`);
    console.log('✅ Migration completed!');
    
  } catch (error) {
    console.error('❌ Migration failed:', error);
    process.exit(1);
  }
}

main();