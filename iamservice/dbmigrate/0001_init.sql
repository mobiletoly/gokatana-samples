CREATE EXTENSION IF NOT EXISTS citext;

DROP TABLE IF EXISTS auth_user CASCADE;
DROP TABLE IF EXISTS auth_user_identity CASCADE;
DROP TABLE IF EXISTS auth_role CASCADE;
DROP TABLE IF EXISTS auth_user_role CASCADE;
DROP TABLE IF EXISTS contact CASCADE;

-- Holds application’s user profiles
CREATE TABLE auth_user
(
    id             TEXT PRIMARY KEY,
    email          CITEXT UNIQUE NOT NULL,
    password_hash  TEXT          NOT NULL,
    first_name     TEXT          NOT NULL,
    last_name      TEXT          NOT NULL,
    is_active      BOOLEAN       NOT NULL DEFAULT TRUE,
    email_verified BOOLEAN       NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT now()
);

-- Link a user to one or more auth methods
CREATE TABLE auth_user_identity
(
    id               TEXT PRIMARY KEY,
    user_id          TEXT        NOT NULL REFERENCES auth_user (id) ON DELETE CASCADE,
    provider         TEXT        NOT NULL, -- e.g. 'local', 'google', 'partner_xyz'
    provider_user_id TEXT        NOT NULL, -- e.g. OAuth sub, partner’s user ID
    access_token     TEXT,                 -- optional, for OAuth flows
    refresh_token    TEXT,                 -- optional, for OAuth refresh
    token_expires_at TIMESTAMPTZ,          -- when access_token expires
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider, provider_user_id)
);

-- Role-Based Access Control (RBAC) tables
CREATE TABLE auth_role
(
    id          SERIAL PRIMARY KEY,
    name        TEXT        NOT NULL UNIQUE, -- e.g. 'admin', 'user', 'moderator'
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE auth_user_role
(
    user_id     TEXT        NOT NULL REFERENCES auth_user (id) ON DELETE CASCADE,
    role_id     INT         NOT NULL REFERENCES auth_role (id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    assigned_by TEXT, -- ID of user who assigned this role (optional)
    PRIMARY KEY (user_id, role_id)
);

-- Insert default roles
INSERT INTO auth_role (name, description)
VALUES ('user', 'Standard user with basic permissions'),
       ('admin', 'Administrator with full system access'),
       ('moderator', 'Moderator with limited administrative permissions');

-- Create indexes for better performance
CREATE INDEX idx_auth_user_role_user_id ON auth_user_role (user_id);
CREATE INDEX idx_auth_user_role_role_id ON auth_user_role (role_id);
-- For auth_user_identity lookups & deletes
CREATE INDEX idx_auth_user_identity_user_id ON auth_user_identity (user_id);
CREATE INDEX idx_auth_user_identity_token_expires_at
    ON auth_user_identity (token_expires_at);
