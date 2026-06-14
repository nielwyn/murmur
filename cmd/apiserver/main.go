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

	"murmur/internal/api"
	"murmur/internal/config"
	"murmur/internal/database"
	"murmur/internal/feedfetch"

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
		Addr:    ":" + port,
		Handler: handler,
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
