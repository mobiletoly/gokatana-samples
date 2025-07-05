package repo

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/model"
	"github.com/mobiletoly/gokatana/katpg"
)

//go:generate go tool gobetter -input $GOFILE

type AuthUserEntity struct { //+gob:Constructor
	ID            *string   `db:"id"`
	Email         string    `db:"email"`
	PasswordHash  string    `db:"password_hash"`
	FirstName     string    `db:"first_name"`
	LastName      string    `db:"last_name"`
	TenantID      string    `db:"tenant_id"`
	IsActive      bool      `db:"is_active"`
	EmailVerified bool      `db:"email_verified"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type AuthRoleEntity struct { //+gob:Constructor
	ID          *int    `db:"id"`
	Name        string  `db:"name"`
	Description *string `db:"description"`
}

type AuthUserRoleEntity struct { //+gob:Constructor
	UserID string `db:"user_id"`
	RoleID int    `db:"role_id"`
}

type TenantEntity struct { //+gob:Constructor
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type UserProfileEntity struct { //+gob:Constructor
	ID        *int       `db:"id"`
	UserID    string     `db:"user_id"`
	Height    *int       `db:"height"`
	Weight    *int       `db:"weight"`
	Gender    *string    `db:"gender"`
	BirthDate *time.Time `db:"birth_date"`
	IsMetric  bool       `db:"is_metric"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}

func SelectUserByEmail(ctx context.Context, tx pgx.Tx, email string, tenantID string) (*AuthUserEntity, error) {
	rows, _ := tx.Query(ctx, selectUserByEmailSql, pgx.NamedArgs{"email": email, "tenant_id": tenantID})
	ent, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AuthUserEntity])
	if katpg.IsNoRows(err) {
		return nil, nil
	}
	return &ent, err
}

func SelectUserByID(ctx context.Context, tx pgx.Tx, userID string) (*AuthUserEntity, error) {
	rows, _ := tx.Query(ctx, selectUserByIdSql, pgx.NamedArgs{"id": userID})
	ent, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AuthUserEntity])
	if katpg.IsNoRows(err) {
		return nil, nil
	}
	return &ent, err
}

func SelectUserWithPasswordByEmail(ctx context.Context, tx pgx.Tx, email string, tenantID string) (*AuthUserEntity, error) {
	rows, _ := tx.Query(ctx, selectUserWithPasswordByEmailSql, pgx.NamedArgs{"email": email, "tenant_id": tenantID})
	ent, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AuthUserEntity])
	if katpg.IsNoRows(err) {
		return nil, nil
	}
	return &ent, err
}

func InsertUser(ctx context.Context, tx pgx.Tx, user *AuthUserEntity) error {
	_, err := tx.Exec(ctx, insertUserSql, pgx.NamedArgs{
		"id":             *user.ID,
		"email":          user.Email,
		"password_hash":  user.PasswordHash,
		"first_name":     user.FirstName,
		"last_name":      user.LastName,
		"tenant_id":      user.TenantID,
		"is_active":      user.IsActive,
		"email_verified": user.EmailVerified,
		"created_at":     user.CreatedAt,
		"updated_at":     user.UpdatedAt,
	})
	return err
}

func UpdateUser(ctx context.Context, tx pgx.Tx, userID string, updates map[string]interface{}) error {
	// Build dynamic UPDATE query based on the updates map
	setParts := make([]string, 0, len(updates))
	args := pgx.NamedArgs{"id": userID}

	for field, value := range updates {
		setParts = append(setParts, field+" = @"+field)
		args[field] = value
	}

	if len(setParts) == 0 {
		return nil // No updates to perform
	}

	query := "UPDATE auth_user SET " + strings.Join(setParts, ", ") + " WHERE id = @id"
	_, err := tx.Exec(ctx, query, args)
	return err
}

// Email confirmation token methods

