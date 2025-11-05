package middleware

import (
	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.RecoveryWithWriter(nil)
}
