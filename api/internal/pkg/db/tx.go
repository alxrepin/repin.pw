package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type txKey struct{}

type TxRunner struct {
	client *Client
}

func NewTxRunner(client *Client) *TxRunner { return &TxRunner{client: client} }

func (r *TxRunner) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return fn(ctx)
	}

	tx, err := r.client.Connection().BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func Executor(ctx context.Context, fallback sqlx.ExtContext) sqlx.ExtContext {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx
	}

	return fallback
}
