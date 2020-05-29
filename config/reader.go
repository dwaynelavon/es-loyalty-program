package config

import (
	"errors"
	"os"

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
		return nil, errors.New("Missing firebase config")
	}
	return &firebaseConfigFile, nil
}

// LoadEnvWithPath create a funtion that can be used to
// load env variables from the filesystem when invoked
func LoadEnvWithPath(configPath string) error {
	if err := godotenv.Load(configPath); err != nil {
		return errors.New("No .env file found")
	}
	return nil
}
