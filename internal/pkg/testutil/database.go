package testutil

import (
	authEntity "github.com/novriyantoAli/moodly/internal/application/auth/entity"
	billEntity "github.com/novriyantoAli/moodly/internal/application/bill/entity"
	"github.com/novriyantoAli/moodly/internal/application/payment/entity"
	securityEntity "github.com/novriyantoAli/moodly/internal/application/security/entity"
	subscribeEntity "github.com/novriyantoAli/moodly/internal/application/subscribe/entity"
	userEntity "github.com/novriyantoAli/moodly/internal/application/user/entity"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	// Auto-migrate all entities
	err = db.AutoMigrate(
		&userEntity.User{},
		&entity.Payment{},
		&subscribeEntity.Subscriber{},
		&billEntity.Bill{},
		&securityEntity.UserPIN{},
		&securityEntity.UserPassword{},
		&securityEntity.UserOAuth{},
		&authEntity.AuthSession{},
		&authEntity.LoginAttempt{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// CleanDB cleans all data from test database
func CleanDB(db *gorm.DB) error {
	// Delete in reverse order of dependencies
	if err := db.Exec("DELETE FROM bills").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM payments").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM subscribers").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM users").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM user_pins").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM user_passwords").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM auth_sessions").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM login_attempts").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM user_oauths").Error; err != nil {
		return err
	}
	return nil
}
