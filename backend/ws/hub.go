package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"nvtop-server/gpu"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// client represents a single websocket connection
type client struct {
	conn *websocket.Conn
	send chan []byte
}

// Hub manages all websocket connections and broadcasts GPU data
type Hub struct {
	mu      sync.RWMutex
	clients map[*client]bool

	driverVer string
	cudaVer   string

	stopCh chan struct{}
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*client]bool),
		stopCh:  make(chan struct{}),
	}
}

// Start begins the data collection and broadcast loop
func (h *Hub) Start() {
	// Get driver/cuda version once at start
	h.driverVer, _ = gpu.GetDriverVersion()
	h.cudaVer, _ = gpu.GetCUDAVersion()

	go h.broadcastLoop()
}

// Stop shuts down the hub
func (h *Hub) Stop() {
	close(h.stopCh)
}

func (h *Hub) broadcastLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-h.stopCh:
			return
		case <-ticker.C:
			snapshot := h.collect()
			data, err := json.Marshal(snapshot)
			if err != nil {
				log.Printf("Error marshaling GPU snapshot: %v", err)
				continue
			}

			h.mu.RLock()
			for c := range h.clients {
				select {
				case c.send <- data:
				default:
					// Client buffer full, skip
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) collect() gpu.Snapshot {
	snapshot := gpu.Snapshot{
		Timestamp:     time.Now().Unix(),
		DriverVersion: h.driverVer,
		CUDAVersion:   h.cudaVer,
		System:        gpu.GetSystemInfo(),
	}

	gpus, err := gpu.GetAllGPUInfo()
	if err != nil {
		log.Printf("Error collecting GPU info: %v", err)
	} else {
		snapshot.GPUs = gpus
	}

	return snapshot
}

// ServeWS handles websocket upgrade requests
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	c := &client{
		conn: conn,
		send: make(chan []byte, 64),
	}

	h.mu.Lock()
	h.clients[c] = true
	h.mu.Unlock()

	go h.writePump(c)
	go h.readPump(c)
}

func (h *Hub) writePump(c *client) {
	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *Hub) readPump(c *client) {
	defer func() {
		h.mu.Lock()
		delete(h.clients, c)
		h.mu.Unlock()
		close(c.send)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
	}
}
