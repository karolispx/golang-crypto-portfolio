package helpers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// Configuration - settings etc
type Configuration struct {
	DBHost        string `json:"DBHost"`
	DBPort        string `json:"DBPort"`
	DBUser        string `json:"DBUser"`
	DBPassword    string `json:"DBPassword"`
	DBName        string `json:"DBName"`
	Port          string `json:"Port"`
	JWTSigningKey string `json:"JWTSigningKey"`
	RestAPIPath   string `json:"RestAPIPath"`
	RestAPIURL    string `json:"RestAPIURL"`
	APPName       string `json:"APPName"`
	CMCAPIKeys    []struct {
		APIKey string `json:"APIKey"`
	} `json:"CMCApiKeys"`
}

// GetConfig - get APP's variables
func GetConfig() Configuration {
	// Extract variable from config.json file
	file, err := os.Open("config.json")

	if err != nil {
		panic(err)
	}

	var Config Configuration

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&Config)

	if err != nil {
		panic(err)
	}

	return Config
}

// InitDB - initialize DB
func InitDB() *sql.DB {
	var Config = GetConfig()
	var err error

	DBInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		Config.DBHost, Config.DBPort, Config.DBUser, Config.DBPassword, Config.DBName,
	)

	DB, err := sql.Open("postgres", DBInfo)

	if err != nil {
		log.Panic(err)
	}

	if err = DB.Ping(); err != nil {
		log.Panic(err)
	}

	return DB
}
