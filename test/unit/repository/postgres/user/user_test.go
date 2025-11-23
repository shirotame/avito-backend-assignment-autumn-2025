package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/shirotame/avito-backend-assignment-autumn-2025/internal/entity"
	errs "github.com/shirotame/avito-backend-assignment-autumn-2025/internal/errors"
	"github.com/shirotame/avito-backend-assignment-autumn-2025/internal/repository"
	"github.com/shirotame/avito-backend-assignment-autumn-2025/internal/repository/postgres"

	"github.com/jackc/pgx/v5"
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
		slog.Error("unable to load env, using default environment variables")
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

	repo = postgres.NewPostgresUserRepository(logger)

	exitCode := m.Run()
	os.Exit(exitCode)
}

func createTeam(ctx context.Context, db repository.Querier, teamName string) error {
	_, err := db.Exec(ctx, "INSERT INTO teams(name) VALUES ($1)", teamName)
	if err != nil {
		return err
	}
	return nil
}

func createPrAndAssign(
	ctx context.Context,
	db repository.Querier,
	prId string,
	authorId string,
	reviewerId string,
) error {
	query := `
		INSERT INTO pull_requests (id, name, author_id, status)
		VALUES ($1, $2, $3, $4)
	`
	_, err := db.Exec(ctx, query, prId, prId, authorId, entity.StatusOpen)
	if err != nil {
		return err
	}

	query = `
		INSERT INTO pull_requests_users (user_id, pr_id)
		VALUES ($1, $2)
	`
	_, err = db.Exec(ctx, query, reviewerId, prId)
	if err != nil {
		return err
	}
	return nil
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

func TestAddUsers(t *testing.T) {
	t.Run("No teams", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		users := make([]entity.User, 1)
		users[0] = entity.User{
			Id:       "u1",
			Username: "test",
			TeamName: "team",
		}

		err := repo.AddUsers(ctx, tx, users)
		if err == nil {
			t.Error("AddUsers expected to fail")
		}
	})
	t.Run("One user", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
		if err != nil {
			t.Errorf("CreateTeam expected to succeed, got: %v", err)
		}

		users := make([]entity.User, 1)
		users[0] = entity.User{
			Id:       "u1",
			Username: "test",
			TeamName: "team",
		}

		err = repo.AddUsers(ctx, tx, users)
		if err != nil {
			t.Errorf("AddUsers expected to succeed, got: %v", err)
		}
	})
	t.Run("Multiple users", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
		if err != nil {
			t.Errorf("CreateTeam expected to succeed, got: %v", err)
		}

		users := make([]entity.User, 3)
		for i := range users {
			users[i] = entity.User{
				Id:       fmt.Sprintf("u%d", i),
				Username: fmt.Sprintf("user%d", i),
				TeamName: "team",
			}
		}

		err = repo.AddUsers(ctx, tx, users)
		if err != nil {
			t.Errorf("AddUsers expected to succeed, got: %v", err)
		}
	})
}

func TestGetById(t *testing.T) {
	t.Run("Invalid id", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		_, err := repo.GetById(ctx, tx, "invalid")
		if err != nil {
			if !errors.Is(err, errs.ErrBaseNotFound) {
				t.Fatalf("TestGetById expected to fail with ErrBaseNotFound, got: %v", err)
			}
		}
	})
	t.Run("All ok", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
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
		err = repo.AddUsers(ctx, tx, users)
		if err != nil {
			t.Errorf("AddUsers expected to succeed, got: %v", err)
		}

		res, err := repo.GetById(ctx, tx, "u1")
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

func TestGetReviewersByPrId(t *testing.T) {
	t.Run("No reviewers", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		res, err := repo.GetReviewersByPrId(ctx, tx, "pr1")
		if err != nil {
			t.Errorf("GetReviewersByPrId expected to succeed, got: %v", err)
		}
		if len(res) != 0 {
			t.Errorf("GetReviewersByPrId expected to return a 0 users, got %v", len(res))
		}
	})
	t.Run("With reviewers", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
		if err != nil {
			t.Errorf("CreateTeam expected to succeed, got: %v", err)
		}
		users := make([]entity.User, 3)
		for i := range users {
			users[i] = entity.User{
				Id:       fmt.Sprintf("u%d", i),
				Username: fmt.Sprintf("user%d", i),
				TeamName: "team",
			}
		}
		err = repo.AddUsers(ctx, tx, users)
		if err != nil {
			t.Errorf("AddUsers expected to succeed, got: %v", err)
		}

		err = createPrAndAssign(ctx, tx, "pr1", "u1", "u2")
		if err != nil {
			t.Errorf("CreatePrAndAssign expected to succeed, got: %v", err)
		}

		res, err := repo.GetReviewersByPrId(ctx, tx, "pr1")
		if err != nil {
			t.Errorf("GetReviewersByPrId expected to succeed, got: %v", err)
		}
		if len(res) != 1 {
			t.Errorf("GetReviewersByPrId expected to return a 1 users, got %v", len(res))
		}
	})
}

