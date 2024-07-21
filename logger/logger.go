package logger

import (
	"errors"
	"net/http"

	"github.com/ayo-ajayi/bloggy/apperrors"
	"github.com/ayo-ajayi/logger"
	"github.com/gin-gonic/gin"
)

type Logger struct {
	*logger.Logger
}

func NewLogger() *Logger {
	return &Logger{
		Logger: logger.NewLogger(logger.DEBUG, false),
	}
}

func SetLoggerInstance(c *gin.Context, log *Logger) {
	c.Set("logger", log)
}

func getLoggerInstance(c *gin.Context) *Logger {
	logger, exists := c.Get("logger")
	if !exists {
		return nil
	}
	log, ok := logger.(*Logger)
	if !ok {
		return nil
	}
	return log
}

func handleError(c *gin.Context, err error) {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.StatusCode, gin.H{"error": appErr.Sanitize().UserMessage})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}

func LogSuccess(c *gin.Context, message string, data interface{}) {
	response := gin.H{"message": message}
	if data != nil {
		response["data"] = data
	}
	if log := getLoggerInstance(c); log != nil {
		log.Info("success: %s\n", message)
	}
	c.JSON(http.StatusOK, response)
}

func LogError(c *gin.Context, err error) {
	if log := getLoggerInstance(c); log != nil {
		log.Error("error: %v\n", err)
	}
	handleError(c, err)
}
