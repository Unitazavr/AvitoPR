package domain

import (
	"errors"
	"fmt"
)

type ErrorResponse struct {
	ErrorContent ErrorBody `json:"error"`
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf("Code: %s \n"+
		"Message: %s \n", e.ErrorContent.Code, e.ErrorContent.Message)
}

type ErrorBody struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

type ErrorCode string

const (
	ErrCodeTeamExists  ErrorCode = "TEAM_EXISTS"
	ErrCodePRExists    ErrorCode = "PR_EXISTS"
	ErrCodePRMerged    ErrorCode = "PR_MERGED"
	ErrCodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrCodeNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrCodeNotFound    ErrorCode = "NOT_FOUND"
	ErrUnknown         ErrorCode = "UNKNOWN ERROR"
)

var ErrNotFound = errors.New("not found")
