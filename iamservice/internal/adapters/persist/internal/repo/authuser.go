package repo

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mobiletoly/gokatana/katpg"
	"time"
)

//go:generate go tool gobetter -input $GOFILE

type AuthUserRepo struct {
	db *pgxpool.Pool
}

func NewAuthRepo(db *pgxpool.Pool) *AuthUserRepo {
	return &AuthUserRepo{
		db: db,
	}
}

type AuthUserEntity struct { //+gob:Constructor
	ID            *string    `db:"id"`
	Email         string     `db:"email"`
	PasswordHash  string     `db:"password_hash"`
	FirstName     string     `db:"first_name"`
	LastName      string     `db:"last_name"`
	IsActive      bool       `db:"is_active"`
	EmailVerified bool       `db:"email_verified"`
	CreatedAt     *time.Time `db:"created_at"`
	UpdatedAt     *time.Time `db:"updated_at"`
}

type AuthRoleEntity struct { //+gob:Constructor
	ID          *int       `db:"id"`
	Name        string     `db:"name"`
	Description *string    `db:"description"`
	CreatedAt   *time.Time `db:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at"`
}

type AuthUserRoleEntity struct { //+gob:Constructor
	UserID     string     `db:"user_id"`
	RoleID     int        `db:"role_id"`
	AssignedAt *time.Time `db:"assigned_at"`
	AssignedBy *string    `db:"assigned_by"`
}

func (r *AuthUserRepo) SelectUserByEmail(ctx context.Context, email string) (*AuthUserEntity, error) {
	rows, _ := r.db.Query(ctx, selectUserByEmailSql, pgx.NamedArgs{"email": email})
	ent, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AuthUserEntity])
	if katpg.IsNoRows(err) {
		return nil, nil
	}
	return &ent, err
}

func (r *AuthUserRepo) SelectUserByID(ctx context.Context, userID string) (*AuthUserEntity, error) {
	rows, _ := r.db.Query(ctx, selectUserByIdSql, pgx.NamedArgs{"id": userID})
	ent, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AuthUserEntity])
	if katpg.IsNoRows(err) {
		return nil, nil
	}
	return &ent, err
}

func (r *AuthUserRepo) SelectUserWithPasswordByEmail(ctx context.Context, email string) (*AuthUserEntity, error) {
	rows, _ := r.db.Query(ctx, selectUserWithPasswordByEmailSql, pgx.NamedArgs{"email": email})
	ent, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AuthUserEntity])
	if katpg.IsNoRows(err) {
		return nil, nil
	}
	return &ent, err
}

func (r *AuthUserRepo) InsertUser(ctx context.Context, user *AuthUserEntity) error {
	_, err := r.db.Exec(ctx, insertUserSql, pgx.NamedArgs{
		"id":             *user.ID,
		"email":          user.Email,
		"password_hash":  user.PasswordHash,
		"first_name":     user.FirstName,
		"last_name":      user.LastName,
		"is_active":      user.IsActive,
		"email_verified": user.EmailVerified,
		"created_at":     *user.CreatedAt,
		"updated_at":     *user.UpdatedAt,
	})
	return err
}

func (r *AuthUserRepo) UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) error {
	// TODO This is a placeholder for future implementation
	// In a real system, you'd build dynamic UPDATE queries based on the updates map
	return nil
}

func (r *AuthUserRepo) DeleteUser(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, deleteUserSql, pgx.NamedArgs{"id": userID})
	return err
}

func (r *AuthUserRepo) SelectAllUsers(ctx context.Context) ([]AuthUserEntity, error) {
	rows, _ := r.db.Query(ctx, selectAllUsersSql, pgx.NamedArgs{})
	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[AuthUserEntity])
	return users, err
}

// Role-related methods

func (r *AuthUserRepo) SelectUserRoles(ctx context.Context, userID string) ([]AuthRoleEntity, error) {
	rows, _ := r.db.Query(ctx, selectUserRolesSql, pgx.NamedArgs{"user_id": userID})
	roles, err := pgx.CollectRows(rows, pgx.RowToStructByName[AuthRoleEntity])
	return roles, err
}

func (r *AuthUserRepo) InsertUserRole(ctx context.Context, userID string, roleID int, assignedBy *string) error {
	_, err := r.db.Exec(ctx, insertUserRoleSql, pgx.NamedArgs{
		"user_id":     userID,
		"role_id":     roleID,
		"assigned_by": assignedBy,
	})
	return err
}

func (r *AuthUserRepo) DeleteUserRole(ctx context.Context, userID string, roleID int) error {
	_, err := r.db.Exec(ctx, deleteUserRoleSql, pgx.NamedArgs{
		"user_id": userID,
		"role_id": roleID,
	})
	return err
}

func (r *AuthUserRepo) SelectRoleByName(ctx context.Context, roleName string) (*AuthRoleEntity, error) {
	rows, _ := r.db.Query(ctx, selectRoleByNameSql, pgx.NamedArgs{"name": roleName})
	ent, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AuthRoleEntity])
	if katpg.IsNoRows(err) {
		return nil, nil
	}
	return &ent, err
}
