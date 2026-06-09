package modbus

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/drycool/deye-logger-go/internal/logger"
)

// Client wraps a Modbus TCP connection to the Deye WiFi logger.
// Uses raw TCP with Solarman V5 framing (simplified: direct Modbus TCP).
type Client struct {
	mu       sync.Mutex
	conn     net.Conn
	host     string
	port     int
	slaveID  byte
	timeout  time.Duration
	log      *slog.Logger
}

// New creates a new Modbus TCP client.
func New(host string, port int, slaveID int) *Client {
	return &Client{
		host:    host,
		port:    port,
		slaveID: byte(slaveID),
		timeout: 10 * time.Second,
		log:     logger.Get("modbus"),
	}
}

// Connect establishes the TCP connection.
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	conn, err := net.DialTimeout("tcp", addr, c.timeout)
	if err != nil {
		return fmt.Errorf("dial %s: %w", addr, err)
	}
	c.conn = conn
	c.log.Info("connected", "addr", addr)
	return nil
}

// Disconnect closes the connection.
func (c *Client) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
		c.log.Info("disconnected")
	}
}

// ReadHoldingRegisters reads one or more holding registers (FC3).
func (c *Client) ReadHoldingRegisters(addr uint16, quantity uint16) ([]uint16, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Build Modbus TCP frame
	// Transaction ID (2) + Protocol ID (2) + Length (2) + Unit ID (1) + FC (1) + Start (2) + Qty (2)
	frame := make([]byte, 12)
	binary.BigEndian.PutUint16(frame[0:2], 1)           // transaction ID
	binary.BigEndian.PutUint16(frame[2:4], 0)           // protocol ID
	binary.BigEndian.PutUint16(frame[4:6], 6)           // length
	frame[6] = c.slaveID                                // unit ID
	frame[7] = 0x03                                     // FC3
	binary.BigEndian.PutUint16(frame[8:10], addr)       // start register
	binary.BigEndian.PutUint16(frame[10:12], quantity)  // quantity

	c.conn.SetDeadline(time.Now().Add(c.timeout))
	if _, err := c.conn.Write(frame); err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	// Read response header (9 bytes min)
	header := make([]byte, 9)
	if _, err := c.conn.Read(header); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	fc := header[7]
	if fc&0x80 != 0 {
		excCode := header[8]
		return nil, fmt.Errorf("modbus exception FC=0x%02X code=%d", fc, excCode)
	}

	byteCount := int(header[8])
	data := make([]byte, byteCount)
	if _, err := c.conn.Read(data); err != nil {
		return nil, fmt.Errorf("read data: %w", err)
	}

	regs := make([]uint16, quantity)
	for i := range regs {
		regs[i] = binary.BigEndian.Uint16(data[i*2 : i*2+2])
	}
	return regs, nil
}
