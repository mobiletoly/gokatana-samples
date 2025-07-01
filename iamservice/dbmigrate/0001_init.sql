CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE tenant
(
    id          TEXT PRIMARY KEY,
    name        TEXT        NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE auth_user
(
    id             TEXT PRIMARY KEY,
    email          CITEXT      NOT NULL,
    password_hash  TEXT        NOT NULL,
    first_name     TEXT        NOT NULL,
    last_name      TEXT        NOT NULL,
    tenant_id      TEXT        NOT NULL REFERENCES tenant (id) ON DELETE CASCADE,
    is_active      BOOLEAN     NOT NULL DEFAULT TRUE,
    email_verified BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (email, tenant_id) -- Email must be unique within a tenant
);

-- Create indexes for better performance
CREATE INDEX idx_auth_user_tenant_id ON auth_user (tenant_id);
CREATE INDEX idx_auth_user_email_tenant ON auth_user (email, tenant_id);

CREATE TABLE auth_role
(
    id          SERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE, -- e.g. 'sysadmin', 'admin', 'user'
    description TEXT
);

CREATE TABLE auth_user_role
(
    user_id TEXT NOT NULL REFERENCES auth_user (id) ON DELETE CASCADE,
    role_id INT  NOT NULL REFERENCES auth_role (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- Create indexes for auth_user_role table
CREATE INDEX idx_auth_user_role_user_id ON auth_user_role (user_id);
CREATE INDEX idx_auth_user_role_role_id ON auth_user_role (role_id);

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

-- Email confirmation tokens table
CREATE TABLE email_confirmation_token
(
    id         TEXT PRIMARY KEY,
    user_id    TEXT        NOT NULL REFERENCES auth_user (id) ON DELETE CASCADE,
    email      CITEXT      NOT NULL,
    token_hash TEXT        NOT NULL,
    source     TEXT        NOT NULL CHECK (source IN ('web', 'android', 'ios')),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id),            -- Only one active token per user
    UNIQUE (user_id, token_hash) -- Unique combination of user_id and token hash
);

-- Create indexes for email confirmation tokens
CREATE INDEX idx_email_confirmation_token_user_id_hash ON email_confirmation_token (user_id, token_hash);
CREATE INDEX idx_email_confirmation_token_user_id ON email_confirmation_token (user_id);
CREATE INDEX idx_email_confirmation_token_expires_at ON email_confirmation_token (expires_at);

CREATE TABLE user_profile
(
    id         SERIAL PRIMARY KEY,
    user_id    TEXT        NOT NULL REFERENCES auth_user (id) ON DELETE CASCADE,
    height     INT         NULL,
    weight     INT         NULL,
    gender     TEXT        NULL CHECK (gender IN ('male', 'female', 'other')) DEFAULT 'other',
    birth_date DATE        NULL,
    is_metric  BOOLEAN     NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_user_profile_user_id ON user_profile (user_id);
CREATE INDEX idx_user_profile_birth_date ON user_profile (birth_date);

-- Insert default tenant
INSERT INTO tenant (id, name, description)
VALUES ('default-tenant', 'Default Tenant', 'Default tenant for the application');

INSERT INTO tenant (id, name, description)
VALUES ('test-tenant', 'Test Tenant', 'Test tenant for the application');

-- Insert system-wide roles
INSERT INTO auth_role (name, description)
VALUES ('user', 'Standard user with basic permissions within their tenant'),
       ('admin', 'Administrator with full access within their tenant'),
       ('sysadmin', 'System administrator with access across all tenants');

-- Create admin user for default tenant
INSERT INTO auth_user (id, email, password_hash, first_name, last_name, tenant_id, is_active, email_verified,
                       created_at, updated_at)
VALUES ('default-admin-1', 'john.doe.admin@example.com', '$2a$10$Nk.Isu283VbMJatqaon/CuQrIxvcnaGCsFBjv4jUmoQGGrUpsr/sa',
        'Joe-Admin', 'Doe', 'default-tenant', true, true, now(), now());
INSERT INTO auth_user_role (user_id, role_id)
VALUES ('default-admin-1', 2);
INSERT INTO user_profile (user_id)
VALUES ('default-admin-1');

-- Create sample users for default tenant
INSERT INTO auth_user (id, email, password_hash, first_name, last_name, tenant_id, is_active, email_verified,
                       created_at, updated_at)
VALUES ('default-user-1', 'john.doe.default.user1@example.com',
        '$2a$10$Nk.Isu283VbMJatqaon/CuQrIxvcnaGCsFBjv4jUmoQGGrUpsr/sa', 'Joe-User1', 'Doe', 'default-tenant', true,
        true,
        now(), now()),
       ('default-user-2', 'john.doe.default.user2@example.com',
        '$2a$10$Nk.Isu283VbMJatqaon/CuQrIxvcnaGCsFBjv4jUmoQGGrUpsr/sa', 'Joe-User2', 'Doe', 'default-tenant', true,
        true,
        now(), now());
INSERT INTO auth_user_role (user_id, role_id)
VALUES ('default-user-1', 1),
       ('default-user-2', 1);
INSERT INTO user_profile (user_id, birth_date)
VALUES ('default-user-1', '1990-05-23'),
       ('default-user-2', NULL);

-- Create admin user for test tenant
INSERT INTO auth_user (id, email, password_hash, first_name, last_name, tenant_id, is_active, email_verified,
                       created_at, updated_at)
VALUES ('test-admin-1', 'john.doe.admin@example.com', '$2a$10$Nk.Isu283VbMJatqaon/CuQrIxvcnaGCsFBjv4jUmoQGGrUpsr/sa',
        'Joe-Admin', 'Doe', 'test-tenant', true, true, now(), now());
INSERT INTO auth_user_role (user_id, role_id)
VALUES ('test-admin-1', 2);
INSERT INTO user_profile (user_id)
VALUES ('test-admin-1');

-- Create sample users for test tenant
INSERT INTO auth_user (id, email, password_hash, first_name, last_name, tenant_id, is_active, email_verified,
                       created_at, updated_at)
VALUES ('test-user-1', 'john.doe.test.user1@example.com',
        '$2a$10$Nk.Isu283VbMJatqaon/CuQrIxvcnaGCsFBjv4jUmoQGGrUpsr/sa', 'Joe-TestUser1', 'Doe', 'test-tenant', true,
        true,
        now(), now()),
       ('test-user-2', 'john.doe.test.user2@example.com',
        '$2a$10$Nk.Isu283VbMJatqaon/CuQrIxvcnaGCsFBjv4jUmoQGGrUpsr/sa', 'Joe-TestUser2', 'Doe', 'test-tenant', true,
        true,
        now(), now());
INSERT INTO auth_user_role (user_id, role_id)
VALUES ('test-user-1', 1),
       ('test-user-2', 1);
INSERT INTO user_profile (user_id)
VALUES ('test-user-1'),
       ('test-user-2');

-- Create system admin
INSERT INTO auth_user (id, email, password_hash, first_name, last_name, tenant_id, is_active, email_verified,
                       created_at, updated_at)
VALUES ('sys-admin-1', 'john.doe.sysadmin@example.com', '$2a$10$Nk.Isu283VbMJatqaon/CuQrIxvcnaGCsFBjv4jUmoQGGrUpsr/sa',
        'Joe-SysAdmin', 'Doe', 'default-tenant', true, true, now(), now());
INSERT INTO auth_user_role (user_id, role_id)
VALUES ('sys-admin-1', 3);
INSERT INTO user_profile (user_id)
VALUES ('sys-admin-1');
