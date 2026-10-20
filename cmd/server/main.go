package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sinet/internal/api"
	"sinet/internal/mail"
	"sinet/internal/model"
	"sinet/internal/store"
	"sinet/migrations"
	"strings"
	"syscall"
	"time"
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	dbURL := env("DATABASE_URL", "postgres://projectname:local-development-only@127.0.0.1:5432/projectname?sslmode=disable")
	development := env("APP_ENV", "development") == "development"
	publicURL := env("APP_PUBLIC_URL", "http://127.0.0.1:8080")
	if !development {
		u, err := url.Parse(publicURL)
		if err != nil || u.Scheme != "https" || u.Host == "" {
			return fmt.Errorf("production APP_PUBLIC_URL must be HTTPS")
		}
		if !strings.Contains(dbURL, "sslmode=verify-full") {
			return fmt.Errorf("production database requires sslmode=verify-full")
		}
		if os.Getenv("OTP_HASH_KEY") == "" || len(os.Getenv("OTP_HASH_KEY")) < 32 {
			return fmt.Errorf("production OTP_HASH_KEY requires at least 32 random characters")
		}
		if os.Getenv("SMTP_ADDR") == "" || os.Getenv("SMTP_FROM") == "" {
			return fmt.Errorf("production SMTP configuration required")
		}
	}
	db, err := store.Open(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("database connection unavailable")
	}
	defer db.Close()
	if err = migrations.Apply(ctx, db); err != nil {
		return err
	}
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		fmt.Println("migrations applied")
		return nil
	}
	provider, err := model.New(os.Getenv("MODEL_API_BASE_URL"), os.Getenv("MODEL_API_KEY"), os.Getenv("MODEL_NAME"), development)
	if err != nil {
		return err
	}
	sender := mail.SMTP{Addr: env("SMTP_ADDR", "127.0.0.1:1025"), From: env("SMTP_FROM", "no-reply@project-name.example"), Username: os.Getenv("SMTP_USERNAME"), Password: os.Getenv("SMTP_PASSWORD"), RequireTLS: !development}
	handler := api.New(db, sender, provider, publicURL)
	if key := os.Getenv("OTP_HASH_KEY"); key != "" {
		handler.OTPKey = []byte(key)
	}
	server := &http.Server{Addr: env("APP_HTTP_ADDR", "127.0.0.1:8080"), Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16384}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	log.Printf("SINet API listening on %s; model configured=%t", server.Addr, provider != nil)
	err = server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}
func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
