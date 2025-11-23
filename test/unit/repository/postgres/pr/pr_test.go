package pr

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"prservice/internal/entity"
	errs "prservice/internal/errors"
	"prservice/internal/repository"
	"prservice/internal/repository/postgres"
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
	repo      repository.BasePullRequestRepository
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

	repo = postgres.NewPostgresPullRequestRepository(logger)

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

func createTeam(ctx context.Context, db repository.Querier, teamName string) error {
	query := `
		INSERT INTO teams (name) 
		VALUES ($1)
	`
	_, err := db.Exec(ctx, query, teamName)
	if err != nil {
		return err
	}
	return nil
}

func createUser(ctx context.Context, db repository.Querier, userId string, userName string, teamName string) error {
	query := `
		INSERT INTO users (id, username, team_name, is_active) 
		VALUES ($1, $2, $3, $4)
	`
	_, err := db.Exec(ctx, query, userId, userName, teamName, true)
	if err != nil {
		return err
	}
	return nil
}

func TestGetPullRequestsByReviewerId(t *testing.T) {
	t.Run("Invalid reviewerId", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		res, err := repo.GetPullRequestsByReviewerId(ctx, tx, "invalid_id")
		if err != nil {
			t.Fatalf("GetPullRequestsByReviewerId expected to succeed, got: %v", err)
		}
		if len(res) != 0 {
			t.Fatalf("GetPullRequestsByReviewerId expected to have len 0, got: %v", len(res))
		}
	})
	t.Run("No PRs with reviewerId", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
		if err != nil {
			t.Fatalf("createTeam expected to succeed, got: %v", err)
		}
		err = createUser(ctx, tx, "u1", "user1", "team")
		if err != nil {
			t.Fatalf("createUserWithTeam expected to succeed, got: %v", err)
		}

		res, err := repo.GetPullRequestsByReviewerId(ctx, tx, "u1")
		if err != nil {
			t.Fatalf("GetPullRequestsByReviewerId expected to succeed, got: %v", err)
		}
		if len(res) != 0 {
			t.Fatalf("GetPullRequestsByReviewerId expected to have len 0, got: %v", len(res))
		}
	})
	t.Run("All ok", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
		if err != nil {
			t.Fatalf("createTeam expected to succeed, got: %v", err)
		}
		err = createUser(ctx, tx, "u1", "user1", "team")
		if err != nil {
			t.Fatalf("createUserWithTeam expected to succeed, got: %v", err)
		}
		err = createUser(ctx, tx, "u2", "user2", "team")
		if err != nil {
			t.Fatalf("createUserWithTeam expected to succeed, got: %v", err)
		}
		err = repo.AddPullRequest(ctx, tx, &entity.PullRequest{
			Id:              "pr1",
			PullRequestName: "pr1",
			AuthorId:        "u1",
			Status:          entity.StatusOpen,
		})
		if err != nil {
			t.Fatalf("AddPullRequest expected to succeed, got: %v", err)
		}
		err = repo.AddPullRequest(ctx, tx, &entity.PullRequest{
			Id:              "pr2",
			PullRequestName: "pr2",
			AuthorId:        "u2",
			Status:          entity.StatusOpen,
		})
		if err != nil {
			t.Fatalf("AddPullRequest expected to succeed, got: %v", err)
		}
		err = repo.AddReviewerToPullRequest(ctx, tx, "pr1", "u2")
		if err != nil {
			t.Fatalf("AddReviewerToPullRequest expected to succeed, got: %v", err)
		}

		res, err := repo.GetPullRequestsByReviewerId(ctx, tx, "u2")
		if err != nil {
			t.Fatalf("GetPullRequestsByReviewerId expected to succeed, got: %v", err)
		}
		if len(res) != 1 {
			t.Fatalf("GetPullRequestsByReviewerId expected to have len 1, got: %v", len(res))
		}

		if err != nil {
			t.Fatalf("createUserWithTeam expected to succeed, got: %v", err)
		}
	})
}

