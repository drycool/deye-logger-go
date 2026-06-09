package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/drycool/deye-logger-go/internal/config"
	"github.com/drycool/deye-logger-go/internal/logger"
	"github.com/drycool/deye-logger-go/internal/modbus"
	"github.com/drycool/deye-logger-go/internal/registers"
	"github.com/drycool/deye-logger-go/internal/writer"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var unitsRU = map[string]string{
	"V": "В", "A": "А", "W": "Вт", "%": "%",
	"C": "°C", "Hz": "Гц", "kWh": "кВт·ч", "": "",
}

func init() {
	// No-op on non-Windows; build-tag files handle encoding
}

func main() {
	var (
		cfgPath           string
		interval          int
		csvPath           string
		noCSV             bool
		quiet             bool
		enableInflux      bool
		enableMQTT        bool
		includeUnverified bool
		debug             bool
	)

	rootCmd := &cobra.Command{
		Use:   "deye-logger",
		Short: "Deye 6kW inverter metrics daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Init(debug)
			log := logger.Get("main")

			cfg, err := config.Load(cfgPath)
			if err != nil {
				return fmt.Errorf("config: %w", err)
			}

			regs := registers.Active(includeUnverified)
			names := registers.Names(includeUnverified)

			// CSV
			var csvW *writer.CSVWriter
			if !noCSV && csvPath != "" {
				csvW, err = writer.NewCSV(csvPath, includeUnverified)
				if err != nil {
					return fmt.Errorf("csv: %w", err)
				}
				defer csvW.Close()
			}

			// InfluxDB
			var influxW *writer.InfluxWriter
			if enableInflux {
				influxW = writer.NewInflux(cfg.Influx.URL, cfg.Influx.Token, cfg.Influx.Org, cfg.Influx.Bucket)
				defer influxW.Close()
			}

			// MQTT
			var mqttW *writer.MqttWriter
			if enableMQTT {
				mqttW, err = writer.NewMqtt(cfg.Mqtt.Host, cfg.Mqtt.Port, cfg.Mqtt.Topic)
				if err != nil {
					log.Warn("MQTT disabled", "err", err)
				} else {
					defer mqttW.Close()
				}
			}

			client := modbus.New(cfg.Logger.Host, cfg.Logger.Port, cfg.Logger.ModbusSlaveID)
			defer client.Disconnect()

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			ticker := time.NewTicker(time.Duration(interval) * time.Second)
			defer ticker.Stop()

			log.Info("started", "interval", interval, "registers", len(regs))

			for {
				if err := client.Connect(); err != nil {
					log.Error("connect", "err", err)
					select {
					case <-ctx.Done():
						return nil
					case <-time.After(time.Duration(cfg.Logger.ReconnectDelay) * time.Second):
						continue
					}
				}

				values := make(map[string]*float64, len(regs))
				readOK := true
				for _, name := range names {
					reg := regs[name]
					raw, err := client.ReadHoldingRegisters(reg.Address, 1)
					if err != nil {
						log.Debug("skip", "reg", name, "err", err)
						readOK = false
						continue
					}
					v := float64(raw[0]) * reg.Multiplier
					if raw[0] == 0 || raw[0] == 65535 {
						continue
					}
					v = math.Round(v*100) / 100
					values[name] = &v
				}

				if !readOK {
					client.Disconnect()
					log.Debug("read errors detected, reconnecting immediately")
					select {
					case <-ctx.Done():
						log.Info("shutdown")
						return nil
					case <-time.After(time.Duration(cfg.Logger.ReconnectDelay) * time.Second):
						continue
					}
				}

				tsISO := time.Now().Format("2006-01-02T15:04:05")
				tsEpoch := float64(time.Now().UnixNano()) / 1e9

				if !quiet {
					printSnapshot(values, names, regs, tsISO, interval)
				}
				if csvW != nil {
					csvW.Write(tsISO, values)
				}
				if influxW != nil {
					influxW.Send(tsEpoch, values)
				}
				if mqttW != nil {
					mqttW.Send(tsISO, values)
				}

				select {
				case <-ctx.Done():
					log.Info("shutdown")
					return nil
				case <-ticker.C:
				}
			}
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("deye-logger %s\ncommit: %s\nbuilt:  %s\ngo:     %s\n", version, commit, date, runtime.Version())
		},
	}

	rootCmd.AddCommand(versionCmd)
	rootCmd.Flags().StringVar(&cfgPath, "config", "", "YAML config path")
	rootCmd.Flags().IntVarP(&interval, "interval", "i", 30, "Poll interval (s)")
	rootCmd.Flags().StringVar(&csvPath, "csv", "data/deye.csv", "CSV output path")
	rootCmd.Flags().BoolVar(&noCSV, "no-csv", false, "Disable CSV")
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "No console output")
	rootCmd.Flags().BoolVar(&enableInflux, "influx", false, "Send to InfluxDB")
	rootCmd.Flags().BoolVar(&enableMQTT, "mqtt", false, "Send to MQTT")
	rootCmd.Flags().BoolVar(&includeUnverified, "include-unverified", false, "Read invalid registers")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "Debug logging")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func printSnapshot(values map[string]*float64, names []string, regs map[string]registers.Register, ts string, interval int) {
	fmt.Printf("\n%s\n", repeat("-", 55))
	fmt.Printf("  Deye 6kW  @  %s  (шаг %dс)\n", ts, interval)
	fmt.Printf("%s\n", repeat("-", 55))
	for _, name := range names {
		reg := regs[name]
		unit := unitsRU[reg.Unit]
		if unit == "" {
			unit = reg.Unit
		}
		if v, ok := values[name]; ok && v != nil {
			fmt.Printf("  %-20s %8.2f %s\n", reg.Description, *v, unit)
		} else {
			fmt.Printf("  %-20s %8s %s\n", reg.Description, "-", unit)
		}
	}
	fmt.Printf("%s\n", repeat("-", 55))
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
