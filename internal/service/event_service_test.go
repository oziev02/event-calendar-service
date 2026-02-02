package service

import (
	"errors"
	"testing"
	"time"

	"github.com/oziev02/event-calendar-service/internal/domain"
	"github.com/oziev02/event-calendar-service/internal/storage"
)

func TestEventService_CreateEvent(t *testing.T) {
	repo := storage.NewMemoryRepository()
	service := NewEventService(repo)

	userID := "user1"
	text := "Test event"
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	event, err := service.CreateEvent(userID, text, date, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if event.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, event.UserID)
	}

	if event.Text != text {
		t.Errorf("Expected Text %s, got %s", text, event.Text)
	}

	if !event.Date.Equal(date) {
		t.Errorf("Expected Date %v, got %v", date, event.Date)
	}

	if event.Archived {
		t.Error("Expected event not to be archived")
	}
}

func TestEventService_CreateEvent_InvalidData(t *testing.T) {
	repo := storage.NewMemoryRepository()
	service := NewEventService(repo)

	tests := []struct {
		name    string
		userID  string
		text    string
		date    time.Time
		wantErr error
	}{
		{
			name:    "empty user ID",
			userID:  "",
			text:    "Test",
			date:    time.Now(),
			wantErr: domain.ErrInvalidUserID,
		},
		{
			name:    "empty text",
			userID:  "user1",
			text:    "",
			date:    time.Now(),
			wantErr: domain.ErrInvalidEventText,
		},
		{
			name:    "zero date",
			userID:  "user1",
			text:    "Test",
			date:    time.Time{},
			wantErr: domain.ErrInvalidDate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateEvent(tt.userID, tt.text, tt.date, nil)
			if err == nil {
				t.Fatal("Expected error, got nil")
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestEventService_UpdateEvent(t *testing.T) {
	repo := storage.NewMemoryRepository()
	service := NewEventService(repo)

	userID := "user1"
	text := "Original event"
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	event, err := service.CreateEvent(userID, text, date, nil)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	newText := "Updated event"
	newDate := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)

	updated, err := service.UpdateEvent(userID, event.ID, newText, newDate, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if updated.Text != newText {
		t.Errorf("Expected Text %s, got %s", newText, updated.Text)
	}

	if !updated.Date.Equal(newDate) {
		t.Errorf("Expected Date %v, got %v", newDate, updated.Date)
	}
}

func TestEventService_UpdateEvent_NotFound(t *testing.T) {
	repo := storage.NewMemoryRepository()
	service := NewEventService(repo)

	_, err := service.UpdateEvent("user1", "nonexistent", "Text", time.Now(), nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
		if !errors.Is(err, domain.ErrEventNotFound) {
		t.Errorf("Expected error %v, got %v", domain.ErrEventNotFound, err)
	}
}

func TestEventService_DeleteEvent(t *testing.T) {
	repo := storage.NewMemoryRepository()
	service := NewEventService(repo)

	userID := "user1"
	text := "Test event"
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	event, err := service.CreateEvent(userID, text, date, nil)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	err = service.DeleteEvent(userID, event.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify event is deleted
	_, err = service.GetEventsForDay(userID, date)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestEventService_DeleteEvent_NotFound(t *testing.T) {
	repo := storage.NewMemoryRepository()
	service := NewEventService(repo)

	err := service.DeleteEvent("user1", "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
		if !errors.Is(err, domain.ErrEventNotFound) {
		t.Errorf("Expected error %v, got %v", domain.ErrEventNotFound, err)
	}
}

func TestEventService_GetEventsForDay(t *testing.T) {
	repo := storage.NewMemoryRepository()
	service := NewEventService(repo)

	userID := "user1"
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create events on the same day
	_, err := service.CreateEvent(userID, "Event 1", date, nil)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	_, err = service.CreateEvent(userID, "Event 2", date, nil)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	// Create event on different day
	_, err = service.CreateEvent(userID, "Event 3", date.Add(24*time.Hour), nil)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	events, err := service.GetEventsForDay(userID, date)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}
}

func TestEventService_GetEventsForWeek(t *testing.T) {
	repo := storage.NewMemoryRepository()
	service := NewEventService(repo)

	userID := "user1"
	startDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create events throughout the week
	for i := 0; i < 7; i++ {
		_, err := service.CreateEvent(userID, "Event", startDate.Add(time.Duration(i)*24*time.Hour), nil)
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}

	events, err := service.GetEventsForWeek(userID, startDate)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(events) != 7 {
		t.Errorf("Expected 7 events, got %d", len(events))
	}
}

func TestEventService_GetEventsForMonth(t *testing.T) {
	repo := storage.NewMemoryRepository()
	service := NewEventService(repo)

	userID := "user1"
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create events in January
	for i := 1; i <= 5; i++ {
		eventDate := time.Date(2024, 1, i, 0, 0, 0, 0, time.UTC)
		_, err := service.CreateEvent(userID, "Event", eventDate, nil)
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}

	events, err := service.GetEventsForMonth(userID, date)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(events) != 5 {
		t.Errorf("Expected 5 events, got %d", len(events))
	}
}

