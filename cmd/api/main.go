package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/itsahyarr/youtube-watcher/docs"
	"github.com/itsahyarr/youtube-watcher/internal"
)

// @title           YouTube Watcher API
// @version         1.0
// @description     YouTube scraping service that opens a YouTube URL in a browser, clicks play, and logs to MongoDB.

// @contact.name   Muhammad Ahyaruddin
// @contact.email  ahyaruddin07@gmail.com

// @host           localhost:8080
// @BasePath       /

// @schemes        http https

func main() {
	cfg, err := internal.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo, err := internal.NewRepository(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := repo.Close(context.Background()); err != nil {
			log.Printf("error closing MongoDB: %v", err)
		}
	}()

	browser := internal.NewBrowserClient(cfg)
	svc := internal.NewService(repo, browser)
	handler := internal.NewHandler(cfg, svc)

	router := gin.Default()

	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api/v1/scrape")
	{
		api.POST("/youtube/play", handler.ScrapeYouTube)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: router,
	}

	go func() {
		fmt.Printf("YouTube Watcher listening on :%s\n", cfg.AppPort)
		fmt.Printf("Swagger UI available at http://localhost:%s/docs/index.html\n", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nshutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	fmt.Println("server stopped")
}
