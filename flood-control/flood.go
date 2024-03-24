package floodcontrol

import (
	"context"
	"database/sql"
)

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
	query := "select count(*) from flood_record where user_id = $1 and date > (NOW() - $1 * '1 SECOND'::interval)"
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
