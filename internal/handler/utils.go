package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/shirotame/avito-backend-assignment-autumn-2025/internal/entity"
	errs "github.com/shirotame/avito-backend-assignment-autumn-2025/internal/errors"
	"github.com/shirotame/avito-backend-assignment-autumn-2025/internal/errors/codes"
)

func mapToErrorDTO(err error) entity.ErrorDTO {
	if errors.Is(err, errs.ErrBaseNotFound) {
		return entity.ErrorDTO{
			Code:    codes.NotFound,
			Message: "resource not found",
		}
	}
	if errors.Is(err, errs.ErrTeamAlreadyExists) {
		return entity.ErrorDTO{
			Code:    codes.TeamExists,
			Message: "team_name already exists",
		}
	}
	if errors.Is(err, errs.ErrPullRequestAlreadyExists) {
		return entity.ErrorDTO{
			Code:    codes.PullRequestExists,
			Message: "pr_id already exists",
		}
	}
	if errors.Is(err, errs.ErrUserNotAssigned) {
		return entity.ErrorDTO{
			Code:    codes.NotAssigned,
			Message: errs.ErrUserNotAssigned.Error(),
		}
	}
	if errors.Is(err, errs.ErrNoActiveUsers) {
		return entity.ErrorDTO{
			Code:    codes.NoCandidate,
			Message: errs.ErrNoActiveUsers.Error(),
		}
	}
	if errors.Is(err, errs.ErrReassignOnMergedPR) {
		return entity.ErrorDTO{
			Code:    codes.PullRequestMerged,
			Message: errs.ErrReassignOnMergedPR.Error(),
		}
	}
	if errors.Is(err, errs.ErrBaseInternal) {
		return entity.ErrorDTO{
			Code:    codes.Internal,
			Message: "internal error",
		}
	}
	if errors.Is(err, errs.ErrBaseBadFilter) {
		return entity.ErrorDTO{
			Code:    codes.BadRequest,
			Message: "bad filter in request",
		}
	}
	if errors.Is(err, errs.ErrBaseBadRequest) {
		return entity.ErrorDTO{
			Code:    codes.BadRequest,
			Message: "bad request",
		}
	}
	return entity.ErrorDTO{
		Code:    codes.Internal,
		Message: "internal error",
	}
}

func mapToHttpStatus(err error) int {
	if errors.Is(err, errs.ErrBaseNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, errs.ErrTeamAlreadyExists) ||
		errors.Is(err, errs.ErrPullRequestAlreadyExists) ||
		errors.Is(err, errs.ErrUserNotAssigned) ||
		errors.Is(err, errs.ErrNoActiveUsers) ||
		errors.Is(err, errs.ErrReassignOnMergedPR) {
		return http.StatusConflict
	}
	if errors.Is(err, errs.ErrBaseBadFilter) ||
		errors.Is(err, errs.ErrBaseBadRequest) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func WriteError(w http.ResponseWriter, err error) {
	dto := mapToErrorDTO(err)
	code := mapToHttpStatus(err)

	jsonItem, err := json.Marshal(dto)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(jsonItem)
}

func DecodeDTOFromJson(w http.ResponseWriter, r *http.Request, v any) {
	err := json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		WriteError(w, errs.ErrBaseBadRequest)
		return
	}
}

func WriteJsonDTO(w http.ResponseWriter, status int, dto interface{}) {
	jsonData, err := json.Marshal(dto)
	if err != nil {
		WriteError(w, errs.ErrBaseInternal)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonData)
}
