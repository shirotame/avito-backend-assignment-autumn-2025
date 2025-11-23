package e2e

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"prservice/internal/entity"
	errs "prservice/internal/errors"
	"prservice/internal/repository/postgres"
	"prservice/internal/service"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var (
	globalCtx   = context.Background()
	pool        *pgxpool.Pool
	logger      *slog.Logger
	userService service.BaseUserService
	teamService service.BaseTeamService
	prService   service.BasePullRequestService
)

func TestMain(m *testing.M) {
	err := godotenv.Load("test.env")
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

	prRepo := postgres.NewPostgresPullRequestRepository(logger)
	userRepo := postgres.NewPostgresUserRepository(logger)
	teamRepo := postgres.NewPostgresTeamRepository(logger)

	prService = service.NewPullRequestService(logger, pool, prRepo, userRepo)
	userService = service.NewUserService(logger, pool, userRepo, prRepo)
	teamService = service.NewTeamService(logger, pool, userRepo, teamRepo)

	_, err = pool.Exec(globalCtx, "TRUNCATE TABLE pull_requests_users, pull_requests, users, teams RESTART IDENTITY CASCADE")
	if err != nil {
		logger.Error("failed to truncate tables", "err", err)
		os.Exit(1)
	}

	exitCode := m.Run()
	os.Exit(exitCode)
}

func setupTest(t *testing.T) context.Context {
	ctx := context.Background()

	t.Cleanup(func() {
		_, err := pool.Exec(globalCtx, "TRUNCATE TABLE pull_requests_users, pull_requests, users, teams RESTART IDENTITY CASCADE")
		if err != nil {
			t.Fatalf("failed to truncate tables: %v", err)
		}
	})
	return ctx
}

func TestCreatePullRequest(t *testing.T) {
	t.Run(">2 available", func(t *testing.T) {
		ctx := setupTest(t)

		users := make([]entity.UserDTO, 10)
		for i := range users {
			users[i] = entity.UserDTO{
				UserId:   fmt.Sprintf("u%d", i),
				Username: fmt.Sprintf("user%d", i),
				IsActive: true,
			}
		}
		_, err := teamService.AddTeam(ctx, entity.TeamDTO{
			TeamName: "team1",
			Members:  users,
		})
		if err != nil {
			t.Fatalf("AddTeam should succeed, got: %v", err)
		}

		res, err := prService.CreatePullRequest(ctx, entity.PullRequestCreateDTO{
			PullRequestId:   "pr1",
			PullRequestName: "pr1",
			AuthorId:        "u1",
		})
		if err != nil {
			t.Fatalf("CreatePullRequest should succeed, got: %v", err)
		}
		if len(res.PullRequest.AssignedReviewers) != 2 {
			t.Fatalf("AssignedReviewers expected 2, got: %d", len(res.PullRequest.AssignedReviewers))
		}
	})
	t.Run("2 available (with author)", func(t *testing.T) {
		ctx := setupTest(t)

		users := make([]entity.UserDTO, 2)
		for i := range users {
			users[i] = entity.UserDTO{
				UserId:   fmt.Sprintf("u%d", i),
				Username: fmt.Sprintf("user%d", i),
				IsActive: true,
			}
		}
		_, err := teamService.AddTeam(ctx, entity.TeamDTO{
			TeamName: "team1",
			Members:  users,
		})
		if err != nil {
			t.Fatalf("AddTeam should succeed, got: %v", err)
		}

		res, err := prService.CreatePullRequest(ctx, entity.PullRequestCreateDTO{
			PullRequestId:   "pr1",
			PullRequestName: "pr1",
			AuthorId:        "u1",
		})
		if err != nil {
			t.Fatalf("CreatePullRequest should succeed, got: %v", err)
		}
		if len(res.PullRequest.AssignedReviewers) != 1 {
			t.Fatalf("AssignedReviewers expected 1, got: %d", len(res.PullRequest.AssignedReviewers))
		}
	})
	t.Run("1 available (author)", func(t *testing.T) {
		ctx := setupTest(t)

		users := make([]entity.UserDTO, 1)
		for i := range users {
			users[i] = entity.UserDTO{
				UserId:   fmt.Sprintf("u%d", i),
				Username: fmt.Sprintf("user%d", i),
				IsActive: true,
			}
		}
		_, err := teamService.AddTeam(ctx, entity.TeamDTO{
			TeamName: "team1",
			Members:  users,
		})
		if err != nil {
			t.Fatalf("AddTeam should succeed, got: %v", err)
		}

		res, err := prService.CreatePullRequest(ctx, entity.PullRequestCreateDTO{
			PullRequestId:   "pr1",
			PullRequestName: "pr1",
			AuthorId:        "u0",
		})
		if err != nil {
			t.Fatalf("CreatePullRequest should succeed, got: %v", err)
		}
		if len(res.PullRequest.AssignedReviewers) != 0 {
			t.Fatalf("AssignedReviewers expected 0, got: %d", len(res.PullRequest.AssignedReviewers))
		}
	})
	t.Run("0 available (author is active)", func(t *testing.T) {
		ctx := setupTest(t)

		users := make([]entity.UserDTO, 2)
		for i := range users {
			users[i] = entity.UserDTO{
				UserId:   fmt.Sprintf("u%d", i),
				Username: fmt.Sprintf("user%d", i),
				IsActive: i%2 == 0,
			}
		}
		_, err := teamService.AddTeam(ctx, entity.TeamDTO{
			TeamName: "team1",
			Members:  users,
		})
		if err != nil {
			t.Fatalf("AddTeam should succeed, got: %v", err)
		}

		res, err := prService.CreatePullRequest(ctx, entity.PullRequestCreateDTO{
			PullRequestId:   "pr1",
			PullRequestName: "pr1",
			AuthorId:        "u0",
		})
		if err != nil {
			t.Fatalf("CreatePullRequest should succeed, got: %v", err)
		}
		if len(res.PullRequest.AssignedReviewers) != 0 {
			t.Fatalf("AssignedReviewers expected 0, got: %d", len(res.PullRequest.AssignedReviewers))
		}
	})
}

