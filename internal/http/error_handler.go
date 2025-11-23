package http

import (
	"errors"
	"github.com/Unitazavr/AvitoPR/internal/domain"
	"net/http"
)

// HandleError для определения статуса ответа
func HandleError(err error) int {
	var errResp *domain.ErrorResponse
	if errors.As(err, &errResp) {
		switch errResp.ErrorContent.Code {
		case domain.ErrCodeTeamExists:
			return http.StatusBadRequest
		case domain.ErrCodePRExists:
			return http.StatusConflict
		case domain.ErrCodePRMerged:
			return http.StatusConflict
		case domain.ErrCodeNotAssigned:
			return http.StatusConflict
		case domain.ErrCodeNoCandidate:
			return http.StatusConflict
		case domain.ErrCodeNotFound:
			return http.StatusNotFound
		default:
			return http.StatusInternalServerError
		}
	}

	return http.StatusInternalServerError
}
