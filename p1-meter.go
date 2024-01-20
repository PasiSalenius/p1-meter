package main

import (
	"encoding/json"
	"io"
	"log"
	"time"

	"bytes"
	"flag"
	"fmt"

	"net/http"
)

func main() {
	var mode string
	flag.StringVar(&mode, "mode", "victoriametrics", "'victoriametrics' to write to VictoriaMetrics, 'sqlite' to write to local SQLite")

	var host string
	flag.StringVar(&host, "host", "localhost", "hostname of VictoriaMetrics instance")

	var pollInterval int64
	flag.Int64Var(&pollInterval, "interval", 10, "poll interval")

	flag.Parse()

	log.Println("Selected mode:", mode)
	if mode == "sqlite" {
		log.Println("Writing data to local SQLite database (p1-meter.db)")
		log.Println("Selected mode:", mode)
		if err := connectDB(); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("Writing data to VictoriaMetrics at", host)
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

			log.Printf("Loaded P1 Meter reading: ActivePowerW: %f", reading.ActivePowerW)

			if mode == "sqlite" {
				sqliteStoreReading(reading)
			} else if mode == "victoriametrics" {
				writeInfluxDB(reading, host)
			}
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

func writeInfluxDB(r MeterReading, host string) error {
	str := fmt.Sprintf(
		"p1,tag=p1 wifi_strength=%f,total_power_import_kwh=%f,total_power_import_t1_kwh=%f,total_power_export_kwh=%f,total_power_export_t1_kwh=%f,active_power_w=%f,active_power_l1_w=%f,active_power_l2_w=%f,active_power_l3_w=%f,active_voltage_l1_v=%f,active_voltage_l2_v=%f,active_voltage_l3_v=%f,active_current_l1_a=%f,active_current_l2_a=%f,active_current_l3_a=%f",
		r.WifiStrength,
		r.TotalPowerImportKWH,
		r.TotalPowerExportT1KWH,
		r.TotalPowerExportKWH,
		r.TotalPowerExportT1KWH,
		r.ActivePowerW,
		r.ActivePowerL1W,
		r.ActivePowerL2W,
		r.ActivePowerL3W,
		r.ActiveVoltageL1V,
		r.ActiveVoltageL2V,
		r.ActiveVoltageL3V,
		r.ActiveCurrentL1A,
		r.ActiveCurrentL2A,
		r.ActiveCurrentL3A,
	)

	b := []byte(str)

	url := fmt.Sprintf("http://%s:8428/write", host)

	_, err := request("POST", url, b)
	if err != nil {
		return fmt.Errorf("! failed load: %v", err)
	}

	return nil
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