func TestGetPullRequestById(t *testing.T) {
	t.Run("Invalid Id", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		_, err := repo.GetPullRequestById(ctx, tx, "invalid_id")
		if err != nil {
			if !errors.Is(err, errs.ErrBaseNotFound) {
				t.Fatalf("GetPullRequestById expected to fail with ErrBaseNotFound, got: %v", err)
			}
		}
	})
	t.Run("All ok", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
		if err != nil {
			t.Fatalf("createTeam expected to succeed, got: %v", err)
		}
		err = createUser(ctx, tx, "u1", "user1", "team")
		if err != nil {
			t.Fatalf("createUserWithTeam expected to succeed, got: %v", err)
		}

		err = repo.AddPullRequest(ctx, tx, &entity.PullRequest{
			"pr1",
			"pr1",
			"u1",
			entity.StatusOpen,
			nil,
		})
		if err != nil {
			t.Fatalf("AddPullRequest expected to succeed, got: %v", err)
		}

		res, err := repo.GetPullRequestById(ctx, tx, "pr1")
		if err != nil {
			t.Fatalf("GetPullRequestById expected to succeed, got: %v", err)
		}
		if res.Id != "pr1" {
			t.Fatalf("GetPullRequestById expected to have Id `pr1`, got: %v", res.Id)
		}
	})
}

func TestAddPullRequest(t *testing.T) {
	t.Run("Already exists", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
		if err != nil {
			t.Fatalf("createTeam expected to succeed, got: %v", err)
		}
		err = createUser(ctx, tx, "u1", "user1", "team")
		if err != nil {
			t.Fatalf("createUserWithTeam expected to succeed, got: %v", err)
		}

		ent := &entity.PullRequest{
			"pr1",
			"pr1",
			"u1",
			entity.StatusOpen,
			nil,
		}
		err = repo.AddPullRequest(ctx, tx, ent)
		if err != nil {
			t.Fatalf("AddPullRequest expected to succeed, got: %v", err)
		}
		err = repo.AddPullRequest(ctx, tx, ent)
		if err == nil {
			t.Fatal("AddPullRequest expected to fail")
		}
	})
	t.Run("All ok", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
		if err != nil {
			t.Fatalf("createTeam expected to succeed, got: %v", err)
		}
		err = createUser(ctx, tx, "u1", "user1", "team")
		if err != nil {
			t.Fatalf("createUserWithTeam expected to succeed, got: %v", err)
		}

		ent := &entity.PullRequest{
			"pr1",
			"pr1",
			"u1",
			entity.StatusOpen,
			nil,
		}
		err = repo.AddPullRequest(ctx, tx, ent)
		if err != nil {
			t.Fatalf("AddPullRequest expected to succeed, got: %v", err)
		}
	})
}

func TestUpdatePullRequestStatus(t *testing.T) {
	t.Run("Invalid Id", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := repo.UpdatePullRequestStatus(ctx, tx, "pr1", entity.StatusMerged)
		if err != nil {
			if !errors.Is(err, errs.ErrBaseNotFound) {
				t.Fatalf("UpdatePullRequestStatus expected to fail with ErrBaseNotFound, got: %v", err)
			}
		}
	})
	t.Run("All ok", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
		if err != nil {
			t.Fatalf("createTeam expected to succeed, got: %v", err)
		}
		err = createUser(ctx, tx, "u1", "user1", "team")
		if err != nil {
			t.Fatalf("createUserWithTeam expected to succeed, got: %v", err)
		}
		err = repo.AddPullRequest(ctx, tx, &entity.PullRequest{
			Id:              "pr1",
			PullRequestName: "pr1",
			AuthorId:        "u1",
			Status:          entity.StatusOpen,
		})
		if err != nil {
			t.Fatalf("AddPullRequest expected to succeed, got: %v", err)
		}

		err = repo.UpdatePullRequestStatus(ctx, tx, "pr1", entity.StatusMerged)
		if err != nil {
			t.Fatalf("UpdatePullRequestStatus expected to succeed, got: %v", err)
		}
	})
}

