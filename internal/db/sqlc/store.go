package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
	*Queries
}



func NewStore(db *pgxpool.Pool) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

func (store *Store) ExecTx(ctx context.Context, fn func(*Queries) error) error {
    tx, err := store.db.Begin(ctx)
    if err != nil {
        return err
    }
    
    q := store.WithTx(tx)
    err = fn(q)
    if err != nil {
        if rbErr := tx.Rollback(ctx); rbErr != nil {
            return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
        }
        return err
    }
    
    return tx.Commit(ctx)
}

func (store *Store) WithTx(tx pgx.Tx) *Queries {
    return store.Queries.WithTx(tx)
}