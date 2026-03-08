package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeOK                 = 0
	CodeBadRequest         = 40001
	CodeInternalError      = 50001
	CodeServiceUnavailable = 50003
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func WriteJSON(c *gin.Context, httpStatus int, code int, message string, data interface{}) {
	c.JSON(httpStatus, Response{Code: code, Message: message, Data: data})
}

func WriteOK(c *gin.Context, data interface{}) {
	WriteJSON(c, http.StatusOK, CodeOK, "ok", data)
}

func WriteBadRequest(c *gin.Context, message string) {
	WriteJSON(c, http.StatusBadRequest, CodeBadRequest, message, nil)
}

func WriteInternalError(c *gin.Context, message string) {
	WriteJSON(c, http.StatusInternalServerError, CodeInternalError, message, nil)
}

func WriteServiceUnavailable(c *gin.Context, message string) {
	WriteJSON(c, http.StatusServiceUnavailable, CodeServiceUnavailable, message, nil)
}
