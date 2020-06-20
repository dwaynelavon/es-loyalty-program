package config

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Reader is a struct containing all the methods used for retrieving config values
type Reader struct{}

// NewReader is a function that creates a new reader
func NewReader() *Reader {
	return &Reader{}
}

// ReadFirebaseCredentialsFileLocation reads the Firebase config file location
func (r *Reader) ReadFirebaseCredentialsFileLocation() (*string, error) {
	firebaseConfigFile, configExists := os.LookupEnv("FIREBASE_CONFIG_FILE")
	if !configExists {
		return nil, errors.New("missing firebase config")
	}
	return &firebaseConfigFile, nil
}

type EventBusBackoffConfig struct {
	InitialIntervalMillis time.Duration
	MaxElapsedMillis      time.Duration
	MaxRetry              int
}

// ReadFirebaseCredentialsFileLocation reads the Firebase config file location
func (r *Reader) EventBusBackoffConfig() (*EventBusBackoffConfig, error) {
	intervalStr, intervalExists := os.LookupEnv("EVENT_BUS_BACKOFF_INITIAL_INTERVAL")
	maxTimeStr, maxTimeExists := os.LookupEnv("EVENT_BUS_BACKOFF_MAX_ELAPSED_TIME")
	maxRetryStr, maxRetryExists := os.LookupEnv("EVENT_BUS_BACKOFF_MAX_RETRY")
	if !intervalExists || !maxTimeExists || !maxRetryExists {
		return nil, errors.New("missing event bus backoff config values")
	}

	interval, errParseInterval := strconv.ParseFloat(intervalStr, 32)
	maxTime, errParseMaxTime := strconv.ParseFloat(maxTimeStr, 32)
	maxRetry, errMaxRetry := strconv.ParseFloat(maxRetryStr, 32)
	if errParseInterval != nil || errParseMaxTime != nil || errMaxRetry != nil {
		return nil, errors.New("unable to parse event bus backoff values")
	}

	return &EventBusBackoffConfig{
		InitialIntervalMillis: time.Duration(interval) * time.Millisecond,
		MaxElapsedMillis:      time.Duration(maxTime) * time.Millisecond,
		MaxRetry:              int(maxRetry),
	}, nil
}

// LoadEnvWithPath create a funtion that can be used to
// load env variables from the filesystem when invoked
func LoadEnvWithPath(configPath string) error {
	if err := godotenv.Load(configPath); err != nil {
		return errors.New("No .env file found")
	}
	return nil
}
