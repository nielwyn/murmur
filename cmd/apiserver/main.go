package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nielwyn/murmur/internal/api"
	"github.com/nielwyn/murmur/internal/config"
	"github.com/nielwyn/murmur/internal/database"
	"github.com/nielwyn/murmur/internal/feedfetch"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultPort = "8080"

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), cfg.DBUrl)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	defer pool.Close()

	db := database.New(pool)
	handler := api.NewServer(db, &cfg)

	port := os.Getenv("MURMUR_PORT")
	if port == "" {
		port = defaultPort
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	wg.Go(func() {
		feedfetch.NewScheduler(db, feedfetch.Config{}).Run(ctx)
	})

	go func() {
		log.Printf("murmur api server listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("error during shutdown: %v", err)
	}

	wg.Wait()
}
