package main

import (
	"context"
	"fmt"
	"task/pkg/configs"
	floodcontrol "task/pkg/flood-control"

	_ "github.com/jackc/pgx/stdlib"
)

func main() {
	config := configs.ParseConfig(configs.ConfName)
	if config == nil {
		return
	}
	db, err := configs.GetPostgres(config.DBconfig)
	if err != nil {
		fmt.Printf("Cant connect to db: %v", err)
		return
	}

	floodControl := floodcontrol.NewFloodController(db,
		config.AppConfig.MaxSecondsPassed,
		config.AppConfig.MaxTotalRecords)

	for i := 1; i < 15; i++ {
		res, err := floodControl.Check(context.Background(), 1)
		fmt.Println(res, err)
	}
}

// FloodControl интерфейс, который нужно реализовать.
// Рекомендуем создать директорию-пакет, в которой будет находиться реализация.
type FloodControl interface {
	// Check возвращает false если достигнут лимит максимально разрешенного
	// кол-ва запросов согласно заданным правилам флуд контроля.
	Check(ctx context.Context, userID int64) (bool, error)
}
