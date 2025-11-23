package database

import (
	"log"
	"time"
)

// MockDB simulates database operations (in-memory)
type MockDB struct {
    PublicKeys map[string]bool      // Public Key presence check (string is the PEM key)
    Sessions   map[string]time.Time // Token expiry (key is the session token)
}

// NewMockDB initializes the mock database
func NewMockDB() *MockDB {
    log.Println("Database initialized (Mocked successfully)")
    return &MockDB{
        PublicKeys: make(map[string]bool),
        Sessions:   make(map[string]time.Time),
    }
}

// PublicKeyExists checks if a key is already in the DB
func (db *MockDB) PublicKeyExists(publicKey string) bool {
    // In a real DB, you'd run a SELECT query
    _, exists := db.PublicKeys[publicKey]
    return exists
}

// SavePublicKey adds a new public key to the DB
func (db *MockDB) SavePublicKey(publicKey string) {
    // In a real DB, you'd run an INSERT/UPSERT query
    db.PublicKeys[publicKey] = true
    log.Printf("DB: Public Key saved/updated (Total keys: %d)", len(db.PublicKeys))
}

// StoreSession associates a token with a key/user and its expiry
func (db *MockDB) StoreSession(publicKey, token string, expiry time.Time) {
    // In a real DB, you'd store the token and expiry against a user ID
    db.Sessions[token] = expiry
    log.Printf("DB: Session stored for key starting with '%s...' (Expires: %s)", publicKey[:20], expiry.Format(time.Kitchen))
}