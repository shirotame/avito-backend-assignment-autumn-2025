package team

import (
	"context"
	"errors"
	"github.com/shirotame/avito-backend-assignment-autumn-2025/internal/entity"
	errs "github.com/shirotame/avito-backend-assignment-autumn-2025/internal/errors"
	"github.com/shirotame/avito-backend-assignment-autumn-2025/internal/repository"
	"github.com/shirotame/avito-backend-assignment-autumn-2025/internal/repository/postgres"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var (
	globalCtx = context.Background()
	pool      *pgxpool.Pool
	logger    *slog.Logger
	repo      repository.BaseTeamRepository
)

func TestMain(m *testing.M) {
	err := godotenv.Load("..\\test.env")
	if err != nil {
		slog.Error("unable to load env, using default environment variables", "err", err)
	}
	logHandler := slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		})

	logger = slog.New(logHandler)
	slog.SetDefault(logger)

	connString := os.Getenv("TEST_DATABASE_URL")
	pool, err = pgxpool.New(globalCtx, connString)
	if err != nil {
		slog.Error("Unable to connect to database", "err", err)
		os.Exit(1)
	}

	if err := pool.Ping(globalCtx); err != nil {
		slog.Error("Unable to ping database", "err", err)
		os.Exit(1)
	}

	repo = postgres.NewPostgresTeamRepository(logger)

	exitCode := m.Run()
	os.Exit(exitCode)
}

func setupTest(t *testing.T) (context.Context, func(), pgx.Tx) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("failed to start transaction: %v", err)
	}

	t.Cleanup(func() {
		cancel()
		if err := tx.Rollback(globalCtx); err != nil {
			t.Fatalf("error rolling back: %v", err)
		}
	})
	return ctx, cancel, tx
}
func TestAddTeam(t *testing.T) {
	t.Run("Already exists", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := repo.AddTeam(ctx, tx, &entity.Team{
			TeamName: "test",
		})
		if err != nil {
			t.Fatalf("AddTeam expected to succeed, got: %v", err)
		}

		err = repo.AddTeam(ctx, tx, &entity.Team{
			TeamName: "test",
		})
		if err == nil {
			t.Fatal("AddTeam expected to fail")
		}
	})
	t.Run("All ok", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := repo.AddTeam(ctx, tx, &entity.Team{
			TeamName: "test",
		})
		if err != nil {
			t.Fatalf("AddTeam expected to succeed, got: %v", err)
		}
	})
}

func TestGetTeam(t *testing.T) {
	t.Run("Invalid name", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		_, err := repo.GetTeam(ctx, tx, "test")
		if err != nil {
			if !errors.Is(err, errs.ErrBaseNotFound) {
				t.Fatalf("GetTeam expected to fail with ErrBaseNotFound, got: %v", err)
			}
		}
	})
	t.Run("All ok", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := repo.AddTeam(ctx, tx, &entity.Team{
			TeamName: "test",
		})
		if err != nil {
			t.Fatalf("AddTeam expected to succeed, got: %v", err)
		}

		team, err := repo.GetTeam(ctx, tx, "test")
		if err != nil {
			t.Fatalf("GetTeam expected to succeed, got: %v", err)
		}
		if team.TeamName != "test" {
			t.Fatalf("GetTeam expected TeamName to be `test`, got: %v", team.TeamName)
		}
	})
}
