package main

import "testing"

func TestProductionURLValidation(t *testing.T) {
	for _, db := range []string{
		"postgres://db.example/sinet?sslmode=verify-full&host=%2Fvar%2Frun%2Fpostgresql",
		"postgres://user:sslmode=verify-full@db.example/sinet?sslmode=disable",
		"postgres://db.example/sinet?sslmode=disable&application_name=sslmode=verify-full",
		"postgres://db.example/sinet?sslmode=verify-full&sslmode=disable",
		"postgres://db.example/sinet?sslmode=verify-full&sslmode=verify-full",
		"postgres://db.example/sinet?sslmode=verify-full;sslmode=disable",
		"postgres://db.example/sinet?sslmode=verify-full#fragment",
		"https://db.example/sinet?sslmode=verify-full",
	} {
		if productionURLs("https://sinet.example", db) == nil {
			t.Fatalf("accepted insecure/ambiguous database config %q", db)
		}
	}
	db := "postgres://user:password@db.example/sinet?sslmode=verify-full"
	for _, origin := range []string{"https://user:secret@sinet.example", "https://sinet.example/path", "https://sinet.example/?query", "https://sinet.example/#fragment", "http://sinet.example"} {
		if productionURLs(origin, db) == nil {
			t.Fatalf("accepted invalid origin %q", origin)
		}
	}
	if err := productionURLs("https://sinet.example", db); err != nil {
		t.Fatal(err)
	}
}
