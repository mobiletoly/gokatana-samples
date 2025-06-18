-- Sample auth users for testing (with hashed password for 'password123')
-- Hash generated with: bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
INSERT INTO auth_user (id, email, password_hash, first_name, last_name, is_active, email_verified, created_at, updated_at) VALUES
    ('test-user-1', 'testuser@example.com', '$2a$10$EOkK7.1xiPEwo1kzhZoIWO3kqXk7ZbAuWCPjxUwj.ju5MPnn3wGju', 'Test', 'User', true, true, now(), now()),
    ('test-admin-1', 'testadmin@example.com', '$2a$10$EOkK7.1xiPEwo1kzhZoIWO3kqXk7ZbAuWCPjxUwj.ju5MPnn3wGju', 'Test', 'Admin', true, true, now(), now());

-- Assign roles to sample users
INSERT INTO auth_user_role (user_id, role_id) VALUES
    ('test-user-1', 1),   -- user role
    ('test-admin-1', 1),  -- user role
    ('test-admin-1', 2);  -- admin role
