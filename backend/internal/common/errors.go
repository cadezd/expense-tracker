package common

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AppError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}

	return e.Message
}

func NewAppError(status int, code, msg string) *AppError {
	return &AppError{Status: status, Code: code, Message: msg}
}

func NewBadRequestError(code, msg string) *AppError {
	return &AppError{Status: http.StatusBadRequest, Code: code, Message: msg}
}

func NewInternalServerError(code, msg string) *AppError {
	return &AppError{Status: http.StatusInternalServerError, Code: code, Message: msg}
}

var (
	ErrUnauthorized = &AppError{Status: http.StatusUnauthorized, Code: "UNAUTHORIZED", Message: "authentication required"}
	InternalError   = &AppError{Status: http.StatusInternalServerError, Code: "INTERNAL_SERVER_ERROR", Message: "unexpected error occurred"}
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		log.Printf("%s %s: %v", c.Request.Method, path, err)

		var appErr *AppError
		if ok := errors.As(err, &appErr); ok {
			Fail(c, appErr.Status, appErr.Code, appErr.Message)
		} else {
			Fail(c, InternalError.Status, InternalError.Code, InternalError.Message)
		}
	}
}
