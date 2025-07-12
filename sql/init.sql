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
  incorrect_answers VARCHAR(255)[],
  question VARCHAR(255),
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

-- Optional: add index if needed (you already have one on email)
CREATE INDEX IF NOT EXISTS idx_students_email ON students(email);


-- Add your table scripts here