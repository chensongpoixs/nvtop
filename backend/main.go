package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"nvtop-server/api"
	"nvtop-server/gpu"
	"nvtop-server/ws"
)

//go:embed all:frontend/dist
var frontendFiles embed.FS

func main() {
	// Initialize NVML
	if err := gpu.Init(); err != nil {
		log.Printf("WARNING: NVML init failed: %v (running without GPU monitoring)", err)
	} else {
		defer gpu.Shutdown()
	}

	// Start WebSocket hub
	hub := ws.NewHub()
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
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("nvtop-web server starting on http://localhost%s", addr)

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
