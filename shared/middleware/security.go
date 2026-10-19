package middleware

import (
	"net/http"

	"BootDevClone/shared/env"

	"github.com/gin-gonic/gin"
)

// GatewayAuthMiddleware ensures that the request came from the API Gateway
// by verifying the X-Gateway-Secret header.
func GatewayAuthMiddleware() gin.HandlerFunc {
	expectedSecret := env.GetString("GATEWAY_SECRET", "super_secret_gateway_key_12345")
	
	return func(c *gin.Context) {
		// Allow health checks through directly (important for orchestrators like Portainer/K8s)
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		secret := c.GetHeader("X-Gateway-Secret")
		if secret == "" || secret != expectedSecret {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Forbidden: Request must pass through API Gateway",
			})
			return
		}

		c.Next()
	}
}