func TestAddReviewerToPullRequest(t *testing.T) {
	t.Run("Invalid reviewerId", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
		if err != nil {
			t.Fatalf("createTeam expected to succeed, got: %v", err)
		}
		err = createUser(ctx, tx, "u1", "user1", "team")
		if err != nil {
			t.Fatalf("createUserWithTeam expected to succeed, got: %v", err)
		}
		err = createUser(ctx, tx, "u2", "user2", "team")
		if err != nil {
			t.Fatalf("createUserWithTeam expected to succeed, got: %v", err)
		}

		err = repo.AddPullRequest(ctx, tx, &entity.PullRequest{
			Id:              "pr1",
			PullRequestName: "pr1",
			AuthorId:        "u1",
			Status:          entity.StatusOpen,
		})
		if err != nil {
			t.Fatalf("AddPullRequest expected to succeed, got: %v", err)
		}

		err = repo.AddReviewerToPullRequest(ctx, tx, "pr1", "u3")
		if err == nil {
			t.Fatal("AddReviewerToPullRequest expected to fail")
		}
	})
	t.Run("Invalid prId", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
		if err != nil {
			t.Fatalf("createTeam expected to succeed, got: %v", err)
		}
		err = createUser(ctx, tx, "u1", "user1", "team")
		if err != nil {
			t.Fatalf("createUserWithTeam expected to succeed, got: %v", err)
		}
		err = createUser(ctx, tx, "u2", "user2", "team")
		if err != nil {
			t.Fatalf("createUserWithTeam expected to succeed, got: %v", err)
		}

		err = repo.AddPullRequest(ctx, tx, &entity.PullRequest{
			Id:              "pr1",
			PullRequestName: "pr1",
			AuthorId:        "u1",
			Status:          entity.StatusOpen,
		})
		if err != nil {
			t.Fatalf("AddPullRequest expected to succeed, got: %v", err)
		}

		err = repo.AddReviewerToPullRequest(ctx, tx, "pr2", "u2")
		if err == nil {
			t.Fatal("AddReviewerToPullRequest expected to fail")
		}
	})
	t.Run("All ok", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
		if err != nil {
			t.Fatalf("createTeam expected to succeed, got: %v", err)
		}
		err = createUser(ctx, tx, "u1", "user1", "team")
		if err != nil {
			t.Fatalf("createUserWithTeam expected to succeed, got: %v", err)
		}
		err = createUser(ctx, tx, "u2", "user2", "team")
		if err != nil {
			t.Fatalf("createUserWithTeam expected to succeed, got: %v", err)
		}

		err = repo.AddPullRequest(ctx, tx, &entity.PullRequest{
			Id:              "pr1",
			PullRequestName: "pr1",
			AuthorId:        "u1",
			Status:          entity.StatusOpen,
		})
		if err != nil {
			t.Fatalf("AddPullRequest expected to succeed, got: %v", err)
		}

		err = repo.AddReviewerToPullRequest(ctx, tx, "pr1", "u2")
		if err != nil {
			t.Fatalf("AddReviewerToPullRequest expected to succeed, got: %v", err)
		}
	})
}

