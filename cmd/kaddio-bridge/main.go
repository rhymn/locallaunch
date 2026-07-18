package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"kaddio-bridge/internal/api"
	"kaddio-bridge/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v\n", err)
	}

	cfgPath, _ := config.Dir()
	log.Printf("Kaddio Bridge started\n\nConfig:\n%s/config.json\n\nToken:\n%s\n\n", cfgPath, cfg.Token)

	srv := api.New(cfg)

	ln, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		if isAddrInUse(err) {
			log.Fatalf("Port %s is already in use. Another instance may be running.\n", cfg.Address)
		}
		log.Fatalf("Listening on %s: %v\n", cfg.Address, err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Listening:\n%s\n", cfg.Address)
		if err := srv.Serve(ln); err != nil && !errors.Is(err, context.Canceled) {
			log.Fatalf("Server error: %v\n", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}

func isAddrInUse(err error) bool {
	var opErr *net.OpError
	if !errors.As(err, &opErr) {
		return false
	}
	return strings.Contains(opErr.Err.Error(), "address already in use")
}
