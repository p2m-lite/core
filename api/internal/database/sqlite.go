package database

import (
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

type Store struct {
	DB *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{DB: db}
}

// Models
type Recorder struct {
	Address string `gorm:"primaryKey;collate:nocase"`
	Lat     float64
	Lon     float64
}

type Log struct {
	ID        uint   `gorm:"primaryKey"`
	Recorder  string `gorm:"collate:nocase;index"`
	Ph        int
	Turbidity int
	Timestamp int64 `gorm:"index"`
	Lat       float64
	Lon       float64
}

type ProcessedRecorder struct {
	Recorder    string `gorm:"primaryKey;collate:nocase"`
	ProcessedAt int64
	ExpiresAt   int64 `gorm:"index"`
}

type User struct {
	PublicKey string `gorm:"primaryKey"`
}

type Session struct {
	Token     string `gorm:"primaryKey"`
	PublicKey string `gorm:"index"`
	Expiry    int64
}

func InitDB(filepath string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(filepath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto Migrate
	err = DB.AutoMigrate(&Recorder{}, &Log{}, &ProcessedRecorder{}, &User{}, &Session{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
}

// Helper to get recorder location
func GetRecorderLocation(address string) (float64, float64) {
	var recorder Recorder
	result := DB.First(&recorder, "address = ?", address)
	if result.Error != nil {
		return 0.0, 0.0 // Default if not found
	}
	return recorder.Lat, recorder.Lon
}

// PublicKeyExists checks if a key is already in the DB
func (s *Store) PublicKeyExists(publicKey string) bool {
	var count int64
	s.DB.Model(&User{}).Where("public_key = ?", publicKey).Count(&count)
	return count > 0
}

// SavePublicKey adds a new public key to the DB
func (s *Store) SavePublicKey(publicKey string) {
	user := User{PublicKey: publicKey}
	if err := s.DB.FirstOrCreate(&user).Error; err != nil {
		log.Printf("DB: Error saving public key: %v", err)
	} else {
		log.Println("DB: Public Key saved/updated")
	}
}

// StoreSession associates a token with a key/user and its expiry
func (s *Store) StoreSession(publicKey, token string, expiry time.Time) {
	session := Session{
		Token:     token,
		PublicKey: publicKey,
		Expiry:    expiry.Unix(),
	}
	if err := s.DB.Create(&session).Error; err != nil {
		log.Printf("DB: Error storing session: %v", err)
	} else {
		log.Printf("DB: Session stored for key starting with '%s...' (Expires: %s)", publicKey[:20], expiry.Format(time.Kitchen))
	}
}
