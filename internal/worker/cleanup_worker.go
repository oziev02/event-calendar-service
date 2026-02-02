package worker

import (
	"time"

	"github.com/oziev02/event-calendar-service/internal/domain"
	"github.com/oziev02/event-calendar-service/pkg/logger"
)

// CleanupWorker архивирует старые события
type CleanupWorker struct {
	repo         domain.EventRepository
	logger       logger.Logger
	interval     time.Duration
	archiveAfter time.Duration
	done         chan struct{}
}

// NewCleanupWorker создает новый воркер очистки
func NewCleanupWorker(
	repo domain.EventRepository,
	log logger.Logger,
	interval time.Duration,
	archiveAfter time.Duration,
) *CleanupWorker {
	return &CleanupWorker{
		repo:         repo,
		logger:       log,
		interval:     interval,
		archiveAfter: archiveAfter,
		done:         make(chan struct{}),
	}
}

// Start запускает воркер очистки
func (w *CleanupWorker) Start() {
	go w.process()
}

// Stop останавливает воркер очистки
func (w *CleanupWorker) Stop() {
	close(w.done)
}

// process периодически архивирует старые события
func (w *CleanupWorker) process() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Запустить сразу при старте
	w.archiveOldEvents()

	for {
		select {
		case <-ticker.C:
			w.archiveOldEvents()
		case <-w.done:
			return
		}
	}
}

// archiveOldEvents архивирует события старше archiveAfter
func (w *CleanupWorker) archiveOldEvents() {
	cutoff := time.Now().Add(-w.archiveAfter)
	if err := w.repo.ArchiveOldEvents(cutoff); err != nil {
		w.logger.Log(logger.LevelError, "Failed to archive old events", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		w.logger.Log(logger.LevelInfo, "Archived old events", map[string]interface{}{
			"cutoff": cutoff,
		})
	}
}

