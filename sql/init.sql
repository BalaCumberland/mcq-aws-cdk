-- Database initialization script
-- Run this after connecting to your PostgreSQL database

-- Example table structure (replace with your actual tables)
CREATE TABLE IF NOT EXISTS quiz_questions (
  id BIGSERIAL PRIMARY KEY,
  quiz_name VARCHAR(255) NOT NULL UNIQUE,
  duration INTEGER NOT NULL,
  category VARCHAR(255) NOT NULL,
  questions JSONB NOT NULL,
  correct_answer VARCHAR(255),
  explanation VARCHAR(255),
  all_answers VARCHAR(255)[],
  question TEXT,
  uploaded_time TIMESTAMP
);


CREATE TABLE IF NOT EXISTS students (
  id SERIAL PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  phone_number VARCHAR(20),
  name VARCHAR(255),
  student_class VARCHAR(50) NOT NULL DEFAULT 'DEMO',
  created_time TIMESTAMP NOT NULL DEFAULT now(),
  updated_time TIMESTAMP NOT NULL DEFAULT now(),
  payment_time TIMESTAMP,
  updated_by TEXT,
  sub_exp_date DATE,
  amount NUMERIC,
  last_upgrade_time TIMESTAMP,
  role TEXT
);

CREATE TABLE IF NOT EXISTS student_quizzes (
  email VARCHAR(255) NOT NULL PRIMARY KEY,
  quiz_names JSONB NOT NULL DEFAULT '[]'::jsonb,
  created_time TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT student_quizzes_email_fkey
    FOREIGN KEY (email)
    REFERENCES students(email)
    ON DELETE CASCADE
);


-- Student progress tracking tables

CREATE TABLE IF NOT EXISTS student_quiz_attempts (
  id BIGSERIAL PRIMARY KEY,
  email VARCHAR(255) NOT NULL,
  quiz_name VARCHAR(255) NOT NULL,
  category VARCHAR(255) NOT NULL,
  correct_count INTEGER NOT NULL,
  wrong_count INTEGER NOT NULL,
  skipped_count INTEGER NOT NULL,
  total_count INTEGER NOT NULL,
  percentage DECIMAL(5,2) NOT NULL,
  attempt_number INTEGER NOT NULL DEFAULT 1,
  attempted_at TIMESTAMP NOT NULL DEFAULT now(),
  results JSONB,
  CONSTRAINT fk_student_attempts_email 
    FOREIGN KEY (email) REFERENCES students(email) ON DELETE CASCADE,
  UNIQUE(email, quiz_name)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_student_attempts_email ON student_quiz_attempts(email);
CREATE INDEX IF NOT EXISTS idx_student_attempts_category ON student_quiz_attempts(category);
CREATE INDEX IF NOT EXISTS idx_student_attempts_quiz_name ON student_quiz_attempts(quiz_name);
CREATE INDEX IF NOT EXISTS idx_student_attempts_email_category ON student_quiz_attempts(email, category);
CREATE INDEX IF NOT EXISTS idx_student_attempts_attempted_at ON student_quiz_attempts(attempted_at DESC);

-- Optional: add index if needed (you already have one on email)
CREATE INDEX IF NOT EXISTS idx_students_email ON students(email);


-- Add your table scripts here