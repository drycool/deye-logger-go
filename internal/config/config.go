package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

const DefaultConfigPath = "config/deye.yml"

// LoggerConfig holds connection settings for the Deye WiFi logger.
type LoggerConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	Serial         int    `yaml:"serial"`   // SN дата-логгера (не инвертора!)
	ModbusSlaveID  int    `yaml:"modbus_slave_id"`
	ReconnectDelay int    `yaml:"reconnect_delay"`
}

// InfluxConfig holds InfluxDB v2 write settings.
type InfluxConfig struct {
	URL    string `yaml:"url"`
	Token  string `yaml:"token"`
	Org    string `yaml:"org"`
	Bucket string `yaml:"bucket"`
}

// MqttConfig holds MQTT publish settings.
type MqttConfig struct {
	Host  string `yaml:"host"`
	Port  int    `yaml:"port"`
	Topic string `yaml:"topic"`
}

// AppConfig is the complete application configuration.
type AppConfig struct {
	Logger LoggerConfig `yaml:"logger"`
	Influx InfluxConfig `yaml:"influx"`
	Mqtt   MqttConfig   `yaml:"mqtt"`
}

func defaults() AppConfig {
	return AppConfig{
		Logger: LoggerConfig{
			Port:           8899,
			ModbusSlaveID:  1,
			ReconnectDelay: 15,
		},
		Influx: InfluxConfig{
			URL:    "http://localhost:8086",
			Org:    "deye",
			Bucket: "deye",
		},
		Mqtt: MqttConfig{
			Port:  1883,
			Topic: "deye/metrics",
		},
	}
}

// Validate checks that required fields are set and returns all errors at once.
func (c *AppConfig) Validate() error {
	var errs []string

	if c.Logger.Host == "" {
		errs = append(errs, "logger.host is required")
	}
	if c.Logger.Port == 0 {
		errs = append(errs, "logger.port is required")
	}
	if c.Logger.Serial == 0 {
		errs = append(errs, "logger.serial is required")
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed:\n  - %s", join(errs, "\n  - "))
	}
	return nil
}

// Load reads configuration from a YAML file and applies environment variable overrides.
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

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

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

func join(strs []string, sep string) string {
	out := ""
	for i, s := range strs {
		if i > 0 {
			out += sep
		}
		out += s
	}
	return out
}
