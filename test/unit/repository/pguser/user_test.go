package pguser

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"prservice/internal/entity"
	errs "prservice/internal/errors"
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
	repo      repository.BaseUserRepository
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

	repo = repository.NewPostgresUserRepository(logger, pool)

	err = cleanupTables()
	if err != nil {
		slog.Error("Unable to cleanup tables", "err", err)
		os.Exit(1)
	}

	exitCode := m.Run()
	os.Exit(exitCode)
}

func createTeam(ctx context.Context, teamName string) error {
	_, err := pool.Exec(ctx, "INSERT INTO teams(name) VALUES ($1)", teamName)
	if err != nil {
		return err
	}
	return nil
}

func cleanupTables() error {
	err := repository.TruncateTables(globalCtx, pool, "users")
	if err != nil {
		return err
	}
	return repository.TruncateTables(globalCtx, pool, "teams")
}

func setupTest(t *testing.T) (context.Context, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	t.Cleanup(func() {
		err := cleanupTables()
		if err != nil {
			t.Fatalf("failed to cleanup tables: %v", err)
		}
	})
	return ctx, cancel
}

func TestAddUsers(t *testing.T) {
	t.Run("No teams", func(t *testing.T) {
		ctx, cancel := setupTest(t)
		defer cancel()

		users := make([]entity.User, 1)
		users[0] = entity.User{
			"u1",
			"test",
			"team",
			false,
		}

		err := repo.AddUsers(ctx, users)
		if err == nil {
			t.Error("AddUsers expected to fail")
		}
	})
	t.Run("One user", func(t *testing.T) {
		ctx, cancel := setupTest(t)
		defer cancel()

		err := createTeam(ctx, "team")
		if err != nil {
			t.Errorf("CreateTeam expected to succeed, got: %v", err)
		}

		users := make([]entity.User, 1)
		users[0] = entity.User{
			"u1",
			"test",
			"team",
			false,
		}

		err = repo.AddUsers(ctx, users)
		if err != nil {
			t.Errorf("AddUsers expected to succeed, got: %v", err)
		}
	})
	t.Run("Multiple users", func(t *testing.T) {
		ctx, cancel := setupTest(t)
		defer cancel()

		err := createTeam(ctx, "team")
		if err != nil {
			t.Errorf("CreateTeam expected to succeed, got: %v", err)
		}

		users := make([]entity.User, 3)
		for i := range users {
			users[i] = entity.User{
				fmt.Sprintf("u%d", i),
				fmt.Sprintf("user%d", i),
				"team",
				false,
			}
		}

		err = repo.AddUsers(ctx, users)
		if err != nil {
			t.Errorf("AddUsers expected to succeed, got: %v", err)
		}
	})
}

func TestGetById(t *testing.T) {
	t.Run("Invalid id", func(t *testing.T) {
		ctx, cancel := setupTest(t)
		defer cancel()

		_, err := repo.GetById(ctx, "invalid")
		if err == nil {
			t.Error("GetUser expected to fail")
		}
	})
	t.Run("All ok", func(t *testing.T) {
		ctx, cancel := setupTest(t)
		defer cancel()

		err := createTeam(ctx, "team")
		if err != nil {
			t.Errorf("CreateTeam expected to succeed, got: %v", err)
		}
		users := make([]entity.User, 1)
		users[0] = entity.User{
			Id:       "u1",
			Username: "test",
			TeamName: "team",
			IsActive: true,
		}
		err = repo.AddUsers(ctx, users)
		if err != nil {
			t.Errorf("AddUsers expected to succeed, got: %v", err)
		}

		res, err := repo.GetById(ctx, "u1")
		if err != nil {
			t.Errorf("GetById expected to succeed, got: %v", err)
		}
		if res == nil {
			t.Fatal("GetById expected to return a user, got nil instead")
		}
		if res.Id != "u1" {
			t.Errorf("GetById expected id = u1, got: %v", res.Id)
		}
	})
}

func TestGetByTeamName(t *testing.T) {
	t.Run("Invalid teamName", func(t *testing.T) {
		ctx, cancel := setupTest(t)
		defer cancel()

		res, err := repo.GetByTeamName(ctx, "invalid")
		if err != nil {
			t.Errorf("GetByTeamName expected to succeed, got: %v", err)
		}
		if len(res) != 0 {
			t.Errorf("GetByTeamName expected to return 0 users, got: %v", len(res))
		}
	})
	t.Run("No users", func(t *testing.T) {
		ctx, cancel := setupTest(t)
		defer cancel()

		err := createTeam(ctx, "team")
		if err != nil {
			t.Errorf("CreateTeam expected to succeed, got: %v", err)
		}

		res, err := repo.GetByTeamName(ctx, "team")
		if err != nil {
			t.Errorf("GetByTeamName expected to succeed, got: %v", err)
		}
		if len(res) != 0 {
			t.Errorf("GetByTeamName expected to return 0 users, got: %v", len(res))
		}
	})
	t.Run("All ok", func(t *testing.T) {
		ctx, cancel := setupTest(t)
		defer cancel()

		err := createTeam(ctx, "team0")
		if err != nil {
			t.Errorf("CreateTeam expected to succeed, got: %v", err)
		}
		err = createTeam(ctx, "team1")
		if err != nil {
			t.Errorf("CreateTeam expected to succeed, got: %v", err)
		}

		users := make([]entity.User, 10)
		for i := range users {
			users[i] = entity.User{
				Id:       fmt.Sprintf("u%d", i),
				Username: fmt.Sprintf("user%d", i),
				TeamName: fmt.Sprintf("team%d", i%2),
				IsActive: true,
			}
		}
		err = repo.AddUsers(ctx, users)
		if err != nil {
			t.Errorf("AddUsers expected to succeed, got: %v", err)
		}

		res, err := repo.GetByTeamName(ctx, "team1")
		if err != nil {
			t.Errorf("GetByTeamName expected to succeed, got: %v", err)
		}
		if len(res) != 5 {
			t.Errorf("GetByTeamName expected to return 5 users, got: %d", len(res))
		}
	})
}