func TestReassignPullRequest(t *testing.T) {
	t.Run(">2 available", func(t *testing.T) {
		ctx := setupTest(t)

		users := make([]entity.UserDTO, 10)
		for i := range users {
			users[i] = entity.UserDTO{
				UserId:   fmt.Sprintf("u%d", i),
				Username: fmt.Sprintf("user%d", i),
				IsActive: true,
			}
		}
		_, err := teamService.AddTeam(ctx, entity.TeamDTO{
			TeamName: "team1",
			Members:  users,
		})
		if err != nil {
			t.Fatalf("AddTeam should succeed, got: %v", err)
		}

		res, err := prService.CreatePullRequest(ctx, entity.PullRequestCreateDTO{
			PullRequestId:   "pr1",
			PullRequestName: "pr1",
			AuthorId:        "u0",
		})
		if err != nil {
			t.Fatalf("CreatePullRequest should succeed, got: %v", err)
		}
		if len(res.PullRequest.AssignedReviewers) != 2 {
			t.Fatalf("AssignedReviewers expected 2, got: %d", len(res.PullRequest.AssignedReviewers))
		}

		res, err = prService.ReassignPullRequest(ctx, entity.ReassignPullRequestDTO{
			PullRequestId: "pr1",
			OldReviewerId: res.PullRequest.AssignedReviewers[0],
		})
		if err != nil {
			t.Fatalf("ReassignPullRequest should succeed, got: %v", err)
		}
		if len(res.PullRequest.AssignedReviewers) != 2 {
			t.Fatalf("AssignedReviewers expected 2, got: %d", len(res.PullRequest.AssignedReviewers))
		}
	})
	t.Run("1 available (author)", func(t *testing.T) {
		ctx := setupTest(t)

		users := make([]entity.UserDTO, 3)
		for i := range users {
			users[i] = entity.UserDTO{
				UserId:   fmt.Sprintf("u%d", i),
				Username: fmt.Sprintf("user%d", i),
				IsActive: true,
			}
		}
		_, err := teamService.AddTeam(ctx, entity.TeamDTO{
			TeamName: "team1",
			Members:  users,
		})
		if err != nil {
			t.Fatalf("AddTeam should succeed, got: %v", err)
		}

		res, err := prService.CreatePullRequest(ctx, entity.PullRequestCreateDTO{
			PullRequestId:   "pr1",
			PullRequestName: "pr1",
			AuthorId:        "u0",
		})
		if err != nil {
			t.Fatalf("CreatePullRequest should succeed, got: %v", err)
		}
		if len(res.PullRequest.AssignedReviewers) != 2 {
			t.Fatalf("AssignedReviewers expected 2, got: %d", len(res.PullRequest.AssignedReviewers))
		}

		_, err = userService.SetIsActive(ctx, entity.SetUserIsActiveDTO{
			UserId:   "u1",
			IsActive: false,
		})
		if err != nil {
			t.Fatalf("SetIsActive should succeed, got: %v", err)
		}
		_, err = userService.SetIsActive(ctx, entity.SetUserIsActiveDTO{
			UserId:   "u2",
			IsActive: false,
		})
		if err != nil {
			t.Fatalf("SetIsActive should succeed, got: %v", err)
		}

		_, err = prService.ReassignPullRequest(ctx, entity.ReassignPullRequestDTO{
			PullRequestId: "pr1",
			OldReviewerId: res.PullRequest.AssignedReviewers[0],
		})
		if err != nil {
			if !errors.Is(err, errs.ErrNoActiveUsers) {
				t.Fatalf("ReassignPullRequest expected error ErrNoActiveUsers, got: %v", err)
			}
		}
	})
	t.Run("0 available", func(t *testing.T) {
		ctx := setupTest(t)

		users := make([]entity.UserDTO, 2)
		for i := range users {
			users[i] = entity.UserDTO{
				UserId:   fmt.Sprintf("u%d", i),
				Username: fmt.Sprintf("user%d", i),
				IsActive: true,
			}
		}
		_, err := teamService.AddTeam(ctx, entity.TeamDTO{
			TeamName: "team1",
			Members:  users,
		})
		if err != nil {
			t.Fatalf("AddTeam should succeed, got: %v", err)
		}

		res, err := prService.CreatePullRequest(ctx, entity.PullRequestCreateDTO{
			PullRequestId:   "pr1",
			PullRequestName: "pr1",
			AuthorId:        "u0",
		})
		if err != nil {
			t.Fatalf("CreatePullRequest should succeed, got: %v", err)
		}
		if len(res.PullRequest.AssignedReviewers) != 1 {
			t.Fatalf("AssignedReviewers expected 1, got: %d", len(res.PullRequest.AssignedReviewers))
		}

		_, err = userService.SetIsActive(ctx, entity.SetUserIsActiveDTO{
			UserId:   "u1",
			IsActive: false,
		})
		if err != nil {
			t.Fatalf("SetIsActive should succeed, got: %v", err)
		}

		_, err = prService.ReassignPullRequest(ctx, entity.ReassignPullRequestDTO{
			PullRequestId: "pr1",
			OldReviewerId: res.PullRequest.AssignedReviewers[0],
		})
		if err != nil {
			if !errors.Is(err, errs.ErrNoActiveUsers) {
				t.Fatalf("ReassignPullRequest expected error ErrNoActiveUsers, got: %v", err)
			}
		}
	})
}