func InsertEmailConfirmationToken(ctx context.Context, tx pgx.Tx, token *model.EmailConfirmationToken) error {
	query := `
		INSERT INTO email_confirmation_token (id, user_id, email, token_hash, source, expires_at, created_at)
		VALUES (@id, @user_id, @email, @token_hash, @source, @expires_at, @created_at)
		ON CONFLICT (user_id) DO UPDATE SET
			email = EXCLUDED.email,
			token_hash = EXCLUDED.token_hash,
			source = EXCLUDED.source,
			expires_at = EXCLUDED.expires_at,
			created_at = EXCLUDED.created_at,
			used_at = NULL
	`
	args := pgx.NamedArgs{
		"id":         token.ID,
		"user_id":    token.UserID,
		"email":      token.Email,
		"token_hash": token.TokenHash,
		"source":     token.Source,
		"expires_at": token.ExpiresAt,
		"created_at": token.CreatedAt,
	}
	_, err := tx.Exec(ctx, query, args)
	return err
}

func GetEmailConfirmationTokenByUserIDAndHash(ctx context.Context, tx pgx.Tx, userID string, tokenHash string) (*model.EmailConfirmationToken, error) {
	query := `
		SELECT id, user_id, email, token_hash, source, expires_at, used_at, created_at
		FROM email_confirmation_token
		WHERE user_id = @user_id AND token_hash = @token_hash
	`
	args := pgx.NamedArgs{
		"user_id":    userID,
		"token_hash": tokenHash,
	}

	var confirmationToken model.EmailConfirmationToken
	err := tx.QueryRow(ctx, query, args).Scan(
		&confirmationToken.ID,
		&confirmationToken.UserID,
		&confirmationToken.Email,
		&confirmationToken.TokenHash,
		&confirmationToken.Source,
		&confirmationToken.ExpiresAt,
		&confirmationToken.UsedAt,
		&confirmationToken.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
			return nil, nil
		}
		return nil, err
	}

	return &confirmationToken, nil
}

func MarkEmailConfirmationTokenAsUsed(ctx context.Context, tx pgx.Tx, tokenID string) error {
	query := `
		UPDATE email_confirmation_token
		SET used_at = now()
		WHERE id = @token_id
	`
	args := pgx.NamedArgs{"token_id": tokenID}
	_, err := tx.Exec(ctx, query, args)
	return err
}

func SetUserEmailVerified(ctx context.Context, tx pgx.Tx, userID string, verified bool) error {
	query := `
		UPDATE auth_user
		SET email_verified = @verified, updated_at = now()
		WHERE id = @user_id
	`
	args := pgx.NamedArgs{
		"verified": verified,
		"user_id":  userID,
	}
	_, err := tx.Exec(ctx, query, args)
	return err
}

// DeleteUser deletes a user from the system, returning the number of rows deleted
func DeleteUser(ctx context.Context, tx pgx.Tx, userID string) (int64, error) {
	cmd, err := tx.Exec(ctx, deleteUserSql, pgx.NamedArgs{"id": userID})
	if err != nil {
		return 0, err
	}
	rowsAffected := cmd.RowsAffected()
	return rowsAffected, err
}

func SelectAllUsersByTenantId(ctx context.Context, tx pgx.Tx, tenantID string) ([]AuthUserEntity, error) {
	rows, _ := tx.Query(ctx, selectAllUsersByTenantIdSql, pgx.NamedArgs{"tenant_id": tenantID})
	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[AuthUserEntity])
	return users, err
}

func SelectAllUsers(ctx context.Context, tx pgx.Tx) ([]AuthUserEntity, error) {
	rows, _ := tx.Query(ctx, selectAllUsersSql, pgx.NamedArgs{})
	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[AuthUserEntity])
	return users, err
}

// Role-related methods

func SelectUserRoles(ctx context.Context, tx pgx.Tx, userID string) ([]AuthRoleEntity, error) {
	rows, _ := tx.Query(ctx, selectUserRolesSql, pgx.NamedArgs{"user_id": userID})
	roles, err := pgx.CollectRows(rows, pgx.RowToStructByName[AuthRoleEntity])
	return roles, err
}

func InsertUserRole(ctx context.Context, tx pgx.Tx, userID string, roleID int) error {
	_, err := tx.Exec(ctx, insertUserRoleSql, pgx.NamedArgs{
		"user_id": userID,
		"role_id": roleID,
	})
	return err
}

func DeleteUserRole(ctx context.Context, tx pgx.Tx, userID string, roleID int) error {
	_, err := tx.Exec(ctx, deleteUserRoleSql, pgx.NamedArgs{
		"user_id": userID,
		"role_id": roleID,
	})
	return err
}

