package storage

import (
	"errors"
	"sync"
	"time"

	"github.com/oziev02/event-calendar-service/internal/domain"
)

// MemoryRepository реализует хранение событий в памяти
type MemoryRepository struct {
	mu     sync.RWMutex
	events map[string]*domain.Event // ключ: userID:eventID
}

// NewMemoryRepository создает новый репозиторий в памяти
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		events: make(map[string]*domain.Event),
	}
}

// key генерирует ключ для хранения
func (r *MemoryRepository) key(userID, eventID string) string {
	return userID + ":" + eventID
}

// Create создает новое событие
func (r *MemoryRepository) Create(event *domain.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.key(event.UserID, event.ID)
	if _, exists := r.events[key]; exists {
		return errors.New("event already exists")
	}

	r.events[key] = event
	return nil
}

// Update обновляет существующее событие
func (r *MemoryRepository) Update(event *domain.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.key(event.UserID, event.ID)
	if _, exists := r.events[key]; !exists {
		return domain.ErrEventNotFound
	}

	r.events[key] = event
	return nil
}

// Delete удаляет событие
func (r *MemoryRepository) Delete(userID, eventID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.key(userID, eventID)
	if _, exists := r.events[key]; !exists {
		return domain.ErrEventNotFound
	}

	delete(r.events, key)
	return nil
}

// GetByID получает событие по ID
func (r *MemoryRepository) GetByID(userID, eventID string) (*domain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := r.key(userID, eventID)
	event, exists := r.events[key]
	if !exists {
		return nil, domain.ErrEventNotFound
	}

	return event, nil
}

// GetByDateRange получает события в диапазоне дат
func (r *MemoryRepository) GetByDateRange(userID string, start, end time.Time) ([]*domain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.Event
	for _, event := range r.events {
		if event.UserID == userID && !event.Archived {
			if (event.Date.Equal(start) || event.Date.After(start)) && event.Date.Before(end) {
				result = append(result, event)
			}
		}
	}

	return result, nil
}

// GetAllActive получает все активные события пользователя
func (r *MemoryRepository) GetAllActive(userID string) ([]*domain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.Event
	for _, event := range r.events {
		if event.UserID == userID && !event.Archived {
			result = append(result, event)
		}
	}

	return result, nil
}

// ArchiveOldEvents архивирует события старше указанного времени
func (r *MemoryRepository) ArchiveOldEvents(before time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, event := range r.events {
		if event.Date.Before(before) && !event.Archived {
			event.Archived = true
		}
	}

	return nil
}

