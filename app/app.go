package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	server *http.Server
}

func NewApp(addr string, handler http.Handler) *App {
	return &App{
		server: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}
}

func (a *App) Start() {
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.server.Shutdown(ctx); err != nil {
			log.Println("Server shutdown error: ", err)
			return
		}
		log.Println("Server stopped gracefully")
	}()
	log.Println("Blog Server is running🎉🎉. Press Ctrl+C to stop")
	if err := a.server.ListenAndServe(); err != nil {
		log.Println("Server error: ", err)
		return
	}
}
