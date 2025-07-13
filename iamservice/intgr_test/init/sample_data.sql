-- Sample auth users for testing (with hashed password for 'qazwsxedc')
-- Hash generated with: bcrypt.GenerateFromPassword([]byte("qazwsxedc"), bcrypt.DefaultCost)
INSERT INTO iam.auth_user (id, email, password_hash, first_name, last_name, tenant_id, is_active, email_verified, created_at, updated_at) VALUES
    ('test-user-5', 'testuser@example.com', '$2a$10$Nk.Isu283VbMJatqaon/CuQrIxvcnaGCsFBjv4jUmoQGGrUpsr/sa', 'Test', 'User', 'default-tenant', true, true, now(), now()),
    ('test-admin-5', 'testadmin@example.com', '$2a$10$Nk.Isu283VbMJatqaon/CuQrIxvcnaGCsFBjv4jUmoQGGrUpsr/sa', 'Test', 'Admin', 'default-tenant', true, true, now(), now()),
    ('test-admin-6', 'testuser_different_tenant@example.com', '$2a$10$Nk.Isu283VbMJatqaon/CuQrIxvcnaGCsFBjv4jUmoQGGrUpsr/sa', 'Test', 'User', 'test-tenant', true, true, now(), now());

-- Assign roles to sample users
INSERT INTO iam.auth_user_role (user_id, role_id) VALUES
    ('test-user-5', 1),   -- user role
    ('test-admin-5', 1),  -- user role
    ('test-admin-5', 2);  -- admin role

-- Create user profiles for sample users
INSERT INTO iam.user_profile (user_id, height, weight, gender, birth_date, created_at, updated_at) VALUES
    ('test-user-5', 175, 70, 'male', '1990-01-15', now(), now()),
    ('test-admin-5', 180, 75, 'female', '1985-05-20', now(), now());
