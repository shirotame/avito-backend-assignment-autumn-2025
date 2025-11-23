package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/shirotame/avito-backend-assignment-autumn-2025/api"
	"github.com/shirotame/avito-backend-assignment-autumn-2025/internal/handler"
	"github.com/shirotame/avito-backend-assignment-autumn-2025/internal/repository/postgres"
	"github.com/shirotame/avito-backend-assignment-autumn-2025/internal/service"
	"github.com/shirotame/avito-backend-assignment-autumn-2025/web"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	logHandler := slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		})

	rootLogger := slog.New(logHandler)

	wd, err := os.Getwd()
	if err != nil {
		rootLogger.Error("failed to Getwd", "err", err)
	}
	err = godotenv.Load(filepath.Join(wd, ".env"), filepath.Join(wd, "local.env"))
	if err != nil {
		rootLogger.Info("error loading .env or local.env", "err", err)
	}

	appPort := 8080
	dbUrl := os.Getenv("DATABASE_URL")

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		rootLogger.Error("Unable to connect to database", "err", err)
		os.Exit(1)
	}
	if err := pool.Ping(ctx); err != nil {
		rootLogger.Error("Unable to ping database", "err", err)
		os.Exit(1)
	}
	rootLogger.Info("Connected to database")

	rootLogger.Info("Setting up repositories")
	userRepo := postgres.NewPostgresUserRepository(rootLogger)
	prRepo := postgres.NewPostgresPullRequestRepository(rootLogger)
	teamRepo := postgres.NewPostgresTeamRepository(rootLogger)

	rootLogger.Info("Setting up services")
	userService := service.NewUserService(rootLogger, pool, userRepo, prRepo)
	prService := service.NewPullRequestService(rootLogger, pool, prRepo, userRepo)
	teamService := service.NewTeamService(rootLogger, pool, userRepo, teamRepo)

	rootLogger.Info("Setting up handlers")
	userHandler := handler.NewUserHandler(rootLogger, userService)
	teamHandler := handler.NewTeamHandler(rootLogger, teamService)
	prHandler := handler.NewPullRequestHandler(rootLogger, prService)

	rootLogger.Info("Setting up router")
	router := chi.NewRouter()

	// Swagger
	router.Handle(
		"/swagger/*",
		http.StripPrefix("/swagger/", http.FileServer(http.FS(web.StaticFiles))),
	)
	router.Handle(
		"/swagger/openapi.yml",
		http.StripPrefix("/swagger/", http.FileServer(http.FS(api.SpecFile))),
	)

	// Routes
	router.Route("/team", func(r chi.Router) {
		r.Post("/add", teamHandler.AddTeam)
		r.Get("/get", teamHandler.GetTeam)
	})

	router.Route("/users", func(r chi.Router) {
		r.Post("/setIsActive", userHandler.SetIsActive)
		r.Get("/getReview", userHandler.GetReview)
	})

	router.Route("/pullRequest", func(r chi.Router) {
		r.Get("/openByReviewers", prHandler.GetOpenPullRequestsByReviewers)
		r.Post("/create", prHandler.CreatePullRequest)
		r.Post("/merge", prHandler.MergePullRequest)
		r.Post("/reassign", prHandler.ReassignPullRequest)
	})

	rootLogger.Info("Starting server", "port", appPort)
	err = http.ListenAndServe(fmt.Sprintf(":%d", appPort), router)
	if err != nil {
		rootLogger.Info("Failed starting server", "port", appPort, "err", err)
		return
	}
}
