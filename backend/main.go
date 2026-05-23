package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"nvtop-server/api"
	"nvtop-server/config"
	"nvtop-server/gpu"
	"nvtop-server/ws"
)

//go:embed all:frontend/dist
var frontendFiles embed.FS

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize NVML
	if err := gpu.Init(); err != nil {
		log.Printf("WARNING: NVML init failed: %v (running without GPU monitoring)", err)
	} else {
		defer gpu.Shutdown()
	}

	// Start WebSocket hub
	hub := ws.NewHub(cfg.Monitor.PollIntervalSeconds)
	hub.Start()
	defer hub.Stop()

	mux := http.NewServeMux()

	// API
	mux.HandleFunc("/api/gpus", api.HandleGPUSnapshot)

	// WebSocket
	mux.HandleFunc("/ws", hub.ServeWS)

	// Serve Vue frontend
	frontendFS, err := fs.Sub(frontendFiles, "frontend/dist")
	if err != nil {
		log.Printf("WARNING: frontend not embedded, serving from filesystem: %v", err)
		// Fallback for development: serve from filesystem
		mux.Handle("/", http.FileServer(http.Dir("../frontend/dist")))
	} else {
		fileServer := http.FileServer(http.FS(frontendFS))
		mux.Handle("/", fileServer)
	}

	port := os.Getenv("PORT")
	if port != "" {
		cfg.Server.Port = 0
		fmt.Sscanf(port, "%d", &cfg.Server.Port)
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("nvtop-web server starting on http://%s", addr)
	log.Printf("Config: poll_interval=%ds, history_size=%d, log_level=%s",
		cfg.Monitor.PollIntervalSeconds, cfg.Monitor.HistorySize, cfg.Log.Level)

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		hub.Stop()
		os.Exit(0)
	}()

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
