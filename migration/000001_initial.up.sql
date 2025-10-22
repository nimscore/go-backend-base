CREATE TABLE IF NOT EXISTS "users" (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    salt TEXT NOT NULL,
    is_verified BOOLEAN,
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
