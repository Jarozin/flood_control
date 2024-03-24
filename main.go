package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"

	_ "github.com/jackc/pgx/stdlib"
)

const confName = "conf.json"
const maxConnectionCount = 10

func main() {
	config := ParseConfig(confName)
	if config == nil {
		return
	}
	db, err := getPostgres(config.DBconfig, maxConnectionCount)
	if err != nil {
		fmt.Printf("Cant connect to db: %v", err)
		return
	}

	floodControl := NewFloodController(db,
		config.AppConfig.MaxSecondsPassed,
		config.AppConfig.MaxTotalRecords)

	for i := 0; i < 5; i++ {
		res, err := floodControl.Check(context.Background(), 1)
		fmt.Println(res, err)
	}

	return
}

func getPostgres(config *DBconfig, maxConnectionCount int) (*sql.DB, error) {
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

	db.SetMaxOpenConns(maxConnectionCount)
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

// FloodControl интерфейс, который нужно реализовать.
// Рекомендуем создать директорию-пакет, в которой будет находиться реализация.
type FloodControl interface {
	// Check возвращает false если достигнут лимит максимально разрешенного
	// кол-ва запросов согласно заданным правилам флуд контроля.
	Check(ctx context.Context, userID int64) (bool, error)
}

type DBconfig struct {
	User     string `json:"user"`
	Dbname   string `json:"dbname"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Sslmode  string `json:"sslmode"`
}

type AppConfig struct {
	MaxSecondsPassed int `json:"maxSecondsPassed"`
	MaxTotalRecords  int `json:"maxTotalRecords"`
}

type Config struct {
	DBconfig  *DBconfig  `json:"DB"`
	AppConfig *AppConfig `json:"App"`
}

type FloodController struct {
	DB               *sql.DB
	maxSecondsPassed int
	maxTotalRecords  int
}

func (controller *FloodController) Check(ctx context.Context, userID int64) (bool, error) {
	err := controller.deleteOld()
	if err != nil {
		return false, err
	}

	err = controller.addRecord(userID)
	if err != nil {
		return false, err
	}

	var total int
	query := "select count(*) from flood_record where user_id = $1"
	err = controller.DB.QueryRow(query, userID).Scan(&total)
	if err != nil {
		return false, err
	}

	if total > controller.maxTotalRecords {
		return false, nil
	}
	return true, nil
}

func (controller *FloodController) deleteOld() error {
	query := `delete from flood_record where date < (NOW() - $1 * '1 SECOND'::interval)`
	err := controller.DB.QueryRow(query, controller.maxSecondsPassed).Scan()
	if err == sql.ErrNoRows {
		err = nil
	}
	if err != nil {
		return err
	}
	return nil
}

func (controller *FloodController) addRecord(userID int64) error {
	query := `insert into flood_record ("user_id") values ($1::integer)`
	err := controller.DB.QueryRow(query, userID).Scan()
	if err == sql.ErrNoRows {
		err = nil
	}
	if err != nil {
		return err
	}
	return nil
}

func NewFloodController(DB *sql.DB, maxSecondsPassed, maxTotalRecords int) *FloodController {
	return &FloodController{
		DB:               DB,
		maxSecondsPassed: maxSecondsPassed,
		maxTotalRecords:  maxTotalRecords,
	}
}
