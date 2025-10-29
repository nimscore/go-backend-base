CREATE TABLE IF NOT EXISTS "users" (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    salt TEXT NOT NULL,
    verification_token TEXT NOT NULL,
    reset_token TEXT NOT NULL,
    is_verified BOOLEAN,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "sessions" (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users,
    user_agent TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "communities" (
    id UUID PRIMARY KEY,
    owner_id UUID REFERENCES users,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
