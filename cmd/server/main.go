package main

import (
	"context"
	"embed"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/jisun/tistory-indexer/internal/app"
)

//go:embed templates/index.html
var templateFS embed.FS

func main() {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	a, err := app.Build(ctx)
	if err != nil {
		slog.Error("failed to initialize", "err", err)
		os.Exit(1)
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	tmpl := template.Must(template.New("index.html").ParseFS(templateFS, "templates/index.html"))
	router.SetHTMLTemplate(tmpl)

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	router.GET("/api/state", func(c *gin.Context) {
		st, err := a.Store.Load()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		data, _ := json.Marshal(st)
		c.Data(http.StatusOK, "application/json", data)
	})

	router.POST("/api/run", func(c *gin.Context) {
		if err := a.Service.Run(c.Request.Context()); err != nil {
			slog.Error("indexer run failed", "err", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "실패: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "sitemap 업데이트 완료"})
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8090"
	}

	slog.Info("server started", "url", "http://localhost:"+port)
	if err := router.Run(":" + port); err != nil {
		slog.Error("server error", "err", err)
	}
}
