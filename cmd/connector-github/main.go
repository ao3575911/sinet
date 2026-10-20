package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sinet/internal/githubconnector"
)

func main() {
	address := os.Getenv("SINET_GITHUB_LISTEN")
	if address == "" {
		address = "127.0.0.1:8090"
	}
	server := &http.Server{
		Addr: address, Handler: githubconnector.NewHandler(),
		ReadHeaderTimeout: 3 * time.Second, ReadTimeout: 5 * time.Second,
		WriteTimeout: 12 * time.Second, IdleTimeout: 30 * time.Second,
		MaxHeaderBytes: 16384,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	log.Printf("GitHub public gateway listening on %s; registry approval is separate", address)
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- server.ListenAndServe() }()
	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			_ = server.Close()
		}
	}
}
