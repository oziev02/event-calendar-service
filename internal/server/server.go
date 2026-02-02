package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/oziev02/event-calendar-service/configs"
	"github.com/oziev02/event-calendar-service/internal/domain"
	httphandler "github.com/oziev02/event-calendar-service/internal/http"
	"github.com/oziev02/event-calendar-service/internal/http/handlers"
	"github.com/oziev02/event-calendar-service/internal/reminder"
	"github.com/oziev02/event-calendar-service/internal/service"
	"github.com/oziev02/event-calendar-service/internal/storage"
	"github.com/oziev02/event-calendar-service/internal/worker"
	"github.com/oziev02/event-calendar-service/pkg/logger"
)

// Server представляет HTTP сервер
type Server struct {
	httpServer     *http.Server
	logger         logger.Logger
	reminderWorker *worker.ReminderWorker
	cleanupWorker  *worker.CleanupWorker
	reminderChan   chan *domain.ReminderTask
}

// NewServer создает новый HTTP сервер
func NewServer(cfg *configs.Config) (*Server, error) {
	// Инициализировать логгер
	asyncLogger := logger.NewAsyncLogger(cfg.LoggerBufferSize)

	// Инициализировать репозиторий
	repo := storage.NewMemoryRepository()

	// Инициализировать канал напоминаний
	reminderChan := make(chan *domain.ReminderTask, 100)

	// Инициализировать отправитель напоминаний
	reminderSender := reminder.NewConsoleReminderSender(asyncLogger)

	// Инициализировать воркеры
	reminderWorker := worker.NewReminderWorker(
		reminderChan,
		reminderSender,
		asyncLogger,
		cfg.ReminderCheckInterval,
	)
	reminderWorker.Start()

	cleanupWorker := worker.NewCleanupWorker(
		repo,
		asyncLogger,
		cfg.CleanupInterval,
		cfg.ArchiveAfter,
	)
	cleanupWorker.Start()

	// Инициализировать сервис приложения
	eventService := service.NewEventService(repo)

	// Инициализировать обработчики
	eventHandler := handlers.NewEventHandler(eventService, asyncLogger, reminderChan)

	// Настроить маршруты
	handler := httphandler.Router(eventHandler, asyncLogger)

	// Создать HTTP сервер
	httpServer := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer:     httpServer,
		logger:         asyncLogger,
		reminderWorker: reminderWorker,
		cleanupWorker:  cleanupWorker,
		reminderChan:   reminderChan,
	}, nil
}

// Start запускает HTTP сервер
func (s *Server) Start() error {
	addr := s.httpServer.Addr
	s.logger.Log(logger.LevelInfo, fmt.Sprintf("Starting server on %s", addr), nil)
	return s.httpServer.ListenAndServe()
}

// Shutdown корректно останавливает сервер
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Log(logger.LevelInfo, "Shutting down server", nil)

	s.reminderWorker.Stop()
	s.cleanupWorker.Stop()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	return s.logger.Close()
}

