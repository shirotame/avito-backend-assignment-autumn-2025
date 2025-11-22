package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TruncateTables(ctx context.Context, pool *pgxpool.Pool, tableName string) error {
	_, err := pool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tableName))
	return err
}