func TestGetByTeamName(t *testing.T) {
	t.Run("Invalid teamName", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		res, err := repo.GetByTeamName(ctx, tx, "invalid")
		if err != nil {
			t.Errorf("GetByTeamName expected to succeed, got: %v", err)
		}
		if len(res) != 0 {
			t.Errorf("GetByTeamName expected to return 0 users, got: %v", len(res))
		}
	})
	t.Run("No users", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
		if err != nil {
			t.Errorf("CreateTeam expected to succeed, got: %v", err)
		}

		res, err := repo.GetByTeamName(ctx, tx, "team")
		if err != nil {
			t.Errorf("GetByTeamName expected to succeed, got: %v", err)
		}
		if len(res) != 0 {
			t.Errorf("GetByTeamName expected to return 0 users, got: %v", len(res))
		}
	})
	t.Run("All ok", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team0")
		if err != nil {
			t.Errorf("CreateTeam expected to succeed, got: %v", err)
		}
		err = createTeam(ctx, tx, "team1")
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
		err = repo.AddUsers(ctx, tx, users)
		if err != nil {
			t.Errorf("AddUsers expected to succeed, got: %v", err)
		}

		res, err := repo.GetByTeamName(ctx, tx, "team1")
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
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		res, err := repo.GetActiveByTeamName(ctx, tx, "invalid")
		if err != nil {
			t.Errorf("GetActiveByTeamName expected to succeed, got: %v", err)
		}
		if len(res) != 0 {
			t.Errorf("GetActiveByTeamName expected to return 0 users, got: %v", len(res))
		}
	})
	t.Run("No users", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
		if err != nil {
			t.Errorf("CreateTeam expected to succeed, got: %v", err)
		}

		res, err := repo.GetActiveByTeamName(ctx, tx, "team")
		if err != nil {
			t.Errorf("GetActiveByTeamName expected to succeed, got: %v", err)
		}
		if len(res) != 0 {
			t.Errorf("GetActiveByTeamName expected to return 0 users, got: %v", len(res))
		}
	})
	t.Run("No users active", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
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
		err = repo.AddUsers(ctx, tx, users)
		if err != nil {
			t.Errorf("AddUsers expected to succeed, got: %v", err)
		}

		res, err := repo.GetActiveByTeamName(ctx, tx, "team")
		if err != nil {
			t.Errorf("GetActiveByTeamName expected to succeed, got: %v", err)
		}
		if len(res) != 0 {
			t.Errorf("GetActiveByTeamName expected to return 0 users, got: %v", len(res))
		}
	})
	t.Run("All ok", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
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
		err = repo.AddUsers(ctx, tx, users)
		if err != nil {
			t.Errorf("AddUsers expected to succeed, got: %v", err)
		}

		res, err := repo.GetActiveByTeamName(ctx, tx, "team")
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
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		newStatus := true
		err := repo.UpdateUser(ctx, tx, "u1", &entity.UserUpdate{
			IsActive: &newStatus,
		})
		if err != nil {
			if !errors.Is(err, errs.ErrBaseNotFound) {
				t.Fatalf("UpdateUser expected to fail with ErrBaseNotFound, got: %v", err)
			}
		}
	})
	t.Run("No filters", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
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
		err = repo.AddUsers(ctx, tx, users)
		if err != nil {
			t.Errorf("AddUsers expected to succeed, got: %v", err)
		}

		err = repo.UpdateUser(ctx, tx, "u1", &entity.UserUpdate{})
		if err != nil {
			if !errors.Is(err, errs.ErrBaseBadFilter) {
				t.Errorf("UpdateUser error expected to be ErrBaseBadFilter, got: %v", err)
			}
		}
	})
	t.Run("All ok", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
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
		err = repo.AddUsers(ctx, tx, users)
		if err != nil {
			t.Errorf("AddUsers expected to succeed, got: %v", err)
		}

		newStatus := true
		err = repo.UpdateUser(ctx, tx, "u1", &entity.UserUpdate{
			IsActive: &newStatus,
		})
		if err != nil {
			t.Errorf("UpdateUser expected to succeed, got: %v", err)
		}
	})
}
