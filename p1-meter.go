package main

import (
	"encoding/json"
	"io"
	"log"
	"time"

	"bytes"
	"fmt"

	"net/http"
)

const (
	pollInterval = 1
)

func main() {
	if err := connectDB(); err != nil {
		log.Fatal(err)
	}

	interval := time.Duration(pollInterval) * time.Second
	ticker := time.NewTicker(interval)

	done := make(chan bool)

	go func(done chan bool) {
		for range ticker.C {
			reading, err := getReading()
			if err != nil {
				log.Printf("Failed to load P1 Meter reading: %v", err)
				continue
			}

			log.Println("Loaded P1 Meter reading")
			log.Printf("ActivePowerW: %f", reading.ActivePowerW)

			saveReading(reading)
		}

	}(done)

	log.Printf("Polling P1 Meter with %d seconds interval", pollInterval)

	<-done
}

func getReading() (MeterReading, error) {
	body, err := request("GET", "http://10.0.0.29/api/v1/data", nil)
	if err != nil {
		return MeterReading{}, fmt.Errorf("! failed load: %v", err)
	}

	var reading MeterReading
	if err := json.Unmarshal(body, &reading); err != nil {
		return MeterReading{}, fmt.Errorf("! could not unmarshal json from response: %v", err)
	}

	return reading, nil
}

func request(method, URL string, body []byte) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, URL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("! could not create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("! could not load response: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("! response status code not 200: %s", resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("! could not read response body: %v", err)
	}

	return b, nil
}

// MeterReading JSON type
type MeterReading struct {
	WifiStrength          float64 `json:"wifi_strength"`
	TotalPowerImportKWH   float64 `json:"total_power_import_kwh"`
	TotalPowerImportT1KWH float64 `json:"total_power_import_t1_kwh"`
	TotalPowerExportKWH   float64 `json:"total_power_export_kwh"`
	TotalPowerExportT1KWH float64 `json:"total_power_export_t1_kwh"`
	ActivePowerW          float64 `json:"active_power_w"`
	ActivePowerL1W        float64 `json:"active_power_l1_w"`
	ActivePowerL2W        float64 `json:"active_power_l2_w"`
	ActivePowerL3W        float64 `json:"active_power_l3_w"`
	ActiveVoltageL1V      float64 `json:"active_voltage_l1_v"`
	ActiveVoltageL2V      float64 `json:"active_voltage_l2_v"`
	ActiveVoltageL3V      float64 `json:"active_voltage_l3_v"`
	ActiveCurrentL1A      float64 `json:"active_current_l1_a"`
	ActiveCurrentL2A      float64 `json:"active_current_l2_a"`
	ActiveCurrentL3A      float64 `json:"active_current_l3_a"`
}
