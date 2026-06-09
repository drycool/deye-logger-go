package writer

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/drycool/deye-logger-go/internal/logger"
)

// InfluxWriter sends metrics to InfluxDB v2 using line protocol.
type InfluxWriter struct {
	url    string
	token  string
	org    string
	bucket string
	client *http.Client
	log    *slog.Logger
}

// NewInflux creates an InfluxDB writer.
func NewInflux(url, token, org, bucket string) *InfluxWriter {
	return &InfluxWriter{
		url:    strings.TrimRight(url, "/"),
		token:  token,
		org:    org,
		bucket: bucket,
		client: &http.Client{Timeout: 10 * time.Second},
		log:    logger.Get("influx"),
	}
}

// Send writes one snapshot to InfluxDB.
func (w *InfluxWriter) Send(tsEpoch float64, values map[string]*float64) {
	tsNs := int64(tsEpoch * 1_000_000_000)
	var lines []string
	for name, val := range values {
		if val != nil {
			lines = append(lines, fmt.Sprintf("deye,source=inverter %s=%f %d", name, *val, tsNs))
		}
	}
	if len(lines) == 0 {
		return
	}

	endpoint := fmt.Sprintf("%s/api/v2/write?org=%s&bucket=%s&precision=ns", w.url, w.org, w.bucket)
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(strings.Join(lines, "\n")))
	if err != nil {
		w.log.Error("build request", "err", err)
		return
	}
	req.Header.Set("Authorization", "Token "+w.token)
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")

	resp, err := w.client.Do(req)
	if err != nil {
		w.log.Debug("send failed", "err", err)
		return
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		w.log.Warn("unexpected status", "status", resp.StatusCode)
	}
}

// Close releases HTTP resources.
func (w *InfluxWriter) Close() {
	w.client.CloseIdleConnections()
}
