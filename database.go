package main

import (
	"database/sql"
	"errors"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db *sql.DB
)

func readingCount() (int, error) {
	rows, err := db.Query("SELECT * FROM readings")
	if err != nil {
		return 0, err
	}

	defer rows.Close()

	n := 0
	for rows.Next() {
		n++
	}

	return n, nil
}

func loadReading(t int64) (MeterReading, error) {
	stm, err := db.Prepare("SELECT * FROM groups WHERE timestamp=?")
	if err != nil {
		return MeterReading{}, err
	}

	rows, err := stm.Query(t)
	if err != nil {
		return MeterReading{}, err
	}

	defer rows.Close()

	if !rows.Next() {
		return MeterReading{}, errors.New("no results from query")
	}

	var timestamp int64
	var wifi_ssid string
	var wifi_strength float64
	var meter_model string
	var unique_id string
	var total_power_import_kwh float64
	var total_power_import_t1_kwh float64
	var total_power_export_kwh float64
	var total_power_export_t1_kwh float64
	var active_power_w float64
	var active_power_l1_w float64
	var active_power_l2_w float64
	var active_power_l3_w float64
	var active_voltage_l1_v float64
	var active_voltage_l2_v float64
	var active_voltage_l3_v float64
	var active_current_l1_a float64
	var active_current_l2_a float64
	var active_current_l3_a float64

	err = rows.Scan(&timestamp, &wifi_ssid, &wifi_strength, &meter_model, &unique_id, &total_power_import_kwh, &total_power_import_t1_kwh, &total_power_export_kwh, &total_power_export_t1_kwh, &active_power_w, &active_power_l1_w, &active_power_l2_w, &active_power_l3_w, &active_voltage_l1_v, &active_voltage_l2_v, &active_voltage_l3_v, &active_current_l1_a, &active_current_l2_a, &active_current_l3_a)
	if err != nil {
		return MeterReading{}, err
	}

	reading := MeterReading{
		wifi_ssid,
		wifi_strength,
		meter_model,
		unique_id,
		total_power_import_kwh,
		total_power_import_t1_kwh,
		total_power_export_kwh,
		total_power_export_t1_kwh,
		active_power_w,
		active_power_l1_w,
		active_power_l2_w,
		active_power_l3_w,
		active_voltage_l1_v,
		active_voltage_l2_v,
		active_voltage_l3_v,
		active_current_l1_a,
		active_current_l2_a,
		active_current_l3_a,
	}

	return reading, nil
}

func saveReading(r MeterReading) error {
	stm, err := db.Prepare("INSERT OR REPLACE INTO readings (timestamp, wifi_ssid, wifi_strength, meter_model, unique_id, total_power_import_kwh, total_power_import_t1_kwh, total_power_export_kwh, total_power_export_t1_kwh, active_power_w, active_power_l1_w, active_power_l2_w, active_power_l3_w, active_voltage_l1_v, active_voltage_l2_v, active_voltage_l3_v, active_current_l1_a, active_current_l2_a, active_current_l3_a) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}

	_, err = stm.Exec(time.Now().Unix(), r.WifiSSID, r.WifiStrength, r.MeterModel, r.UniqueID, r.TotalPowerImportKWH, r.TotalPowerImportT1KWH, r.TotalPowerExportKWH, r.TotalPowerExportT1KWH, r.ActivePowerW, r.ActivePowerL1W, r.ActivePowerL2W, r.ActivePowerL3W, r.ActiveVoltageL1V, r.ActiveVoltageL2V, r.ActiveVoltageL3V, r.ActiveCurrentL1A, r.ActiveCurrentL2A, r.ActiveCurrentL3A)
	if err != nil {
		return err
	}

	return nil
}

func deleteReading(t int64) error {
	stm, err := db.Prepare("DELETE FROM readings WHERE timestamp=?")
	if err != nil {
		return err
	}

	_, err = stm.Exec(t)
	if err != nil {
		return err
	}

	return nil
}

//
// Helpers
//

func connectDB() error {
	var err error
	db, err = sql.Open("sqlite3", "p1-meter.db")
	if err != nil {
		log.Fatal(err)
	}

	if err := initDB(); err != nil {
		return err
	}

	log.Println("Connected to SQLite database")

	return nil
}

func initDB() error {
	st, err := db.Prepare("CREATE TABLE IF NOT EXISTS readings (timestamp INTEGER PRIMARY KEY, wifi_ssid TEXT, wifi_strength REAL, meter_model TEXT, unique_id TEXT, total_power_import_kwh REAL, total_power_import_t1_kwh REAL, total_power_export_kwh REAL, total_power_export_t1_kwh REAL, active_power_w REAL, active_power_l1_w REAL, active_power_l2_w REAL, active_power_l3_w REAL, active_voltage_l1_v REAL, active_voltage_l2_v REAL, active_voltage_l3_v REAL, active_current_l1_a REAL, active_current_l2_a REAL, active_current_l3_a REAL)")
	if err != nil {
		return err
	}

	_, err = st.Exec()
	if err != nil {
		return err
	}

	return nil
}