func TestGetActiveByTeamName(t *testing.T) {
	t.Run("Invalid teamName", func(t *testing.T) {
		ctx, cancel := setupTest(t)
		defer cancel()

		res, err := repo.GetActiveByTeamName(ctx, "invalid")
		if err != nil {
			t.Errorf("GetActiveByTeamName expected to succeed, got: %v", err)
		}
		if len(res) != 0 {
			t.Errorf("GetActiveByTeamName expected to return 0 users, got: %v", len(res))
		}
	})
	t.Run("No users", func(t *testing.T) {
		ctx, cancel := setupTest(t)
		defer cancel()

		err := createTeam(ctx, "team")
		if err != nil {
			t.Errorf("CreateTeam expected to succeed, got: %v", err)
		}

		res, err := repo.GetActiveByTeamName(ctx, "team")
		if err != nil {
			t.Errorf("GetActiveByTeamName expected to succeed, got: %v", err)
		}
		if len(res) != 0 {
			t.Errorf("GetActiveByTeamName expected to return 0 users, got: %v", len(res))
		}
	})
	t.Run("No users active", func(t *testing.T) {
		ctx, cancel := setupTest(t)
		defer cancel()

		err := createTeam(ctx, "team")
		if err != nil {
			t.Errorf("CreateTeam expected to succeed, got: %v", err)
		}

		users := make([]entity.User, 10)
		for i := range users {
			users[i] = entity.User{
				Id:       fmt.Sprintf("u%d", i),
				Username: fmt.Sprintf("user%d", i),
				TeamName: "team",
				IsActive: false,
			}
		}
		err = repo.AddUsers(ctx, users)
		if err != nil {
			t.Errorf("AddUsers expected to succeed, got: %v", err)
		}

		res, err := repo.GetActiveByTeamName(ctx, "team")
		if err != nil {
			t.Errorf("GetActiveByTeamName expected to succeed, got: %v", err)
		}
		if len(res) != 0 {
			t.Errorf("GetActiveByTeamName expected to return 0 users, got: %v", len(res))
		}
	})
	t.Run("All ok", func(t *testing.T) {
		ctx, cancel := setupTest(t)
		defer cancel()

		err := createTeam(ctx, "team")
		if err != nil {
			t.Errorf("CreateTeam expected to succeed, got: %v", err)
		}

		users := make([]entity.User, 10)
		for i := range users {
			users[i] = entity.User{
				Id:       fmt.Sprintf("u%d", i),
				Username: fmt.Sprintf("user%d", i),
				TeamName: "team",
				IsActive: i%2 == 0,
			}
		}
		err = repo.AddUsers(ctx, users)
		if err != nil {
			t.Errorf("AddUsers expected to succeed, got: %v", err)
		}

		res, err := repo.GetActiveByTeamName(ctx, "team")
		if err != nil {
			t.Errorf("GetActiveByTeamName expected to succeed, got: %v", err)
		}
		if len(res) != 5 {
			t.Errorf("GetActiveByTeamName expected to return 5 users, got: %d", len(res))
		}
	})
}

func TestUpdateUser(t *testing.T) {
	t.Run("Invalid id", func(t *testing.T) {
		ctx, cancel := setupTest(t)
		defer cancel()

		newStatus := true
		err := repo.UpdateUser(ctx, "u1", &entity.UserUpdate{
			IsActive: &newStatus,
		})
		if err == nil {
			t.Error("UpdateUser expected to fail")
		}
	})
	t.Run("No filters", func(t *testing.T) {
		ctx, cancel := setupTest(t)
		defer cancel()

		err := createTeam(ctx, "team")
		if err != nil {
			t.Errorf("CreateTeam expected to succeed, got: %v", err)
		}
		users := make([]entity.User, 1)
		users[0] = entity.User{
			Id:       "u1",
			Username: "test",
			TeamName: "team",
			IsActive: false,
		}
		err = repo.AddUsers(ctx, users)
		if err != nil {
			t.Errorf("AddUsers expected to succeed, got: %v", err)
		}

		err = repo.UpdateUser(ctx, "u1", &entity.UserUpdate{})
		if err == nil {
			t.Error("UpdateUser expected to fail")
		} else {
			if !errors.Is(err, errs.ErrBaseBadFilter) {
				t.Errorf("UpdateUser error expected to be ErrBaseBadFilter, got: %v", err)
			}
		}
	})
	t.Run("All ok", func(t *testing.T) {
		ctx, cancel := setupTest(t)
		defer cancel()

		err := createTeam(ctx, "team")
		if err != nil {
			t.Errorf("CreateTeam expected to succeed, got: %v", err)
		}
		users := make([]entity.User, 1)
		users[0] = entity.User{
			Id:       "u1",
			Username: "test",
			TeamName: "team",
			IsActive: false,
		}
		err = repo.AddUsers(ctx, users)
		if err != nil {
			t.Errorf("AddUsers expected to succeed, got: %v", err)
		}

		newStatus := true
		err = repo.UpdateUser(ctx, "u1", &entity.UserUpdate{
			IsActive: &newStatus,
		})
		if err != nil {
			t.Errorf("UpdateUser expected to succeed, got: %v", err)
		}
	})
}
