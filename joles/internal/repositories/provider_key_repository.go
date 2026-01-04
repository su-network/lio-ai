package repositories

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"lio-ai/internal/models"
	"os"
	"time"
)

// ProviderKeyRepository handles provider API key operations
type ProviderKeyRepository struct {
	db            *sql.DB
	encryptionKey []byte
}

// NewProviderKeyRepository creates a new provider key repository
func NewProviderKeyRepository(db *sql.DB) *ProviderKeyRepository {
	// Get encryption key from environment or generate one
	encKey := os.Getenv("ENCRYPTION_KEY")
	if encKey == "" {
		// Use a default key (in production, this should be properly managed)
		encKey = "lio-ai-encryption-key-32bytes!"
	}
	
	// Ensure key is 32 bytes for AES-256
	key := []byte(encKey)
	if len(key) < 32 {
		// Pad the key
		padded := make([]byte, 32)
		copy(padded, key)
		key = padded
	} else if len(key) > 32 {
		key = key[:32]
	}
	
	return &ProviderKeyRepository{
		db:            db,
		encryptionKey: key,
	}
}

// encrypt encrypts the API key using AES-256
func (r *ProviderKeyRepository) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(r.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts the API key
func (r *ProviderKeyRepository) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(r.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// Create creates or updates a provider API key for a user
func (r *ProviderKeyRepository) Create(key *models.ProviderAPIKey) error {
	encrypted, err := r.encrypt(key.APIKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt API key: %w", err)
	}

	// Convert models_enabled to JSON
	modelsJSON := "[]"
	if key.ModelsEnabled != "" {
		modelsJSON = key.ModelsEnabled
	}

	query := `
		INSERT INTO provider_api_keys (user_id, provider, api_key_encrypted, models_enabled, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, provider) DO UPDATE SET
			api_key_encrypted = excluded.api_key_encrypted,
			models_enabled = excluded.models_enabled,
			is_active = 1,
			updated_at = excluded.updated_at
	`

	now := time.Now()
	result, err := r.db.Exec(query, key.UserID, key.Provider, encrypted, modelsJSON, true, now, now)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err == nil {
		key.ID = id
	}

	return nil
}

// GetByUserAndProvider gets a specific provider key for a user
func (r *ProviderKeyRepository) GetByUserAndProvider(userID, provider string) (*models.ProviderAPIKey, error) {
	query := `
		SELECT id, user_id, provider, api_key_encrypted, models_enabled, is_active, last_used_at, created_at, updated_at
		FROM provider_api_keys
		WHERE user_id = ? AND provider = ? AND is_active = 1
	`

	key := &models.ProviderAPIKey{}
	var lastUsedAt sql.NullTime
	var modelsEnabled string

	err := r.db.QueryRow(query, userID, provider).Scan(
		&key.ID, &key.UserID, &key.Provider, &key.APIKeyEncrypted,
		&modelsEnabled, &key.IsActive, &lastUsedAt, &key.CreatedAt, &key.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if lastUsedAt.Valid {
		key.LastUsedAt = &lastUsedAt.Time
	}
	key.ModelsEnabled = modelsEnabled

	// Decrypt the API key
	decrypted, err := r.decrypt(key.APIKeyEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt API key: %w", err)
	}
	key.APIKey = decrypted

	return key, nil
}

// GetAllByUser gets all provider keys for a user
func (r *ProviderKeyRepository) GetAllByUser(userID string) ([]*models.ProviderAPIKeyResponse, error) {
	query := `
		SELECT id, provider, models_enabled, is_active, last_used_at, created_at,
		       CASE WHEN api_key_encrypted IS NOT NULL AND api_key_encrypted != '' THEN 1 ELSE 0 END as has_key
		FROM provider_api_keys
		WHERE user_id = ? AND is_active = 1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*models.ProviderAPIKeyResponse
	for rows.Next() {
		key := &models.ProviderAPIKeyResponse{}
		var lastUsedAt sql.NullTime
		var modelsEnabled string
		var hasKey int

		err := rows.Scan(
			&key.ID, &key.Provider, &modelsEnabled, &key.IsActive,
			&lastUsedAt, &key.CreatedAt,
			&hasKey,
		)
		if err != nil {
			return nil, err
		}

		if lastUsedAt.Valid {
			key.LastUsedAt = &lastUsedAt.Time
		}
		key.HasKey = hasKey == 1

		// Parse models_enabled JSON
		if modelsEnabled != "" && modelsEnabled != "[]" {
			json.Unmarshal([]byte(modelsEnabled), &key.ModelsEnabled)
		} else {
			key.ModelsEnabled = []string{}
		}

		keys = append(keys, key)
	}

	return keys, nil
}

// GetAllByUserIncludingInactive gets all provider keys for a user, including inactive ones
func (r *ProviderKeyRepository) GetAllByUserIncludingInactive(userID string) ([]*models.ProviderAPIKeyResponse, error) {
	query := `
		SELECT id, provider, models_enabled, is_active, last_used_at, created_at,
		       CASE WHEN api_key_encrypted IS NOT NULL AND api_key_encrypted != '' THEN 1 ELSE 0 END as has_key
		FROM provider_api_keys
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*models.ProviderAPIKeyResponse
	for rows.Next() {
		key := &models.ProviderAPIKeyResponse{}
		var lastUsedAt sql.NullTime
		var modelsEnabled string
		var hasKey int

		err := rows.Scan(
			&key.ID, &key.Provider, &modelsEnabled, &key.IsActive,
			&lastUsedAt, &key.CreatedAt,
			&hasKey,
		)
		if err != nil {
			return nil, err
		}

		if lastUsedAt.Valid {
			key.LastUsedAt = &lastUsedAt.Time
		}
		key.HasKey = hasKey == 1

		// Parse models_enabled JSON
		if modelsEnabled != "" && modelsEnabled != "[]" {
			json.Unmarshal([]byte(modelsEnabled), &key.ModelsEnabled)
		} else {
			key.ModelsEnabled = []string{}
		}

		keys = append(keys, key)
	}

	return keys, nil
}

// Delete soft deletes a provider API key (sets is_active = 0)
func (r *ProviderKeyRepository) Delete(userID, provider string) error {
	query := `UPDATE provider_api_keys SET is_active = 0, updated_at = ? WHERE user_id = ? AND provider = ?`
	_, err := r.db.Exec(query, time.Now(), userID, provider)
	return err
}

// HardDelete permanently deletes a provider API key from the database
func (r *ProviderKeyRepository) HardDelete(userID, provider string) error {
	query := `DELETE FROM provider_api_keys WHERE user_id = ? AND provider = ?`
	_, err := r.db.Exec(query, userID, provider)
	return err
}

// Restore reactivates a soft-deleted provider API key
func (r *ProviderKeyRepository) Restore(userID, provider string) error {
	query := `UPDATE provider_api_keys SET is_active = 1, updated_at = ? WHERE user_id = ? AND provider = ?`
	_, err := r.db.Exec(query, time.Now(), userID, provider)
	return err
}

// UpdateLastUsed updates the last_used_at timestamp
func (r *ProviderKeyRepository) UpdateLastUsed(userID, provider string) error {
	query := `UPDATE provider_api_keys SET last_used_at = ? WHERE user_id = ? AND provider = ?`
	_, err := r.db.Exec(query, time.Now(), userID, provider)
	return err
}
