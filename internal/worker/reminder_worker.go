package worker

import (
	"time"

	"github.com/oziev02/event-calendar-service/internal/domain"
	"github.com/oziev02/event-calendar-service/pkg/logger"
)

// ReminderWorker обрабатывает задачи напоминаний
type ReminderWorker struct {
	taskChan     chan *domain.ReminderTask
	sender       domain.ReminderSender
	logger       logger.Logger
	checkInterval time.Duration
	done         chan struct{}
}

// NewReminderWorker создает новый воркер напоминаний
func NewReminderWorker(
	taskChan chan *domain.ReminderTask,
	sender domain.ReminderSender,
	log logger.Logger,
	checkInterval time.Duration,
) *ReminderWorker {
	return &ReminderWorker{
		taskChan:      taskChan,
		sender:        sender,
		logger:        log,
		checkInterval: checkInterval,
		done:          make(chan struct{}),
	}
}

// Start запускает воркер напоминаний
func (w *ReminderWorker) Start() {
	go w.process()
}

// Stop останавливает воркер напоминаний
func (w *ReminderWorker) Stop() {
	close(w.done)
}

// process обрабатывает задачи напоминаний
func (w *ReminderWorker) process() {
	ticker := time.NewTicker(w.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case task := <-w.taskChan:
			now := time.Now()
			if task.Time.Before(now) || task.Time.Equal(now) {
				if err := w.sender.SendReminder(task); err != nil {
					w.logger.Log(logger.LevelError, "Failed to send reminder", map[string]interface{}{
						"error":   err.Error(),
						"event_id": task.EventID,
					})
				}
			} else {
				// Переназначить на позже
				go w.scheduleReminder(task)
			}
		case <-ticker.C:
			// Периодическая проверка просроченных напоминаний
		case <-w.done:
			return
		}
	}
}

// scheduleReminder планирует напоминание для проверки позже
func (w *ReminderWorker) scheduleReminder(task *domain.ReminderTask) {
	now := time.Now()
	delay := task.Time.Sub(now)
	if delay > 0 {
		time.Sleep(delay)
		w.taskChan <- task
	}
}

