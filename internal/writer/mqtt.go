package writer

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/drycool/deye-logger-go/internal/logger"
)

// MqttWriter publishes metrics snapshots to MQTT.
type MqttWriter struct {
	client mqtt.Client
	topic  string
	log    *slog.Logger
}

// NewMqtt creates and connects an MQTT writer.
func NewMqtt(host string, port int, topic string) (*MqttWriter, error) {
	broker := fmt.Sprintf("tcp://%s:%d", host, port)
	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID("deye-logger-go").
		SetAutoReconnect(true).
		SetConnectTimeout(10 * time.Second)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("mqtt connect: %w", token.Error())
	}

	return &MqttWriter{
		client: client,
		topic:  topic,
		log:    logger.Get("mqtt"),
	}, nil
}

// Send publishes one snapshot.
func (w *MqttWriter) Send(tsISO string, values map[string]*float64) {
	payload := map[string]any{"ts": tsISO}
	for name, val := range values {
		if val != nil {
			payload[name] = *val
		}
	}

	data, err := json.Marshal(payload)
	if err != nil {
		w.log.Error("marshal payload", "err", err)
		return
	}

	if token := w.client.Publish(w.topic, 0, false, data); token.Wait() && token.Error() != nil {
		w.log.Debug("publish failed", "err", token.Error())
	}
}

// Close disconnects the MQTT client.
func (w *MqttWriter) Close() {
	if w.client.IsConnected() {
		w.client.Disconnect(250)
	}
}
