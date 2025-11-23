package handler

import (
	"log/slog"
	"net/http"

	"github.com/shirotame/avito-backend-assignment-autumn-2025/internal/entity"
	errs "github.com/shirotame/avito-backend-assignment-autumn-2025/internal/errors"
	"github.com/shirotame/avito-backend-assignment-autumn-2025/internal/service"
)

type TeamHandler struct {
	logger *slog.Logger
	srv    service.BaseTeamService
}

func NewTeamHandler(baseLogger *slog.Logger, srv service.BaseTeamService) *TeamHandler {
	logger := baseLogger.With("module", "teamhandler")
	return &TeamHandler{
		logger: logger,
		srv:    srv,
	}
}

func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("AddTeam", "ip", r.RemoteAddr, "user-agent", r.UserAgent())
	data := entity.TeamDTO{}
	DecodeDTOFromJson(w, r, &data)

	res, err := h.srv.AddTeam(r.Context(), data)
	if err != nil {
		h.logger.Debug("AddTeam", "error", err)
		WriteError(w, err)
		return
	}

	WriteJsonDTO(w, http.StatusCreated, res)
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("GetTeam", "ip", r.RemoteAddr, "user-agent", r.UserAgent())
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		h.logger.Debug("GetTeam: query param not found")
		WriteError(w, errs.ErrBaseBadFilter)
		return
	}

	res, err := h.srv.GetTeam(r.Context(), teamName)
	if err != nil {
		h.logger.Debug("GetTeam", "error", err)
		WriteError(w, err)
		return
	}

	WriteJsonDTO(w, http.StatusOK, res)
}