func TestMergePullRequest(t *testing.T) {
	t.Run("Merge one time", func(t *testing.T) {
		ctx := setupTest(t)

		users := make([]entity.UserDTO, 10)
		for i := range users {
			users[i] = entity.UserDTO{
				UserId:   fmt.Sprintf("u%d", i),
				Username: fmt.Sprintf("user%d", i),
				IsActive: true,
			}
		}
		_, err := teamService.AddTeam(ctx, entity.TeamDTO{
			TeamName: "team1",
			Members:  users,
		})
		if err != nil {
			t.Fatalf("AddTeam should succeed, got: %v", err)
		}

		_, err = prService.CreatePullRequest(ctx, entity.PullRequestCreateDTO{
			PullRequestId:   "pr1",
			PullRequestName: "pr1",
			AuthorId:        "u0",
		})
		if err != nil {
			t.Fatalf("CreatePullRequest should succeed, got: %v", err)
		}

		res, err := prService.MergePullRequest(ctx, entity.MergePullRequestDTO{
			PullRequestId: "pr1",
		})
		if err != nil {
			t.Fatalf("MergePullRequest should succeed, got: %v", err)
		}
		if res.PullRequest.MergedAt == nil {
			t.Fatal("MergedAt should not be nil")
		}
	})
	t.Run("Merge two times", func(t *testing.T) {
		ctx := setupTest(t)

		users := make([]entity.UserDTO, 10)
		for i := range users {
			users[i] = entity.UserDTO{
				UserId:   fmt.Sprintf("u%d", i),
				Username: fmt.Sprintf("user%d", i),
				IsActive: true,
			}
		}
		_, err := teamService.AddTeam(ctx, entity.TeamDTO{
			TeamName: "team1",
			Members:  users,
		})
		if err != nil {
			t.Fatalf("AddTeam should succeed, got: %v", err)
		}

		_, err = prService.CreatePullRequest(ctx, entity.PullRequestCreateDTO{
			PullRequestId:   "pr1",
			PullRequestName: "pr1",
			AuthorId:        "u0",
		})
		if err != nil {
			t.Fatalf("CreatePullRequest should succeed, got: %v", err)
		}

		res, err := prService.MergePullRequest(ctx, entity.MergePullRequestDTO{
			PullRequestId: "pr1",
		})
		if err != nil {
			t.Fatalf("MergePullRequest should succeed, got: %v", err)
		}
		prevMergedAt := res.PullRequest.MergedAt

		res, err = prService.MergePullRequest(ctx, entity.MergePullRequestDTO{
			PullRequestId: "pr1",
		})
		if err != nil {
			t.Fatalf("MergePullRequest should succeed, got: %v", err)
		}
		if *res.PullRequest.MergedAt != *prevMergedAt {
			t.Fatalf("MergedAt expected be %v, got: %v", *prevMergedAt, *res.PullRequest.MergedAt)
		}
	})
}

