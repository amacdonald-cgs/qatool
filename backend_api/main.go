package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const authServiceURL = "https://rehearsed-dev.teamworkar.com/api/v1/auth"

// Secret key for JWT signing (store securely, e.g., in environment variables)
var jwtSecret = []byte(getJWTSecret())

func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// In production, this should *never* be empty.  For dev, use a default.
		log.Println("WARNING: JWT_SECRET environment variable not set. Using a default (insecure) secret.")
		return "your-default-secret-key" // CHANGE THIS!
	}
	return secret
}

func main() {
	r := gin.Default()

	devMode := os.Getenv("DEV_MODE") == "true"

	if devMode {
		// Allow all origins in development for CORS
		r.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}))
	}

	r.GET("/ping", pingHandler)
	r.POST("/api/login", loginHandler)

	// Protected routes (require authentication)
	authorized := r.Group("/api", authMiddleware())
	authorized.GET("/protected", protectedHandler) // Example protected route

	r.Run(":8000") // Listen and serve on 0.0.0.0:8000
}

func pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

// Custom Claims struct for JWT
type Claims struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

func loginHandler(c *gin.Context) {
	var loginData map[string]string
	if err := c.BindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Invalid request", "errors": err.Error()})
		return
	}

	// Proxy login request to external auth service
	externalAuthURL := fmt.Sprintf("%s/login", authServiceURL)
	jsonData, _ := json.Marshal(loginData)
	req, err := http.NewRequest("POST", externalAuthURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Failed to create auth request", "errors": err.Error()})
		return
	}

	// *** IMPORTANT: Forward relevant headers from the incoming request ***
	req.Header.Set("Content-Type", "application/json") // Always set Content-Type
	req.Header.Set("Accept", "application/json, text/plain, */*")
	// Forward other important headers.  Crucially, forward cookies.
	forwardHeaders := []string{"User-Agent", "Cookie", "Referer", "Origin"}
	for _, header := range forwardHeaders {
		if value := c.GetHeader(header); value != "" {
			req.Header.Set(header, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "External auth service error", "errors": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		// Decode the response from the external auth service
		var externalAuthResponse struct {
			Status  bool                   `json:"status"`
			Message string                 `json:"message"`
			Errors  interface{}            `json:"errors"`
			Data    map[string]interface{} `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&externalAuthResponse); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Failed to decode external auth response", "errors": err.Error()})
			return
		}

		if externalAuthResponse.Status {

			// Extract user data and create JWT (as before)
			userData, ok := externalAuthResponse.Data["user"].(map[string]interface{})
			if !ok {
				userData = map[string]interface{}{}
			}
			userID, _ := userData["id"].(string)
			email, _ := userData["email"].(string)
			roles := []string{}
			if rolesData, ok := userData["roles"].([]interface{}); ok {
				for _, role := range rolesData {
					if roleStr, ok := role.(string); ok {
						roles = append(roles, roleStr)
					}
				}
			}

			claims := Claims{
				UserID: userID,
				Email:  email,
				Roles:  roles,
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    "qa-test-manager",
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
				},
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			signedToken, err := token.SignedString(jwtSecret)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Failed to generate token", "errors": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"status": true, "message": "Login successful", "token": signedToken})

		} else {
			c.JSON(resp.StatusCode, gin.H{"status": false, "message": "Login failed", "external_status_code": resp.StatusCode, "external_response": externalAuthResponse})
		}
	} else {
		var externalAuthErrorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&externalAuthErrorResponse)
		c.JSON(resp.StatusCode, gin.H{"status": false, "message": "Login failed", "external_status_code": resp.StatusCode, "external_response": externalAuthErrorResponse})
	}
}

// Auth Middleware
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": false, "message": "Authorization header required"})
			return
		}

		// Extract token from "Bearer <token>" format
		tokenString := ""
		fmt.Sscanf(authHeader, "Bearer %s", &tokenString)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": false, "message": "Invalid authorization format"})
			return
		}

		// Parse and validate the token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": false, "message": "Invalid token", "errors": err.Error()})
			return
		}

		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			// Token is valid, set claims in context for later use
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("roles", claims.Roles)
			c.Next() // Proceed to the next handler
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": false, "message": "Invalid token claims"})
		}
	}
}

// Example protected handler (requires authentication)
func protectedHandler(c *gin.Context) {
	userID, _ := c.Get("user_id") // Retrieve user ID from context (set by middleware)
	email, _ := c.Get("email")    // Retrieve email from context
	c.JSON(http.StatusOK, gin.H{"message": "Protected resource accessed", "user_id": userID, "email": email})
}
