package domain

import (
	"errors"
	"time"
)

var (
	ErrEventNotFound       = errors.New("event not found")
	ErrInvalidDate         = errors.New("invalid date format")
	ErrInvalidUserID       = errors.New("invalid user id")
	ErrInvalidEventText    = errors.New("invalid event text")
	ErrInvalidReminderTime = errors.New("invalid reminder time")
)

// Event представляет событие календаря
type Event struct {
	ID           string
	UserID       string
	Date         time.Time
	Text         string
	ReminderTime *time.Time // Опциональное время напоминания
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Archived     bool
}

// Validate валидирует данные события
func (e *Event) Validate() error {
	if e.UserID == "" {
		return ErrInvalidUserID
	}
	if e.Text == "" {
		return ErrInvalidEventText
	}
	if e.Date.IsZero() {
		return ErrInvalidDate
	}
	return nil
}

// IsReminderDue проверяет, наступило ли время напоминания
func (e *Event) IsReminderDue(now time.Time) bool {
	if e.ReminderTime == nil {
		return false
	}
	return !e.ReminderTime.After(now) && !e.Archived
}

// EventRepository определяет интерфейс для хранения событий
type EventRepository interface {
	Create(event *Event) error
	Update(event *Event) error
	Delete(userID, eventID string) error
	GetByID(userID, eventID string) (*Event, error)
	GetByDateRange(userID string, start, end time.Time) ([]*Event, error)
	GetAllActive(userID string) ([]*Event, error)
	ArchiveOldEvents(before time.Time) error
}
