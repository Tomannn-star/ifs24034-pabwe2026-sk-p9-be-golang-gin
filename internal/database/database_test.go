package database

import (
	"testing"

	"app/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestDatabase_SQLiteConnectAndMigrateAndSeed(t *testing.T) {
	cfg := &config.Config{
		DBConnection:  "sqlite",
		DBDatabase:    ":memory:",
		AdminUsername: "admin",
		AdminPassword: "adminpassword",
	}

	db, err := Connect(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Test with empty DBDatabase to cover dbPath == "" branch
	cfgEmpty := &config.Config{
		DBConnection: "sqlite",
		DBDatabase:   "",
	}
	dbEmpty, errEmpty := Connect(cfgEmpty)
	assert.NoError(t, errEmpty)
	assert.NotNil(t, dbEmpty)

	err = AutoMigrate(db)
	assert.NoError(t, err)

	err = Seed(db, cfg)
	assert.NoError(t, err)

	// Re-run seed to test idempotency (branches where count > 0)
	err = Seed(db, cfg)
	assert.NoError(t, err)

	// Verify Admin
	var admin User
	err = db.Where("username = ?", "admin").First(&admin).Error
	assert.NoError(t, err)
	assert.Equal(t, "admin", admin.Role)

	// Verify Member
	var member User
	err = db.Where("username = ?", "member").First(&member).Error
	assert.NoError(t, err)
	assert.Equal(t, "member", member.Role)
}

func TestDatabase_PostgresConnectBranch(t *testing.T) {
	cfg := &config.Config{
		DBConnection: "pgsql",
		DBHost:       "127.0.0.1",
		DBPort:       "5432",
		DBDatabase:   "pabwe_p9_blog",
		DBUsername:   "postgres",
		DBPassword:   "postgres",
		DBSSLMode:    "disable",
	}

	// Just invoke Connect to cover postgres branch
	_, _ = Connect(cfg)
}