func TestRemoveReviewerFromPullRequest(t *testing.T) {
	t.Run("Invalid reviewerId", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
		if err != nil {
			t.Fatalf("createTeam expected to succeed, got: %v", err)
		}
		err = createUser(ctx, tx, "u1", "user1", "team")
		if err != nil {
			t.Fatalf("createUserWithTeam expected to succeed, got: %v", err)
		}
		err = createUser(ctx, tx, "u2", "user2", "team")
		if err != nil {
			t.Fatalf("createUserWithTeam expected to succeed, got: %v", err)
		}

		err = repo.AddPullRequest(ctx, tx, &entity.PullRequest{
			Id:              "pr1",
			PullRequestName: "pr1",
			AuthorId:        "u1",
			Status:          entity.StatusOpen,
		})
		if err != nil {
			t.Fatalf("AddPullRequest expected to succeed, got: %v", err)
		}

		err = repo.RemoveReviewerFromPullRequest(ctx, tx, "pr1", "u3")
		if err == nil {
			t.Fatal("AddReviewerToPullRequest expected to fail")
		}
	})
	t.Run("Invalid prId", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
		if err != nil {
			t.Fatalf("createTeam expected to succeed, got: %v", err)
		}
		err = createUser(ctx, tx, "u1", "user1", "team")
		if err != nil {
			t.Fatalf("createUserWithTeam expected to succeed, got: %v", err)
		}
		err = createUser(ctx, tx, "u2", "user2", "team")
		if err != nil {
			t.Fatalf("createUserWithTeam expected to succeed, got: %v", err)
		}

		err = repo.AddPullRequest(ctx, tx, &entity.PullRequest{
			Id:              "pr1",
			PullRequestName: "pr1",
			AuthorId:        "u1",
			Status:          entity.StatusOpen,
		})
		if err != nil {
			t.Fatalf("AddPullRequest expected to succeed, got: %v", err)
		}

		err = repo.RemoveReviewerFromPullRequest(ctx, tx, "pr2", "u2")
		if err == nil {
			t.Fatal("AddReviewerToPullRequest expected to fail")
		}
	})
	t.Run("Not found", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
		if err != nil {
			t.Fatalf("createTeam expected to succeed, got: %v", err)
		}
		err = createUser(ctx, tx, "u1", "user1", "team")
		if err != nil {
			t.Fatalf("createUserWithTeam expected to succeed, got: %v", err)
		}
		err = createUser(ctx, tx, "u2", "user2", "team")
		if err != nil {
			t.Fatalf("createUserWithTeam expected to succeed, got: %v", err)
		}

		err = repo.AddPullRequest(ctx, tx, &entity.PullRequest{
			Id:              "pr1",
			PullRequestName: "pr1",
			AuthorId:        "u1",
			Status:          entity.StatusOpen,
		})
		if err != nil {
			t.Fatalf("AddPullRequest expected to succeed, got: %v", err)
		}

		err = repo.AddReviewerToPullRequest(ctx, tx, "pr1", "u2")
		if err != nil {
			t.Fatalf("AddReviewerToPullRequest expected to succeed, got: %v", err)
		}

		err = repo.RemoveReviewerFromPullRequest(ctx, tx, "pr1", "u1")
		if err != nil {
			if !errors.Is(err, errs.ErrBaseNotFound) {
				t.Fatalf("AddReviewerToPullRequest expected to fail with ErrBaseNotFound, got: %v", err)
			}
		}
	})
	t.Run("All ok", func(t *testing.T) {
		ctx, cancel, tx := setupTest(t)
		defer cancel()

		err := createTeam(ctx, tx, "team")
		if err != nil {
			t.Fatalf("createTeam expected to succeed, got: %v", err)
		}
		err = createUser(ctx, tx, "u1", "user1", "team")
		if err != nil {
			t.Fatalf("createUserWithTeam expected to succeed, got: %v", err)
		}
		err = createUser(ctx, tx, "u2", "user2", "team")
		if err != nil {
			t.Fatalf("createUserWithTeam expected to succeed, got: %v", err)
		}

		err = repo.AddPullRequest(ctx, tx, &entity.PullRequest{
			Id:              "pr1",
			PullRequestName: "pr1",
			AuthorId:        "u1",
			Status:          entity.StatusOpen,
		})
		if err != nil {
			t.Fatalf("AddPullRequest expected to succeed, got: %v", err)
		}

		err = repo.AddReviewerToPullRequest(ctx, tx, "pr1", "u2")
		if err != nil {
			t.Fatalf("AddReviewerToPullRequest expected to succeed, got: %v", err)
		}

		err = repo.RemoveReviewerFromPullRequest(ctx, tx, "pr1", "u2")
		if err != nil {
			t.Fatalf("RemoveReviewerFromPullRequest expected to succeed, got: %v", err)
		}
	})
}
