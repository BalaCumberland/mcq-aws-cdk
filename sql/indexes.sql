-- Database indexes for performance optimization

-- Index on quiz_questions table
CREATE INDEX IF NOT EXISTS idx_quiz_questions_quiz_name ON quiz_questions(quiz_name);
CREATE INDEX IF NOT EXISTS idx_quiz_questions_category ON quiz_questions(category);

-- Index on students table  
CREATE INDEX IF NOT EXISTS idx_students_email ON students(email);
CREATE INDEX IF NOT EXISTS idx_students_class ON students(student_class);
CREATE INDEX IF NOT EXISTS idx_students_role ON students(role);

-- Index on student_quizzes table
CREATE INDEX IF NOT EXISTS idx_student_quizzes_email ON student_quizzes(email);

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_quiz_questions_category_name ON quiz_questions(category, quiz_name);
CREATE INDEX IF NOT EXISTS idx_students_email_class ON students(email, student_class);