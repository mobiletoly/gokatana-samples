package repo

const selectUserByEmailSql =
/*language=sql*/ `
SELECT
   id, email, password_hash, first_name, last_name, tenant_id, is_active, email_verified, created_at, updated_at
FROM iam.auth_user
WHERE email = @email AND tenant_id = @tenant_id AND is_active = true
LIMIT 1
`

const selectUserByIdSql =
/*language=sql*/ `
SELECT
   id, email, password_hash, first_name, last_name, tenant_id, is_active, email_verified, created_at, updated_at
FROM iam.auth_user
WHERE id = @id AND is_active = true
LIMIT 1
`

const selectUserWithPasswordByEmailSql =
/*language=sql*/ `
SELECT
   id, email, password_hash, first_name, last_name, tenant_id, is_active, email_verified, created_at, updated_at
FROM iam.auth_user
WHERE email = @email AND tenant_id = @tenant_id AND is_active = true
LIMIT 1
`

const selectAllUsersByTenantIdSql =
/*language=sql*/ `
SELECT id, email, password_hash, first_name, last_name, tenant_id, is_active, email_verified, created_at, updated_at
FROM iam.auth_user
WHERE tenant_id = @tenant_id
ORDER BY created_at DESC
`

const selectAllUsersSql =
/*language=sql*/ `
SELECT id, email, password_hash, first_name, last_name, tenant_id, is_active, email_verified, created_at, updated_at
FROM iam.auth_user
ORDER BY created_at DESC
`

const insertUserSql =
/*language=sql*/ `
INSERT INTO iam.auth_user (id, email, password_hash, first_name, last_name, tenant_id, is_active, email_verified, created_at, updated_at)
VALUES (@id, @email, @password_hash, @first_name, @last_name, @tenant_id, @is_active, @email_verified, @created_at, @updated_at)
`

const deleteUserSql =
/*language=sql*/ `
DELETE FROM iam.auth_user WHERE id = @id
`

const selectUserRolesSql =
/*language=sql*/ `
SELECT r.id, r.name, r.description
FROM iam.auth_role r
JOIN iam.auth_user_role ur ON r.id = ur.role_id
WHERE ur.user_id = @user_id
ORDER BY r.name
`

const insertUserRoleSql =
/*language=sql*/ `
INSERT INTO iam.auth_user_role (user_id, role_id)
VALUES (@user_id, @role_id)
`

const deleteUserRoleSql =
/*language=sql*/ `
DELETE FROM iam.auth_user_role
WHERE user_id = @user_id AND role_id = @role_id
`

const selectRoleByNameSql =
/*language=sql*/ `
SELECT id, name, description
FROM iam.auth_role
WHERE name = @name
LIMIT 1
`

const selectTenantByIdSql =
/*language=sql*/ `
SELECT id, name, description, created_at, updated_at
FROM iam.tenant
WHERE id = @id
LIMIT 1
`

const selectAllTenantsSql =
/*language=sql*/ `
SELECT id, name, description, created_at, updated_at
FROM iam.tenant
ORDER BY created_at DESC
`

const insertTenantSql =
/*language=sql*/ `
INSERT INTO iam.tenant (id, name, description, created_at, updated_at)
VALUES (@id, @name, @description, @created_at, @updated_at)
`

const updateTenantSql =
/*language=sql*/ `
UPDATE iam.tenant
SET name = @name, description = @description, updated_at = @updated_at
WHERE id = @id
`

const deleteTenantSql =
/*language=sql*/ `
DELETE FROM iam.tenant
WHERE id = @id
`

const selectUserProfileByUserIdSql =
/*language=sql*/ `
SELECT id, user_id, height, weight, gender, birth_date, is_metric, created_at, updated_at
FROM iam.user_profile
WHERE user_id = @user_id
LIMIT 1
`

const insertUserProfileSql =
/*language=sql*/ `
INSERT INTO iam.user_profile (user_id, is_metric, created_at, updated_at)
VALUES (@user_id, @is_metric, @created_at, @updated_at)
RETURNING id, user_id, height, weight, gender, birth_date, is_metric, created_at, updated_at
`

// Refresh token SQL queries
const insertRefreshTokenSql =
/*language=sql*/ `
INSERT INTO iam.auth_refresh_token (id, user_id, token_hash, issued_at, expires_at, revoked)
VALUES (@id, @user_id, @token_hash, @issued_at, @expires_at, @revoked)
`

const selectRefreshTokenByHashSql =
/*language=sql*/ `
SELECT id, user_id, token_hash, issued_at, expires_at, revoked
FROM iam.auth_refresh_token
WHERE token_hash = @token_hash AND revoked = false AND expires_at > now()
LIMIT 1
`

const revokeRefreshTokenSql =
/*language=sql*/ `
UPDATE iam.auth_refresh_token
SET revoked = true
WHERE token_hash = @token_hash
`

const revokeAllUserRefreshTokensSql =
/*language=sql*/ `
UPDATE iam.auth_refresh_token
SET revoked = true
WHERE user_id = @user_id AND revoked = false
`

const deleteExpiredRefreshTokensSql =
/*language=sql*/ `
DELETE FROM iam.auth_refresh_token
WHERE expires_at < now() OR revoked = true
`

const cleanupUserRefreshTokensSql =
/*language=sql*/ `
WITH tokens_to_keep AS (
    -- Keep the 2 most recent non-revoked tokens (excluding the new one being created)
    SELECT id
    FROM iam.auth_refresh_token
    WHERE user_id = @user_id AND revoked = false
    ORDER BY issued_at DESC
    LIMIT 2
)
DELETE FROM iam.auth_refresh_token
WHERE user_id = @user_id
  AND (revoked = true OR id NOT IN (SELECT id FROM tokens_to_keep))
`

// Email confirmation token SQL queries
const insertEmailConfirmationTokenSql =
/*language=sql*/ `
INSERT INTO iam.email_confirmation_token (id, user_id, email, token_hash, source, expires_at, created_at)
VALUES (@id, @user_id, @email, @token_hash, @source, @expires_at, @created_at)
ON CONFLICT (user_id) DO UPDATE SET
	email = EXCLUDED.email,
	token_hash = EXCLUDED.token_hash,
	source = EXCLUDED.source,
	expires_at = EXCLUDED.expires_at,
	created_at = EXCLUDED.created_at,
	used_at = NULL
`

const selectEmailConfirmationTokenByUserIdAndHashSql =
/*language=sql*/ `
SELECT id, user_id, email, token_hash, source, expires_at, used_at, created_at
FROM iam.email_confirmation_token
WHERE user_id = @user_id AND token_hash = @token_hash
`

const markEmailConfirmationTokenAsUsedSql =
/*language=sql*/ `
UPDATE iam.email_confirmation_token
SET used_at = now()
WHERE id = @token_id
`

const setUserEmailVerifiedSql =
/*language=sql*/ `
UPDATE iam.auth_user
SET email_verified = @verified, updated_at = now()
WHERE id = @user_id
`

const updateUserProfileSql =
/*language=sql*/ `
UPDATE iam.user_profile
SET height = @height, weight = @weight, gender = @gender, birth_date = @birth_date, is_metric = @is_metric, updated_at = @updated_at
WHERE user_id = @user_id
RETURNING id, user_id, height, weight, gender, birth_date, is_metric, created_at, updated_at
`
