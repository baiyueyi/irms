package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Details   interface{} `json:"details,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Code:      "OK",
		Message:   "success",
		Data:      data,
		RequestID: requestIDFromContext(c),
	})
}

func Fail(c *gin.Context, httpStatus int, code string, message string, details interface{}) {
	c.JSON(httpStatus, APIResponse{
		Code:      code,
		Message:   message,
		RequestID: requestIDFromContext(c),
		Details:   details,
	})
}

func requestIDFromContext(c *gin.Context) string {
	if v, ok := c.Get("request_id"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
