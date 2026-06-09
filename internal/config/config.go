package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

const DefaultConfigPath = "config/deye.yml"

type LoggerConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	Serial         int    `yaml:"serial"`
	ModbusSlaveID  int    `yaml:"modbus_slave_id"`
	ReconnectDelay int    `yaml:"reconnect_delay"`
}

type InfluxConfig struct {
	URL    string `yaml:"url"`
	Token  string `yaml:"token"`
	Org    string `yaml:"org"`
	Bucket string `yaml:"bucket"`
}

type MqttConfig struct {
	Host  string `yaml:"host"`
	Port  int    `yaml:"port"`
	Topic string `yaml:"topic"`
}

type AppConfig struct {
	Logger LoggerConfig `yaml:"logger"`
	Influx InfluxConfig `yaml:"influx"`
	Mqtt   MqttConfig   `yaml:"mqtt"`
}

func defaults() AppConfig {
	return AppConfig{
		Logger: LoggerConfig{
			Host:           "192.168.1.5",
			Port:           8899,
			Serial:         3590091076,
			ModbusSlaveID:  1,
			ReconnectDelay: 15,
		},
		Influx: InfluxConfig{
			URL:    "http://localhost:8086",
			Token:  "deye-token-local",
			Org:    "deye",
			Bucket: "deye",
		},
		Mqtt: MqttConfig{
			Host:  "localhost",
			Port:  1883,
			Topic: "deye/metrics",
		},
	}
}

func Load(path string) (*AppConfig, error) {
	cfg := defaults()

	if path == "" {
		path = DefaultConfigPath
	}

	data, err := os.ReadFile(path)
	if err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse config %s: %w", path, err)
		}
	}

	// Environment overrides
	applyEnvStr(&cfg.Logger.Host, "DEYE_LOGGER_HOST")
	applyEnvInt(&cfg.Logger.Port, "DEYE_LOGGER_PORT")
	applyEnvInt(&cfg.Logger.Serial, "DEYE_LOGGER_SN")
	applyEnvInt(&cfg.Logger.ModbusSlaveID, "DEYE_MODBUS_SLAVE_ID")
	applyEnvInt(&cfg.Logger.ReconnectDelay, "DEYE_RECONNECT_DELAY")

	applyEnvStr(&cfg.Influx.URL, "DEYE_INFLUX_URL")
	applyEnvStr(&cfg.Influx.Token, "DEYE_INFLUX_TOKEN")
	applyEnvStr(&cfg.Influx.Org, "DEYE_INFLUX_ORG")
	applyEnvStr(&cfg.Influx.Bucket, "DEYE_INFLUX_BUCKET")

	applyEnvStr(&cfg.Mqtt.Host, "DEYE_MQTT_HOST")
	applyEnvInt(&cfg.Mqtt.Port, "DEYE_MQTT_PORT")
	applyEnvStr(&cfg.Mqtt.Topic, "DEYE_MQTT_TOPIC")

	return &cfg, nil
}

func applyEnvStr(target *string, key string) {
	if v := os.Getenv(key); v != "" {
		*target = v
	}
}

func applyEnvInt(target *int, key string) {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			*target = n
		}
	}
}
