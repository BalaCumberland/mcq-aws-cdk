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