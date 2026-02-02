package domain

import "time"

// ReminderTask представляет задачу напоминания для обработки
type ReminderTask struct {
	EventID string
	UserID  string
	Text    string
	Time    time.Time
}

// ReminderSender определяет интерфейс для отправки напоминаний
type ReminderSender interface {
	SendReminder(task *ReminderTask) error
}
