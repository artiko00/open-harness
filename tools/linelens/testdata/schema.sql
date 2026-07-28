-- Esquema de base de datos de ejemplo
CREATE TABLE users (
  id INTEGER PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  created_at TIMESTAMP DEFAULT now()
);
CREATE INDEX idx_users_email ON users(email);
