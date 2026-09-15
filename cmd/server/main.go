package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mekari/pos-phoenix/internal/app"
	"github.com/mekari/pos-phoenix/internal/auth"
	"github.com/mekari/pos-phoenix/internal/database"
	"github.com/mekari/pos-phoenix/internal/httpserver"
)

func main() {
	cfg := app.ConfigFromEnv()

	loc, err := time.LoadLocation(cfg.BusinessTimezone)
	if err != nil {
		log.Fatalf("invalid timezone %q: %v", cfg.BusinessTimezone, err)
	}

	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	// Bootstrap admin if credentials provided
	if cfg.AdminEmail != "" && cfg.AdminPassword != "" {
		authSvc := auth.Service{DB: db}
		if err := authSvc.BootstrapAdmin(context.Background(), cfg.AdminEmail, cfg.AdminPassword); err != nil {
			log.Printf("bootstrap admin: %v", err)
		}
	}

	handler, err := httpserver.New(db, cfg.SecureCookies, loc)
	if err != nil {
		log.Fatalf("http server: %v", err)
	}

	fmt.Printf("POS Phoenix listening on %s\n", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, handler); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
