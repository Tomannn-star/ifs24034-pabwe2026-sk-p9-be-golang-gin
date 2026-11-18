package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_Default(t *testing.T) {
	os.Clearenv()
	cfg := LoadConfig("non_existent_file.env")

	assert.Equal(t, "development", cfg.AppEnv)
	assert.Equal(t, "8081", cfg.AppPort)
	assert.Equal(t, "delcom_blog_secret_key_super_secure_2026_random_string_32chars", cfg.AppKey)
	assert.Equal(t, "admin", cfg.AdminUsername)
	assert.Equal(t, "admin", cfg.AdminPassword)
	assert.Equal(t, "pgsql", cfg.DBConnection)
	assert.Equal(t, "127.0.0.1", cfg.DBHost)
	assert.Equal(t, "5432", cfg.DBPort)
	assert.Equal(t, "pabwe_p9_blog", cfg.DBDatabase)
	assert.Equal(t, "postgres", cfg.DBUsername)
	assert.Equal(t, "postgres", cfg.DBPassword)
	assert.Equal(t, "disable", cfg.DBSSLMode)
	assert.Equal(t, "https://delcom.org/api/v1/search", cfg.SearchEngineEndpoint)
	assert.Equal(t, "https://delcom.org/api/v1", cfg.OpenAIEndpoint)
	assert.Equal(t, "key", cfg.OpenAIKey)
	assert.Equal(t, "nama-model", cfg.OpenAIModel)
}

func TestLoadConfig_CustomEnv(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	os.Setenv("APP_PORT", "9090")
	os.Setenv("APP_KEY", "custom_key")
	os.Setenv("ADMIN_USERNAME", "superadmin")
	os.Setenv("ADMIN_PASSWORD", "secret")
	os.Setenv("DB_CONNECTION", "sqlite")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_DATABASE", "custom_db")
	os.Setenv("DB_USERNAME", "custom_user")
	os.Setenv("DB_PASSWORD", "custom_pass")
	os.Setenv("DB_SSLMODE", "require")
	os.Setenv("SEARCH_ENGINE_ENDPOINT", "https://search.custom.com")
	os.Setenv("OPENAI_ENDPOINT", "https://ai.custom.com")
	os.Setenv("OPENAI_KEY", "custom_ai_key")
	os.Setenv("OPENAI_MODEL", "gpt-4o")
	defer os.Clearenv()

	cfg := LoadConfig()

	assert.Equal(t, "production", cfg.AppEnv)
	assert.Equal(t, "9090", cfg.AppPort)
	assert.Equal(t, "custom_key", cfg.AppKey)
	assert.Equal(t, "superadmin", cfg.AdminUsername)
	assert.Equal(t, "secret", cfg.AdminPassword)
	assert.Equal(t, "sqlite", cfg.DBConnection)
	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, "5433", cfg.DBPort)
	assert.Equal(t, "custom_db", cfg.DBDatabase)
	assert.Equal(t, "custom_user", cfg.DBUsername)
	assert.Equal(t, "custom_pass", cfg.DBPassword)
	assert.Equal(t, "require", cfg.DBSSLMode)
	assert.Equal(t, "https://search.custom.com", cfg.SearchEngineEndpoint)
	assert.Equal(t, "https://ai.custom.com", cfg.OpenAIEndpoint)
	assert.Equal(t, "custom_ai_key", cfg.OpenAIKey)
	assert.Equal(t, "gpt-4o", cfg.OpenAIModel)
}
