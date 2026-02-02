package reminder

import (
	"fmt"

	"github.com/oziev02/event-calendar-service/internal/domain"
	"github.com/oziev02/event-calendar-service/pkg/logger"
)

// ConsoleReminderSender отправляет напоминания в консоль
type ConsoleReminderSender struct {
	logger logger.Logger
}

// NewConsoleReminderSender создает новый отправитель напоминаний в консоль
func NewConsoleReminderSender(log logger.Logger) *ConsoleReminderSender {
	return &ConsoleReminderSender{logger: log}
}

// SendReminder отправляет напоминание
func (s *ConsoleReminderSender) SendReminder(task *domain.ReminderTask) error {
	message := fmt.Sprintf("REMINDER: Event '%s' for user %s", task.Text, task.UserID)
	s.logger.Log(logger.LevelInfo, message, map[string]interface{}{
		"event_id": task.EventID,
		"user_id":  task.UserID,
		"time":     task.Time,
	})
	return nil
}

