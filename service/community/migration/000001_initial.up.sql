CREATE TABLE IF NOT EXISTS "users" (
    id UUID PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    salt TEXT NOT NULL,
    is_verified BOOLEAN,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "companies" (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    user_id UUID REFERENCES users
);