func TestAddTeam(t *testing.T) {
	t.Run("Add two teams with same members", func(t *testing.T) {
		ctx := setupTest(t)

		users := make([]entity.UserDTO, 10)
		for i := range users {
			users[i] = entity.UserDTO{
				UserId:   fmt.Sprintf("u%d", i),
				Username: fmt.Sprintf("user%d", i),
				IsActive: true,
			}
		}
		_, err := teamService.AddTeam(ctx, entity.TeamDTO{
			TeamName: "team1",
			Members:  users,
		})
		if err != nil {
			t.Fatalf("AddTeam should succeed, got: %v", err)
		}

		resSecondTime, err := teamService.AddTeam(ctx, entity.TeamDTO{
			TeamName: "team2",
			Members:  users,
		})
		if err != nil {
			t.Fatalf("AddTeam should succeed, got: %v", err)
		}

		res, err := teamService.GetTeam(ctx, "team2")
		if err != nil {
			t.Fatalf("GetTeam should succeed, got: %v", err)
		}

		if !slices.Equal(res.Members, resSecondTime.Team.Members) {
			t.Fatal("Team members in `team2` expected to be same as in resSecondTime")
		}

		res, err = teamService.GetTeam(ctx, "team1")
		if err != nil {
			t.Fatalf("GetTeam should succeed, got: %v", err)
		}
		if len(res.Members) != 0 {
			t.Fatal("Team members in `team1` expected to be empty, got", len(res.Members))
		}
	})
	t.Run("Add two teams with one different and one new member", func(t *testing.T) {
		ctx := setupTest(t)

		users := make([]entity.UserDTO, 2)
		for i := range users {
			users[i] = entity.UserDTO{
				UserId:   fmt.Sprintf("u%d", i),
				Username: fmt.Sprintf("user%d", i),
				IsActive: true,
			}
		}
		_, err := teamService.AddTeam(ctx, entity.TeamDTO{
			TeamName: "team1",
			Members:  users,
		})
		if err != nil {
			t.Fatalf("AddTeam should succeed, got: %v", err)
		}

		users[1] = entity.UserDTO{
			UserId:   "u11",
			Username: "user11",
			IsActive: true,
		}
		_, err = teamService.AddTeam(ctx, entity.TeamDTO{
			TeamName: "team2",
			Members:  users,
		})
		if err != nil {
			t.Fatalf("AddTeam should succeed, got: %v", err)
		}

		res, err := teamService.GetTeam(ctx, "team2")
		if err != nil {
			t.Fatalf("GetTeam should succeed, got: %v", err)
		}

		if !slices.Equal(res.Members, users) {
			t.Fatal("Team members in `team2` expected to be same as in users")
		}

		res, err = teamService.GetTeam(ctx, "team1")
		if err != nil {
			t.Fatalf("GetTeam should succeed, got: %v", err)
		}
		if len(res.Members) != 1 {
			t.Fatal("Team members in `team1` expected len = 1, got", len(res.Members))
		}
	})
}
