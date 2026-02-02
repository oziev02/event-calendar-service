package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oziev02/event-calendar-service/configs"
	"github.com/oziev02/event-calendar-service/internal/server"
)

func main() {
	// Загрузить конфигурацию
	cfg := configs.Load()

	// Создать и запустить сервер
	srv, err := server.NewServer(cfg)
	if err != nil {
		panic(err)
	}

	// Запустить сервер в горутине
	go func() {
		if err := srv.Start(); err != nil {
			panic(err)
		}
	}()

	// Ожидать сигнал прерывания для корректного завершения сервера
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Корректное завершение с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		panic(err)
	}
}

