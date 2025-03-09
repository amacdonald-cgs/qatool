package langfuse

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"crypto/sha256"
	"encoding/hex"

	"github.com/google/uuid"
	_ "github.com/lib/pq"        // Import the PostgreSQL driver
	"golang.org/x/crypto/bcrypt" // Import the bcrypt library
)

const (
	apiKeysTable              = "api_keys"
	orgIDColumn               = "project_id"
	hashedSecretKeyColumn     = "hashed_secret_key"
	fastHashedSecretKeyColumn = "fast_hashed_secret_key"
	displaySecretKeyColumn    = "display_secret_key"
	idColumn                  = "id"
	publicKeyColumn           = "public_key"
)

// Function to connect to a database (replace with your connection details)
func connectToDB(dbHost, dbName, dbUser, dbPassword string) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable", dbHost, dbName, dbUser, dbPassword)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return db, nil
}

// createLangfuseAPIKey function - Modified to directly insert into Langfuse DB
func createLangfuseAPIKey(orgID, langfuseDbHost, langfuseDbName, langfuseDbUser, langfuseDbPassword string) (string, string, error) {
	// 1. Connect to the Langfuse database
	db, err := connectToDB(langfuseDbHost, langfuseDbName, langfuseDbUser, langfuseDbPassword)
	if err != nil {
		return "", "", fmt.Errorf("failed to connect to Langfuse database: %w", err)
	}
	defer db.Close()

	// 2. Generate a new API key (UUID)
	apiKeyID := uuid.New().String()  //The id column
	newAPIKey := uuid.New().String() //The actual api key

	// 3. Hash the API key
	hashedAPIKey, err := hashAPIKey(newAPIKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash API key: %w", err)
	}

	fastHashedAPIKey := fastHashAPIKey(newAPIKey)

	//Creates the display secret key
	displaySecretKey := newAPIKey[:8] + "..." // Example: show the first 8 characters, add "..."

	// 4. Insert the new API key into the Langfuse database
	query := fmt.Sprintf(`
        INSERT INTO %s (%s, created_at, note, %s, %s, %s, %s, %s)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `, apiKeysTable, idColumn, publicKeyColumn, hashedSecretKeyColumn, displaySecretKeyColumn, orgIDColumn, fastHashedSecretKeyColumn)

	_, err = db.Exec(query, apiKeyID, time.Now(), "Created by QA Tool", newAPIKey, hashedAPIKey, displaySecretKey, orgID, fastHashedAPIKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to insert API key into Langfuse database: %w", err)
	}

	log.Println("Successfully generated langfuse api key")
	return newAPIKey, apiKeyID, nil //Returning both api and id
}

func hashAPIKey(apiKey string) (string, error) {
	// Generate a bcrypt hash of the API key
	hashedAPIKey, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("could not generate hash from api key: %w", err)
	}

	// The salt is embedded in the hash by bcrypt
	return string(hashedAPIKey), nil
}

func fastHashAPIKey(apiKey string) string {
	hasher := sha256.New()
	hasher.Write([]byte(apiKey))
	return hex.EncodeToString(hasher.Sum(nil))
}

// getUserLangfuseKeys function to retrieve the public/secret key pair from the qa database.
func getUserLangfuseKeys(userID, orgID, qaDbHost, qaDbName, qaDbUser, qaDbPassword string) (string, string, error) {
	db, err := connectToDB(qaDbHost, qaDbName, qaDbUser, qaDbPassword)
	if err != nil {
		return "", "", fmt.Errorf("failed to connect to QA database: %w", err)
	}
	defer db.Close()

	query := `
        SELECT public_key, id
        FROM user_tokens
        WHERE user_id = $1 AND org_id = $2
    `
	row := db.QueryRow(query, userID, orgID)

	var publicKey, apiKeyID string
	err = row.Scan(&publicKey, &apiKeyID)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("no API key found for user %s and org %s", userID, orgID)
		}
		return "", "", fmt.Errorf("failed to retrieve API key: %w", err)
	}

	return publicKey, apiKeyID, nil
}

// qa Database example
func StoreKeysInQa(apiKey, apiKeyID, orgId, userID, qaDbHost, qaDbName, qaDbUser, qaDbPassword string) error {
	db, err := connectToDB(qaDbHost, qaDbName, qaDbUser, qaDbPassword)
	if err != nil {
		return fmt.Errorf("failed to connect to QA database: %w", err)
	}
	defer db.Close()

	// 4. Insert the new API key into the QA database
	query := `
        INSERT INTO user_tokens (user_id, org_id, public_key, secret_key, created_at)
        VALUES ($1, $2, $3, $4, $5)
    ` // Replace "api_keys" and column names with the actual names in your Langfuse DB
	_, err = db.Exec(query, userID, orgId, apiKey, apiKeyID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert API key into QA database: %w", err)
	}
	log.Println("Successfully stored langfuse api key in qa")
	return nil //Returning both api and hashed.
}