func SelectRoleByName(ctx context.Context, tx pgx.Tx, roleName string) (*AuthRoleEntity, error) {
	rows, _ := tx.Query(ctx, selectRoleByNameSql, pgx.NamedArgs{"name": roleName})
	ent, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AuthRoleEntity])
	if katpg.IsNoRows(err) {
		return nil, nil
	}
	return &ent, err
}

func SelectTenantByID(ctx context.Context, tx pgx.Tx, tenantID string) (*TenantEntity, error) {
	rows, _ := tx.Query(ctx, selectTenantByIdSql, pgx.NamedArgs{"id": tenantID})
	ent, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TenantEntity])
	if katpg.IsNoRows(err) {
		return nil, nil
	}
	return &ent, err
}

func SelectAllTenants(ctx context.Context, tx pgx.Tx) ([]TenantEntity, error) {
	rows, _ := tx.Query(ctx, selectAllTenantsSql)
	tenants, err := pgx.CollectRows(rows, pgx.RowToStructByName[TenantEntity])
	return tenants, err
}

func InsertTenant(ctx context.Context, tx pgx.Tx, tenant *TenantEntity) error {
	_, err := tx.Exec(ctx, insertTenantSql, pgx.NamedArgs{
		"id":          tenant.ID,
		"name":        tenant.Name,
		"description": tenant.Description,
		"created_at":  tenant.CreatedAt,
		"updated_at":  tenant.UpdatedAt,
	})
	return err
}

func UpdateTenant(ctx context.Context, tx pgx.Tx, tenant *TenantEntity) error {
	_, err := tx.Exec(ctx, updateTenantSql, pgx.NamedArgs{
		"id":          tenant.ID,
		"name":        tenant.Name,
		"description": tenant.Description,
		"updated_at":  tenant.UpdatedAt,
	})
	return err
}

func DeleteTenant(ctx context.Context, tx pgx.Tx, tenantID string) (int64, error) {
	cmd, err := tx.Exec(ctx, deleteTenantSql, pgx.NamedArgs{"id": tenantID})
	if err != nil {
		return 0, err
	}
	rowsAffected := cmd.RowsAffected()
	return rowsAffected, err
}

// User Profile methods

func SelectUserProfileByUserID(ctx context.Context, tx pgx.Tx, userID string) (*UserProfileEntity, error) {
	rows, _ := tx.Query(ctx, selectUserProfileByUserIdSql, pgx.NamedArgs{"user_id": userID})
	ent, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[UserProfileEntity])
	if katpg.IsNoRows(err) {
		return nil, nil
	}
	return &ent, err
}

func InsertUserProfile(ctx context.Context, tx pgx.Tx, userProfile *UserProfileEntity) (*UserProfileEntity, error) {
	rows, _ := tx.Query(ctx, insertUserProfileSql, pgx.NamedArgs{
		"user_id":    userProfile.UserID,
		"is_metric":  userProfile.IsMetric,
		"created_at": userProfile.CreatedAt,
		"updated_at": userProfile.UpdatedAt,
	})
	ent, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[UserProfileEntity])
	if err != nil {
		return nil, err
	}
	return &ent, nil
}

func UpdateUserProfile(ctx context.Context, tx pgx.Tx, userProfile *UserProfileEntity) (*UserProfileEntity, error) {
	query := `
		UPDATE user_profile
		SET height = @height, weight = @weight, gender = @gender, birth_date = @birth_date, is_metric = @is_metric, updated_at = @updated_at
		WHERE user_id = @user_id
		RETURNING id, user_id, height, weight, gender, birth_date, is_metric, created_at, updated_at
	`

	rows, _ := tx.Query(ctx, query, pgx.NamedArgs{
		"user_id":    userProfile.UserID,
		"height":     userProfile.Height,
		"weight":     userProfile.Weight,
		"gender":     userProfile.Gender,
		"birth_date": userProfile.BirthDate,
		"is_metric":  userProfile.IsMetric,
		"updated_at": userProfile.UpdatedAt,
	})
	ent, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[UserProfileEntity])
	if err != nil {
		return nil, err
	}
	return &ent, nil
}
