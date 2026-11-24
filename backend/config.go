package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config содержит конфигурацию приложения, загружаемую из переменных окружения.
type Config struct {
	RefreshToken string
	DeviceID     string
	APIURL       string
	UserAgent    string
	MongoURI     string
	DebugMode    bool
	Port         string
	ReceiveDelay time.Duration
}

const (
	defaultMongoURI = "mongodb://localhost:27017"
	defaultPort     = "8080"
	defaultDelay    = 0 // секунд, задержка отключена по умолчанию
)

// LoadConfig загружает конфигурацию из переменных окружения.
// Обязательные переменные: REFRESH_TOKEN, FNS_DEVICE_ID, FNS_API_URL, USER_AGENT.
// При отсутствии любой обязательной переменной возвращается ошибка.
func LoadConfig() (*Config, error) {
	refreshToken := os.Getenv("REFRESH_TOKEN")
	if refreshToken == "" {
		return nil, fmt.Errorf("переменная окружения REFRESH_TOKEN не задана")
	}

	deviceID := os.Getenv("FNS_DEVICE_ID")
	if deviceID == "" {
		return nil, fmt.Errorf("переменная окружения FNS_DEVICE_ID не задана")
	}

	apiURL := os.Getenv("FNS_API_URL")
	if apiURL == "" {
		return nil, fmt.Errorf("переменная окружения FNS_API_URL не задана")
	}

	userAgent := os.Getenv("USER_AGENT")
	if userAgent == "" {
		return nil, fmt.Errorf("переменная окружения USER_AGENT не задана")
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = defaultMongoURI
	}

	debugMode := os.Getenv("DEBUG_MODE") == "true"

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	receiveDelay := time.Duration(defaultDelay) * time.Second
	if delayStr := os.Getenv("RECEIVE_DELAY"); delayStr != "" {
		if d, err := strconv.Atoi(delayStr); err == nil && d >= 0 {
			receiveDelay = time.Duration(d) * time.Second
		}
	}

	return &Config{
		RefreshToken: refreshToken,
		DeviceID:     deviceID,
		APIURL:       apiURL,
		UserAgent:    userAgent,
		MongoURI:     mongoURI,
		DebugMode:    debugMode,
		Port:         port,
		ReceiveDelay: receiveDelay,
	}, nil
}
