# Deye Logger Go 🔋

[![Go Version](https://img.shields.io/github/go-mod/go-v/drycool/deye-logger-go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A lightweight and efficient daemon written in Go for monitoring **Deye 6kW** (and compatible) solar inverters. It communicates via Modbus TCP through the Deye WiFi/Ethernet logger and exports metrics to multiple destinations.

## 🚀 Features

- **Real-time Monitoring**: Polls inverter metrics at configurable intervals.
- **Multiple Outlets**:
  - 📊 **CSV**: Local data logging for Excel/Python analysis.
  - 📈 **InfluxDB**: Time-series storage for Grafana dashboards.
  - 📡 **MQTT**: Integration with Home Assistant or Node-RED.
- **Lightweight**: Compiled Go binary with minimal resource footprint.
- **Modbus TCP**: Direct communication with the logger (no cloud required).

## 🛠 Installation

`ash
git clone https://github.com/drycool/deye-logger-go.git
cd deye-logger-go
go mod download
go build -o deye-logger ./cmd/deye-logger
`

## ⚙️ Configuration

Copy the example configuration and edit it with your inverter's IP and logger serial number:

`ash
cp config/deye.example.yml config/deye.yml
`

Edit config/deye.yml:
`yaml
logger:
  host: 192.168.1.5  # Inverter Logger IP
  serial: 3590091076 # Logger Serial Number
  modbus_slave_id: 1
`

## 📖 Usage

Run the logger:
`ash
./deye-logger --config config/deye.yml --interval 30 --mqtt --influx
`

### Flags:
- --config: Path to your YAML config.
- --interval, -i: Polling interval in seconds (default: 30).
- --csv: Path to CSV file (default: data/deye.csv).
- --no-csv: Disable CSV logging.
- --influx: Enable InfluxDB export.
- --mqtt: Enable MQTT export.
- --debug: Enable verbose debug logging.

## 📊 Monitored Metrics

- Phase Voltages & Currents
- Daily/Total Energy Production (kWh)
- Inverter Temperature
- Grid Frequency
- Solar Power (PV1/PV2)
- Battery Status (if applicable)

## 🏷 Keywords
deye, inverter, solar, modbus-tcp, golang, monitoring, home-assistant, influxdb, mqtt, energy-management

---
Developed by [drycool](https://github.com/drycool)
