-- schema.sql → helps sqlc understand DB(current database blueprint)
-- migrations → actually create DB
-- queries → define operations

-- Migration → Schema → Generate
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE tasks (

    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    type TEXT NOT NULL,

    status TEXT NOT NULL,

    payload JSONB NOT NULL DEFAULT '{}'::jsonb,

    result JSONB,

    error_message TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tasks_status ON tasks(status);
