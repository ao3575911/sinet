package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sinet/internal/githubconnector"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "healthcheck" {
		if err := healthcheck(); err != nil {
			log.Print("gateway healthcheck failed")
			os.Exit(1)
		}
		fmt.Println("ok")
		return
	}
	if len(os.Args) != 1 {
		log.Fatal("usage: connector-github [healthcheck]")
	}
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

// healthcheck is a local process probe, never a connector egress exception.
func healthcheck() error {
	address := os.Getenv("SINET_GITHUB_LISTEN")
	if address == "" {
		address = "127.0.0.1:8090"
	}
	_, port, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	client := &http.Client{Transport: &http.Transport{Proxy: nil}, Timeout: 2 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("redirect denied") }}
	defer client.CloseIdleConnections()
	resp, err := client.Get("http://" + net.JoinHostPort("127.0.0.1", port) + "/healthz")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 65))
	if err != nil || resp.StatusCode != 200 || string(body) != "{\"status\":\"ok\"}\n" {
		return errors.New("unhealthy gateway")
	}
	return nil
}
