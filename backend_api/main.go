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
)

const authServiceURL = "https://rehearsed-dev.teamworkar.com/api/v1/auth" // External auth service URL

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
	r.GET("/api/token", tokenHandler)

	r.Run(":8000") // Listen and serve on 0.0.0.0:8000
}

func pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
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
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain, */*") // Mimic browser headers

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "External auth service error", "errors": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		// Forward cookies from external auth service to frontend
		for _, cookie := range resp.Cookies() {
			c.SetCookie(cookie.Name, cookie.Value, cookie.MaxAge, cookie.Path, cookie.Domain, cookie.Secure, cookie.HttpOnly)
		}

		// You might want to parse the body and forward relevant parts, or just a success status
		var externalAuthResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&externalAuthResponse); err != nil {
			log.Println("Error decoding external auth response:", err) // Log error, but still consider login success from status code
		}

		c.JSON(http.StatusOK, gin.H{"status": true, "message": "Login successful", "external_response": externalAuthResponse})
	} else {
		var externalAuthErrorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&externalAuthErrorResponse) // Try to decode error body
		c.JSON(resp.StatusCode, gin.H{"status": false, "message": "Login failed", "external_status_code": resp.StatusCode, "external_response": externalAuthErrorResponse})
	}
}

func tokenHandler(c *gin.Context) {
	// Proxy token validation request to external auth service
	externalAuthURL := fmt.Sprintf("%s/token", authServiceURL)
	req, err := http.NewRequest("GET", externalAuthURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Failed to create token request", "errors": err.Error()})
		return
	}
	req.Header.Set("Accept", "application/json, text/plain, */*") // Mimic browser headers

	// Forward cookies from the incoming request to the external auth service
	for _, cookie := range c.Request.Cookies() {
		req.AddCookie(cookie)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "External token service error", "errors": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var externalAuthResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&externalAuthResponse); err != nil {
			log.Println("Error decoding external token response:", err) // Log error, but still consider token valid from status code
		}
		c.JSON(http.StatusOK, gin.H{"status": true, "message": "Token valid", "external_response": externalAuthResponse})
	} else {
		c.JSON(resp.StatusCode, gin.H{"status": false, "message": "Token invalid", "external_status_code": resp.StatusCode})
	}
}
