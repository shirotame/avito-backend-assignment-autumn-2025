package handler

import (
	"log/slog"
	"net/http"
	"prservice/internal/entity"
	errs "prservice/internal/errors"
	"prservice/internal/service"
)

type UserHandler struct {
	logger *slog.Logger
	srv    service.BaseUserService
}

func NewUserHandler(baseLogger *slog.Logger, srv service.BaseUserService) *UserHandler {
	logger := baseLogger.With("module", "userhandler")
	return &UserHandler{
		logger: logger,
		srv:    srv,
	}
}

func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("SetIsActive", "ip", r.RemoteAddr, "user-agent", r.UserAgent())
	data := entity.SetUserIsActiveDTO{}
	DecodeDTOFromJson(w, r, &data)

	res, err := h.srv.SetIsActive(r.Context(), data)
	if err != nil {
		h.logger.Debug("GetReview", "err", err)
		WriteError(w, err)
		return
	}

	WriteJsonDTO(w, http.StatusOK, res)
}

func (h *UserHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("GetReview", "ip", r.RemoteAddr, "user-agent", r.UserAgent())
	userId := r.URL.Query().Get("user_id")
	if userId == "" {
		h.logger.Debug("GetReview: query param not found")
		WriteError(w, errs.ErrBaseBadFilter)
		return
	}

	res, err := h.srv.GetReview(r.Context(), userId)
	if err != nil {
		h.logger.Debug("GetReview", "err", err)
		WriteError(w, err)
		return
	}

	WriteJsonDTO(w, http.StatusOK, res)
}
