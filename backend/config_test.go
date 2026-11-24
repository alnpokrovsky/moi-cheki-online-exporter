package main

import (
	"os"
	"testing"
	"time"
)

// TestLoadConfig_AllSet возвращает валидный Config, когда все переменные заданы.
func TestLoadConfig_AllSet(t *testing.T) {
	// Устанавливаем все переменные окружения
	os.Setenv("REFRESH_TOKEN", "test-refresh-token")
	os.Setenv("FNS_DEVICE_ID", "device-123")
	os.Setenv("FNS_API_URL", "https://api.example.com")
	os.Setenv("USER_AGENT", "test-agent/1.0")
	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("PORT", "8080")
	os.Setenv("RECEIVE_DELAY", "1")
	os.Setenv("DEBUG_MODE", "true")

	defer func() {
		os.Unsetenv("REFRESH_TOKEN")
		os.Unsetenv("FNS_DEVICE_ID")
		os.Unsetenv("FNS_API_URL")
		os.Unsetenv("USER_AGENT")
		os.Unsetenv("MONGO_URI")
		os.Unsetenv("PORT")
		os.Unsetenv("RECEIVE_DELAY")
		os.Unsetenv("DEBUG_MODE")
	}()

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if config.RefreshToken != "test-refresh-token" {
		t.Fatalf("Expected RefreshToken 'test-refresh-token', got '%s'", config.RefreshToken)
	}
	if config.DeviceID != "device-123" {
		t.Fatalf("Expected DeviceID 'device-123', got '%s'", config.DeviceID)
	}
	if config.APIURL != "https://api.example.com" {
		t.Fatalf("Expected APIURL 'https://api.example.com', got '%s'", config.APIURL)
	}
	if config.UserAgent != "test-agent/1.0" {
		t.Fatalf("Expected UserAgent 'test-agent/1.0', got '%s'", config.UserAgent)
	}
	if config.MongoURI != "mongodb://localhost:27017" {
		t.Fatalf("Expected MongoURI 'mongodb://localhost:27017', got '%s'", config.MongoURI)
	}
	if config.Port != "8080" {
		t.Fatalf("Expected Port '8080', got '%s'", config.Port)
	}
	if config.ReceiveDelay != 1*time.Second {
		t.Fatalf("Expected ReceiveDelay 1s, got %v", config.ReceiveDelay)
	}
	if !config.DebugMode {
		t.Fatal("Expected DebugMode to be true")
	}
}

// TestLoadConfig_MissingRefreshToken возвращает ошибку, когда REFRESH_TOKEN отсутствует.
func TestLoadConfig_MissingRefreshToken(t *testing.T) {
	os.Unsetenv("REFRESH_TOKEN")

	defer os.Unsetenv("REFRESH_TOKEN")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("Expected error for missing REFRESH_TOKEN, got nil")
	}
	if err.Error() != "переменная окружения REFRESH_TOKEN не задана" {
		t.Fatalf("Expected error message about REFRESH_TOKEN, got: %v", err)
	}
}

// TestLoadConfig_MissingDeviceID возвращает ошибку, когда FNS_DEVICE_ID отсутствует.
func TestLoadConfig_MissingDeviceID(t *testing.T) {
	os.Setenv("REFRESH_TOKEN", "test-refresh")
	os.Unsetenv("FNS_DEVICE_ID")

	defer func() {
		os.Unsetenv("FNS_DEVICE_ID")
		os.Unsetenv("REFRESH_TOKEN")
	}()

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("Expected error for missing FNS_DEVICE_ID, got nil")
	}
	if err.Error() != "переменная окружения FNS_DEVICE_ID не задана" {
		t.Fatalf("Expected error message about FNS_DEVICE_ID, got: %v", err)
	}
}

// TestLoadConfig_MissingAPIURL возвращает ошибку, когда FNS_API_URL отсутствует.
func TestLoadConfig_MissingAPIURL(t *testing.T) {
	os.Setenv("REFRESH_TOKEN", "test-refresh")
	os.Setenv("FNS_DEVICE_ID", "device-123")
	os.Unsetenv("FNS_API_URL")

	defer func() {
		os.Unsetenv("FNS_API_URL")
		os.Unsetenv("REFRESH_TOKEN")
		os.Unsetenv("FNS_DEVICE_ID")
	}()

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("Expected error for missing FNS_API_URL, got nil")
	}
	if err.Error() != "переменная окружения FNS_API_URL не задана" {
		t.Fatalf("Expected error message about FNS_API_URL, got: %v", err)
	}
}

