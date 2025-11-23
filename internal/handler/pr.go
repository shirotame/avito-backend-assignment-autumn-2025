package handler

import (
	"log/slog"
	"net/http"
	"prservice/internal/entity"
	"prservice/internal/service"
)

type PullRequestHandler struct {
	logger *slog.Logger
	srv    service.BasePullRequestService
}

func NewPullRequestHandler(
	baseLogger *slog.Logger,
	srv service.BasePullRequestService,
) *PullRequestHandler {
	logger := baseLogger.With("module", "prhandler")
	return &PullRequestHandler{
		logger: logger,
		srv:    srv,
	}
}

func (h *PullRequestHandler) CreatePullRequest(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("CreatePullRequest", "ip", r.RemoteAddr, "user-agent", r.UserAgent())
	data := entity.PullRequestCreateDTO{}
	DecodeDTOFromJson(w, r, &data)

	res, err := h.srv.CreatePullRequest(r.Context(), data)
	if err != nil {
		h.logger.Debug("CreatePullRequest failed", "err", err)
		WriteError(w, err)
		return
	}

	WriteJsonDTO(w, http.StatusCreated, res)
}

func (h *PullRequestHandler) MergePullRequest(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("MergePullRequest", "ip", r.RemoteAddr, "user-agent", r.UserAgent())
	data := entity.MergePullRequestDTO{}
	DecodeDTOFromJson(w, r, &data)

	res, err := h.srv.MergePullRequest(r.Context(), data)
	if err != nil {
		h.logger.Debug("MergePullRequest failed", "err", err)
		WriteError(w, err)
		return
	}

	WriteJsonDTO(w, http.StatusOK, res)
}

func (h *PullRequestHandler) ReassignPullRequest(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("ReassignPullRequest", "ip", r.RemoteAddr, "user-agent", r.UserAgent())
	data := entity.ReassignPullRequestDTO{}
	DecodeDTOFromJson(w, r, &data)

	res, err := h.srv.ReassignPullRequest(r.Context(), data)
	if err != nil {
		h.logger.Debug("ReassignPullRequest failed", "err", err)
		WriteError(w, err)
		return
	}

	WriteJsonDTO(w, http.StatusOK, res)
}
