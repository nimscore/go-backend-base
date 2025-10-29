CREATE TABLE IF NOT EXISTS "users" (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    salt TEXT NOT NULL,
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
    slug TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
