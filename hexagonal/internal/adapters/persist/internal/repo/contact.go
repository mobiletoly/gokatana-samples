package repo

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mobiletoly/gokatana/katpg"
)

//go:generate go tool gobetter -input $GOFILE

type ContactRepo struct {
	db *pgxpool.Pool
}

func NewContactRepo(db *pgxpool.Pool) *ContactRepo {
	return &ContactRepo{
		db: db,
	}
}

type ContactEntity struct { //+gob:Constructor
	ID        *int64 `db:"id"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
}

func (r *ContactRepo) SelectContactByID(ctx context.Context, ID int64) (*ContactEntity, error) {
	rows, _ := r.db.Query(ctx, selectContactByIdSql, pgx.NamedArgs{"id": ID})
	ent, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ContactEntity])
	if katpg.IsNoRows(err) {
		return nil, nil
	}
	return &ent, err
}

func (r *ContactRepo) SelectAllContacts(ctx context.Context) ([]ContactEntity, error) {
	rows, _ := r.db.Query(ctx, selectAllContactsSql)
	ents, err := pgx.CollectRows(rows, pgx.RowToStructByName[ContactEntity])
	return ents, err
}

func (r *ContactRepo) InsertContact(ctx context.Context, c *ContactEntity) (ID int64, err error) {
	err = r.db.QueryRow(ctx, insertContactSql, pgx.NamedArgs{
		"first_name": c.FirstName,
		"last_name":  c.LastName,
	}).Scan(&ID)
	return ID, err
}
