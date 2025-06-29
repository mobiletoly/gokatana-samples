package repo

const selectUserByEmailSql =
/*language=sql*/ `
SELECT
   id, email, password_hash, first_name, last_name, tenant_id, is_active, email_verified, created_at, updated_at
FROM auth_user
WHERE email = @email AND tenant_id = @tenant_id AND is_active = true
LIMIT 1
`

const selectUserByIdSql =
/*language=sql*/ `
SELECT
   id, email, password_hash, first_name, last_name, tenant_id, is_active, email_verified, created_at, updated_at
FROM auth_user
WHERE id = @id AND is_active = true
LIMIT 1
`

const selectUserWithPasswordByEmailSql =
/*language=sql*/ `
SELECT
   id, email, password_hash, first_name, last_name, tenant_id, is_active, email_verified, created_at, updated_at
FROM auth_user
WHERE email = @email AND tenant_id = @tenant_id AND is_active = true
LIMIT 1
`

const selectAllUsersByTenantIdSql =
/*language=sql*/ `
SELECT id, email, password_hash, first_name, last_name, tenant_id, is_active, email_verified, created_at, updated_at
FROM auth_user
WHERE tenant_id = @tenant_id
ORDER BY created_at DESC
`

const selectAllUsersSql =
/*language=sql*/ `
SELECT id, email, password_hash, first_name, last_name, tenant_id, is_active, email_verified, created_at, updated_at
FROM auth_user
ORDER BY created_at DESC
`

const insertUserSql =
/*language=sql*/ `
INSERT INTO auth_user (id, email, password_hash, first_name, last_name, tenant_id, is_active, email_verified, created_at, updated_at)
VALUES (@id, @email, @password_hash, @first_name, @last_name, @tenant_id, @is_active, @email_verified, @created_at, @updated_at)
`

const deleteUserSql =
/*language=sql*/ `
DELETE FROM auth_user WHERE id = @id
`

const selectUserRolesSql =
/*language=sql*/ `
SELECT r.id, r.name, r.description
FROM auth_role r
JOIN auth_user_role ur ON r.id = ur.role_id
WHERE ur.user_id = @user_id
ORDER BY r.name
`

const insertUserRoleSql =
/*language=sql*/ `
INSERT INTO auth_user_role (user_id, role_id)
VALUES (@user_id, @role_id)
`

const deleteUserRoleSql =
/*language=sql*/ `
DELETE FROM auth_user_role
WHERE user_id = @user_id AND role_id = @role_id
`

const selectRoleByNameSql =
/*language=sql*/ `
SELECT id, name, description
FROM auth_role
WHERE name = @name
LIMIT 1
`

const selectTenantByIdSql =
/*language=sql*/ `
SELECT id, name, description, created_at, updated_at
FROM tenant
WHERE id = @id
LIMIT 1
`

const selectAllTenantsSql =
/*language=sql*/ `
SELECT id, name, description, created_at, updated_at
FROM tenant
ORDER BY created_at DESC
`

const insertTenantSql =
/*language=sql*/ `
INSERT INTO tenant (id, name, description, created_at, updated_at)
VALUES (@id, @name, @description, @created_at, @updated_at)
`

const updateTenantSql =
/*language=sql*/ `
UPDATE tenant
SET name = @name, description = @description, updated_at = @updated_at
WHERE id = @id
`

const deleteTenantSql =
/*language=sql*/ `
DELETE FROM tenant
WHERE id = @id
`

const selectUserProfileByUserIdSql =
/*language=sql*/ `
SELECT id, user_id, height, weight, gender, birth_date, is_metric, created_at, updated_at
FROM user_profile
WHERE user_id = @user_id
LIMIT 1
`

const insertUserProfileSql =
/*language=sql*/ `
INSERT INTO user_profile (user_id, is_metric, created_at, updated_at)
VALUES (@user_id, @is_metric, @created_at, @updated_at)
RETURNING id, user_id, height, weight, gender, birth_date, is_metric, created_at, updated_at
`
