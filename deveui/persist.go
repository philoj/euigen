package deveui

import (
	"fmt"
	"github.com/tidwall/buntdb"
	"strings"
)

func checkForPreviousRun(db *buntdb.DB) ([]string, error) {
	var ids []string
	err := db.View(func(tx *buntdb.Tx) error {
		lastRunSuccess, err := tx.Get("lastRunSuccess")
		if err != nil {
			return err
		}
		if lastRunSuccess == "true" {
			return nil
		}
		lastRunResults, err := tx.Get("lastRunResults")
		if err != nil {
			return err
		}
		if lastRunResults != "" {
			ids = strings.Split(lastRunResults, ",")
		}
		return nil
	})
	if err == buntdb.ErrNotFound {
		err = nil
	}
	return ids, err
}

func discardPreviousRun(db *buntdb.DB) error {
	return db.Update(func(tx *buntdb.Tx) error {
		return tx.DeleteAll()
	})
}

func saveCurrentRun(db *buntdb.DB, success bool, ids []string) error {
	err := db.Update(func(tx *buntdb.Tx) error {
		if _, _, err := tx.Set("lastRunSuccess", fmt.Sprint(success), nil); err != nil {
			return err
		}
		if _, _, err := tx.Set("lastRunResults", strings.Join(ids, ","), nil); err != nil {
			return err
		}
		return nil
	})
	return err
}
