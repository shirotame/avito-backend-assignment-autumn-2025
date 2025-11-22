package pgteam

import (
	"context"
	"log/slog"
	"os"
	"prservice/internal/entity"
	"prservice/internal/repository"
	"testing"
	"time"

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
		slog.Error("Unable to load env", "err", err)
		os.Exit(1)
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

	repo = repository.NewPostgresTeamRepository(logger, pool)

	err = cleanupTables()
	if err != nil {
		slog.Error("Unable to cleanup tables", "err", err)
		os.Exit(1)
	}

	exitCode := m.Run()
	os.Exit(exitCode)
}

func cleanupTables() error {
	return repository.TruncateTables(globalCtx, pool, "teams")
}

func setupTest(t *testing.T) (context.Context, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	t.Cleanup(func() {
		err := cleanupTables()
		if err != nil {
			t.Fatalf("error cleaning up tables: %v", err)
		}
	})
	return ctx, cancel
}

func TestAddTeam(t *testing.T) {
	t.Run("Already exists", func(t *testing.T) {
		ctx, cancel := setupTest(t)
		defer cancel()

		err := repo.AddTeam(ctx, &entity.Team{
			TeamName: "test",
		})
		if err != nil {
			t.Fatalf("AddTeam expected to succeed, got: %v", err)
		}

		err = repo.AddTeam(ctx, &entity.Team{
			TeamName: "test",
		})
		if err == nil {
			t.Fatal("AddTeam expected to fail")
		}
	})
	t.Run("All ok", func(t *testing.T) {
		ctx, cancel := setupTest(t)
		defer cancel()

		err := repo.AddTeam(ctx, &entity.Team{
			TeamName: "test",
		})
		if err != nil {
			t.Fatalf("AddTeam expected to succeed, got: %v", err)
		}
	})
}

func TestGetTeam(t *testing.T) {
	t.Run("Invalid name", func(t *testing.T) {
		ctx, cancel := setupTest(t)
		defer cancel()

		_, err := repo.GetTeam(ctx, "test")
		if err == nil {
			t.Fatal("GetTeam expected to fail")
		}
	})
	t.Run("All ok", func(t *testing.T) {
		ctx, cancel := setupTest(t)
		defer cancel()

		err := repo.AddTeam(ctx, &entity.Team{
			TeamName: "test",
		})
		if err != nil {
			t.Fatalf("AddTeam expected to succeed, got: %v", err)
		}

		team, err := repo.GetTeam(ctx, "test")
		if err != nil {
			t.Fatalf("GetTeam expected to succeed, got: %v", err)
		}
		if team.TeamName != "test" {
			t.Fatalf("GetTeam expected TeamName to be `test`, got: %v", team.TeamName)
		}
	})
}
