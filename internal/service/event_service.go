package service

import (
	"time"

	"github.com/oziev02/event-calendar-service/internal/domain"
)

// EventService обрабатывает бизнес-логику для событий
type EventService struct {
	repo domain.EventRepository
}

// NewEventService создает новый сервис событий
func NewEventService(repo domain.EventRepository) *EventService {
	return &EventService{repo: repo}
}

// CreateEvent создает новое событие
func (s *EventService) CreateEvent(userID, text string, date time.Time, reminderTime *time.Time) (*domain.Event, error) {
	event := &domain.Event{
		ID:           generateID(),
		UserID:       userID,
		Text:         text,
		Date:         date,
		ReminderTime: reminderTime,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Archived:     false,
	}

	if err := event.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(event); err != nil {
		return nil, err
	}

	return event, nil
}

// UpdateEvent обновляет существующее событие
func (s *EventService) UpdateEvent(userID, eventID, text string, date time.Time, reminderTime *time.Time) (*domain.Event, error) {
	event, err := s.repo.GetByID(userID, eventID)
	if err != nil {
		return nil, err
	}

	event.Text = text
	event.Date = date
	event.ReminderTime = reminderTime
	event.UpdatedAt = time.Now()

	if err := event.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(event); err != nil {
		return nil, err
	}

	return event, nil
}

// DeleteEvent удаляет событие
func (s *EventService) DeleteEvent(userID, eventID string) error {
	_, err := s.repo.GetByID(userID, eventID)
	if err != nil {
		return err
	}

	return s.repo.Delete(userID, eventID)
}

// GetEventsForDay возвращает события за конкретный день
func (s *EventService) GetEventsForDay(userID string, date time.Time) ([]*domain.Event, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)
	
	return s.repo.GetByDateRange(userID, start, end)
}

// GetEventsForWeek возвращает события за неделю, начиная с указанной даты
func (s *EventService) GetEventsForWeek(userID string, date time.Time) ([]*domain.Event, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(7 * 24 * time.Hour)
	
	return s.repo.GetByDateRange(userID, start, end)
}

// GetEventsForMonth возвращает события за месяц
func (s *EventService) GetEventsForMonth(userID string, date time.Time) ([]*domain.Event, error) {
	start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	nextMonth := start.AddDate(0, 1, 0)
	end := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, nextMonth.Location())
	
	return s.repo.GetByDateRange(userID, start, end)
}

// generateID генерирует простой ID (в продакшене использовать UUID)
func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

