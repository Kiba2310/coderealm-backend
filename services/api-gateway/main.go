package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"BootDevClone/shared/env"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var (
	httpAddr           string
	learningModuleAddr string
	testServiceAddr    string
	userServiceAddr    string
	groupServiceAddr   string
	keycloakURL        string
	keycloakRealm      string
	frontendURL        string
	gatewaySecret      string
	pistonURL          string
	jwks               keyfunc.Keyfunc
)

func init() {
	_ = godotenv.Load() // load .env if exists

	httpAddr = env.GetString("HTTP_ADDR", ":8081")
	learningModuleAddr = env.GetString("LEARNING_MODULE_ADDR", "http://localhost:8082")
	testServiceAddr = env.GetString("TEST_SERVICE_ADDR", "http://localhost:8083")
	userServiceAddr = env.GetString("USER_SERVICE_ADDR", "http://localhost:8084")
	groupServiceAddr = env.GetString("GROUP_SERVICE_ADDR", "http://localhost:8085")
	keycloakURL = env.GetString("KEYCLOAK_URL", "https://keycloak.pollak.info")
	keycloakRealm = env.GetString("KEYCLOAK_REALM", "coderealm")
	frontendURL = env.GetString("FRONTEND_URL", "http://localhost:5173")
	gatewaySecret = env.GetString("GATEWAY_SECRET", "super_secret_gateway_key_12345")
	pistonURL = env.GetString("PISTON_URL", "https://pistonapi.pollak.info")

	jwksURL := keycloakURL + "/realms/" + keycloakRealm + "/protocol/openid-connect/certs"
	k, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		log.Fatalf("Failed to create JWKS from resource at the given URL: %v", err)
	}
	jwks = k
}

func newReverseProxy(target string) *httputil.ReverseProxy {
	u, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Invalid proxy target %q: %v", target, err)
	}
	proxy := httputil.NewSingleHostReverseProxy(u)

	// Inject the gateway secret into all proxied requests
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Header.Set("X-Gateway-Secret", gatewaySecret)
	}

	return proxy
}

func newExternalReverseProxy(target string) *httputil.ReverseProxy {
	u, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Invalid proxy target %q: %v", target, err)
	}
	proxy := httputil.NewSingleHostReverseProxy(u)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = u.Host // CRITICAL for external domains!
	}

	return proxy
}

// RealmAccess matches Keycloak structure
type RealmAccess struct {
	Roles []string `json:"roles"`
}

type CustomClaims struct {
	RealmAccess RealmAccess `json:"realm_access"`
	Email       string      `json:"email"`
	jwt.RegisteredClaims
}

func AuthMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			// If no token but we have required roles, block it.
			if len(requiredRoles) > 0 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
				return
			}
			// Optional auth: continue without setting headers
			c.Next()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &CustomClaims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, jwks.Keyfunc)

		if err != nil || !token.Valid {
			if len(requiredRoles) > 0 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			} else {
				// Optional auth: continue without setting headers
				c.Next()
			}
			return
		}

		// Pass extracted info to downstream services
		c.Request.Header.Set("X-User-Id", claims.Subject)
		c.Request.Header.Set("X-User-Email", claims.Email)
		rolesStr := strings.Join(claims.RealmAccess.Roles, ",")
		c.Request.Header.Set("X-User-Roles", rolesStr)

		// Check roles
		if len(requiredRoles) > 0 {
			hasRole := false
			for _, required := range requiredRoles {
				for _, userRole := range claims.RealmAccess.Roles {
					// Map keycloak ADMIN to our required admin check if necessary.
					// We'll just do a direct string comparison.
					if strings.EqualFold(required, userRole) {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}
			if !hasRole {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden - Insufficient permissions"})
				return
			}
		}

		c.Next()
	}
}

func main() {
	log.Println("Starting API Gateway")

	learningProxy := newReverseProxy(learningModuleAddr)
	testProxy := newReverseProxy(testServiceAddr)
	userProxy := newReverseProxy(userServiceAddr)
	groupProxy := newReverseProxy(groupServiceAddr)

	pistonProxy := newExternalReverseProxy(pistonURL)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowOrigin := ""
		origins := strings.Split(frontendURL, ",")

		// Kifejezetten engedélyezzük a localhostot fejlesztési célból
		origins = append(origins, "http://localhost:5173")

		// If no Origin header is provided, fallback to the first allowed origin.
		if len(origins) > 0 {
			allowOrigin = strings.TrimSpace(origins[0])
		}

		if origin != "" {
			for _, o := range origins {
				o = strings.TrimSpace(o)
				if origin == o {
					allowOrigin = origin
					break
				}
			}
		}

		// Ha a kérés preflight, gyakran szigorúbb ellenőrzések vannak.
		// De ezzel a beállítással át kell mennie.

		// Strict CORS
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-User-Id, X-User-Email")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		// Security Headers (Helmet-like)
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// User Service
	r.Any("/auth/*path", func(c *gin.Context) {
		// Leaderboard is public
		if c.Request.URL.Path == "/auth/leaderboard" {
			userProxy.ServeHTTP(c.Writer, c.Request)
			return
		}

		AuthMiddleware()(c)
		if c.IsAborted() {
			return
		}
		userProxy.ServeHTTP(c.Writer, c.Request)
	})

	// Public read access, protected write access
	// We use the AuthMiddleware dynamically based on HTTP Method

	learningModuleHandler := func(c *gin.Context) {
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut || c.Request.Method == http.MethodDelete {
			// /run and /submit endpoints should be accessible by any authenticated user (students too)
			path := c.Request.URL.Path
			if strings.HasSuffix(path, "/run") || strings.HasSuffix(path, "/submit") {
				AuthMiddleware()(c) // just require a valid token, no specific role
			} else {
				AuthMiddleware("admin", "teacher")(c)
			}
			if c.IsAborted() {
				return
			}
		} else {
			// Optional auth for GET requests
			AuthMiddleware()(c)
			if c.IsAborted() {
				return
			}
		}
		learningProxy.ServeHTTP(c.Writer, c.Request)
	}

	testServiceHandler := func(c *gin.Context) {
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut || c.Request.Method == http.MethodDelete {
			AuthMiddleware("admin", "teacher")(c)
			if c.IsAborted() {
				return
			}
		} else {
			// Optional auth for GET requests
			AuthMiddleware()(c)
			if c.IsAborted() {
				return
			}
		}
		testProxy.ServeHTTP(c.Writer, c.Request)
	}

	groupServiceHandler := func(c *gin.Context) {
		AuthMiddleware()(c)
		if c.IsAborted() {
			return
		}
		groupProxy.ServeHTTP(c.Writer, c.Request)
	}

	// Learning Module Service
	r.Any("/languages", learningModuleHandler)
	r.Any("/courses", learningModuleHandler)
	r.Any("/courses/*path", learningModuleHandler)

	// Test Service
	r.Any("/tests", testServiceHandler)
	r.Any("/tests/*path", testServiceHandler)

	// Group Service
	r.Any("/groups", groupServiceHandler)
	r.Any("/groups/*path", groupServiceHandler)

	// Piston API Proxy
	r.Any("/piston/*path", func(c *gin.Context) {
		// Re-write path from /piston/runtimes -> /api/v2/runtimes
		c.Request.URL.Path = "/api/v2" + c.Param("path")
		pistonProxy.ServeHTTP(c.Writer, c.Request)
	})

	log.Printf("Listening on %s", httpAddr)
	if err := r.Run(httpAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
