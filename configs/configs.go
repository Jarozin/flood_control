package configs

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const ConfName = "conf.json"

type DBconfig struct {
	User               string `json:"user"`
	Dbname             string `json:"dbname"`
	Password           string `json:"password"`
	Host               string `json:"host"`
	Port               int    `json:"port"`
	Sslmode            string `json:"sslmode"`
	MaxConnectionCount int    `json:"connCount"`
}

type AppConfig struct {
	MaxSecondsPassed int `json:"maxSecondsPassed"`
	MaxTotalRecords  int `json:"maxTotalRecords"`
}

type Config struct {
	DBconfig  *DBconfig  `json:"DB"`
	AppConfig *AppConfig `json:"App"`
}

func GetPostgres(config *DBconfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%d sslmode=%s",
		config.User, config.Dbname, config.Password,
		config.Host, config.Port, config.Sslmode)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		fmt.Print(config)
		return nil, err
	}

	db.SetMaxOpenConns(config.MaxConnectionCount)
	return db, nil
}

func ParseConfig(filename string) *Config {
	jsonFile, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Couldnt open the config: %v", err)
		return nil
	}

	defer jsonFile.Close()

	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Printf("Couldnt read json file: %v", err)
		return nil
	}

	config := Config{}
	if err := json.Unmarshal(jsonData, &config); err != nil {
		fmt.Printf("Couldnt unmarshal json file: %v", err)
		return nil
	}

	return &config
}
