package postgres

import (
	"context"
	"fmt"
	"prservice/internal/repository"
)

func TruncateTables(ctx context.Context, db repository.Querier, tableName string) error {
	_, err := db.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tableName))
	return err
}
