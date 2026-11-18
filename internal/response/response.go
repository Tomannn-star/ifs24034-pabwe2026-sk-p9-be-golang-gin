package response

import (
	"github.com/gin-gonic/gin"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, code int, message string, data ...interface{}) {
	res := Response{
		Status:  "success",
		Message: message,
	}
	if len(data) > 0 && data[0] != nil {
		res.Data = data[0]
	}
	c.JSON(code, res)
}

func Fail(c *gin.Context, code int, message string, data ...interface{}) {
	res := Response{
		Status:  "fail",
		Message: message,
	}
	if len(data) > 0 && data[0] != nil {
		res.Data = data[0]
	}
	c.JSON(code, res)
}

func Error(c *gin.Context, code int, message string, data ...interface{}) {
	res := Response{
		Status:  "error",
		Message: message,
	}
	if len(data) > 0 && data[0] != nil {
		res.Data = data[0]
	}
	c.JSON(code, res)
}

func AbortFail(c *gin.Context, code int, message string, data ...interface{}) {
	res := Response{
		Status:  "fail",
		Message: message,
	}
	if len(data) > 0 && data[0] != nil {
		res.Data = data[0]
	}
	c.AbortWithStatusJSON(code, res)
}

func AbortError(c *gin.Context, code int, message string, data ...interface{}) {
	res := Response{
		Status:  "error",
		Message: message,
	}
	if len(data) > 0 && data[0] != nil {
		res.Data = data[0]
	}
	c.AbortWithStatusJSON(code, res)
}
