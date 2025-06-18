package repo

const selectUserByEmailSql =
/*language=sql*/ `
SELECT
   id, email, password_hash, first_name, last_name, is_active, email_verified, created_at, updated_at
FROM auth_user
WHERE email = @email AND is_active = true
LIMIT 1
`

const selectUserByIdSql =
/*language=sql*/ `
SELECT
   id, email, password_hash, first_name, last_name, is_active, email_verified, created_at, updated_at
FROM auth_user
WHERE id = @id AND is_active = true
LIMIT 1
`

const selectUserWithPasswordByEmailSql =
/*language=sql*/ `
SELECT
   id, email, password_hash, first_name, last_name, is_active, email_verified, created_at, updated_at
FROM auth_user
WHERE email = @email AND is_active = true
LIMIT 1
`

const selectAllUsersSql =
/*language=sql*/ `
SELECT id, email, password_hash, first_name, last_name, is_active, email_verified, created_at, updated_at
FROM auth_user
ORDER BY created_at DESC
`

const insertUserSql =
/*language=sql*/ `
INSERT INTO auth_user (id, email, password_hash, first_name, last_name, is_active, email_verified, created_at, updated_at)
VALUES (@id, @email, @password_hash, @first_name, @last_name, @is_active, @email_verified, @created_at, @updated_at)
`

const deleteUserSql =
/*language=sql*/ `
DELETE FROM auth_user WHERE id = @id
`

const selectUserRolesSql =
/*language=sql*/ `
SELECT r.id, r.name, r.description, r.created_at, r.updated_at
FROM auth_role r
JOIN auth_user_role ur ON r.id = ur.role_id
WHERE ur.user_id = @user_id
ORDER BY r.name
`

const insertUserRoleSql =
/*language=sql*/ `
INSERT INTO auth_user_role (user_id, role_id, assigned_by)
VALUES (@user_id, @role_id, @assigned_by)
`

const deleteUserRoleSql =
/*language=sql*/ `
DELETE FROM auth_user_role
WHERE user_id = @user_id AND role_id = @role_id
`

const selectRoleByNameSql =
/*language=sql*/ `
SELECT id, name, description, created_at, updated_at
FROM auth_role
WHERE name = @name
LIMIT 1
`