// TestLoadConfig_MissingUserAgent возвращает ошибку, когда USER_AGENT отсутствует.
func TestLoadConfig_MissingUserAgent(t *testing.T) {
	os.Setenv("REFRESH_TOKEN", "test-refresh")
	os.Setenv("FNS_DEVICE_ID", "device-123")
	os.Setenv("FNS_API_URL", "https://api.example.com")
	os.Unsetenv("USER_AGENT")

	defer func() {
		os.Unsetenv("USER_AGENT")
		os.Unsetenv("REFRESH_TOKEN")
		os.Unsetenv("FNS_DEVICE_ID")
		os.Unsetenv("FNS_API_URL")
	}()

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("Expected error for missing USER_AGENT, got nil")
	}
	if err.Error() != "переменная окружения USER_AGENT не задана" {
		t.Fatalf("Expected error message about USER_AGENT, got: %v", err)
	}
}

// TestLoadConfig_DefaultValues опциональные переменные имеют значения по умолчанию.
func TestLoadConfig_DefaultValues(t *testing.T) {
	// Устанавливаем только обязательные переменные
	os.Setenv("REFRESH_TOKEN", "test-refresh")
	os.Setenv("FNS_DEVICE_ID", "device-123")
	os.Setenv("FNS_API_URL", "https://api.example.com")
	os.Setenv("USER_AGENT", "test-agent/1.0")

	defer func() {
		os.Unsetenv("REFRESH_TOKEN")
		os.Unsetenv("FNS_DEVICE_ID")
		os.Unsetenv("FNS_API_URL")
		os.Unsetenv("USER_AGENT")
	}()

	// Не задаем MONGO_URI, PORT, RECEIVE_DELAY, DEBUG_MODE
	// MONGO_URI и PORT должны получиться по умолчанию
	// RECEIVE_DELAY по умолчанию 0
	// DEBUG_MODE по умолчанию false

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if config.MongoURI != "mongodb://localhost:27017" {
		t.Fatalf("Expected default MongoURI, got '%s'", config.MongoURI)
	}

	if config.Port != "8080" {
		t.Fatalf("Expected default Port '8080', got '%s'", config.Port)
	}

	if config.ReceiveDelay != 0 {
		t.Fatalf("Expected default ReceiveDelay 0, got %v", config.ReceiveDelay)
	}

	if config.DebugMode {
		t.Fatal("Expected default DebugMode false, got true")
	}
}

// TestLoadConfig_OptionalOverridden опциональные переменные могут переопределять значения по умолчанию.
func TestLoadConfig_OptionalOverridden(t *testing.T) {
	os.Setenv("REFRESH_TOKEN", "test-refresh")
	os.Setenv("FNS_DEVICE_ID", "device-123")
	os.Setenv("FNS_API_URL", "https://api.example.com")
	os.Setenv("USER_AGENT", "test-agent/1.0")
	os.Setenv("MONGO_URI", "mongodb://custom:27017")
	os.Setenv("PORT", "9090")
	os.Setenv("RECEIVE_DELAY", "5")
	os.Setenv("DEBUG_MODE", "true")

	defer func() {
		os.Unsetenv("REFRESH_TOKEN")
		os.Unsetenv("FNS_DEVICE_ID")
		os.Unsetenv("FNS_API_URL")
		os.Unsetenv("USER_AGENT")
		os.Unsetenv("MONGO_URI")
		os.Unsetenv("PORT")
		os.Unsetenv("RECEIVE_DELAY")
		os.Unsetenv("DEBUG_MODE")
	}()

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if config.MongoURI != "mongodb://custom:27017" {
		t.Fatalf("Expected custom MongoURI, got '%s'", config.MongoURI)
	}

	if config.Port != "9090" {
		t.Fatalf("Expected custom Port '9090', got '%s'", config.Port)
	}

	if config.ReceiveDelay != 5*time.Second {
		t.Fatalf("Expected ReceiveDelay 5s, got %v", config.ReceiveDelay)
	}

	if !config.DebugMode {
		t.Fatal("Expected DebugMode true, got false")
	}
}