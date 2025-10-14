CREATE TABLE IF NOT EXISTS "users" (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS "companies" (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    user_id UUID REFERENCES users
);

