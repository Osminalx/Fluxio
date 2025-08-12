-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create database if it doesn't exist (this will be handled by docker-compose)
-- The database 'fluxio' will be created automatically

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE fluxio TO postgres;

-- Connect to the fluxio database
\c fluxio;

-- Enable UUID extension in the fluxio database
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create a simple test table to verify the setup
CREATE TABLE IF NOT EXISTS health_check (
    id SERIAL PRIMARY KEY,
    message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert a test record
INSERT INTO health_check (message) VALUES ('Database initialized successfully');

-- Grant all privileges on all tables to postgres
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO postgres;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO postgres;
