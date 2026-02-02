package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/oziev02/event-calendar-service/internal/domain"
	"github.com/oziev02/event-calendar-service/internal/service"
	"github.com/oziev02/event-calendar-service/pkg/logger"
)

// EventHandler обрабатывает HTTP запросы для событий
type EventHandler struct {
	service    *service.EventService
	logger     logger.Logger
	reminderChan chan *domain.ReminderTask
}

// NewEventHandler создает новый обработчик событий
func NewEventHandler(
	service *service.EventService,
	log logger.Logger,
	reminderChan chan *domain.ReminderTask,
) *EventHandler {
	return &EventHandler{
		service:      service,
		logger:       log,
		reminderChan: reminderChan,
	}
}

// CreateEvent handles POST /create_event
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateEventRequest
	if err := h.decodeRequest(r, &req); err != nil {
		h.sendError(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		h.sendError(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	var reminderTime *time.Time
	if req.ReminderTime != "" {
		rt, err := time.Parse(time.RFC3339, req.ReminderTime)
		if err != nil {
			h.sendError(w, "Invalid reminder time format. Use RFC3339", http.StatusBadRequest)
			return
		}
		reminderTime = &rt
	}

	event, err := h.service.CreateEvent(req.UserID, req.Event, date, reminderTime)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidDate) || errors.Is(err, domain.ErrInvalidUserID) || errors.Is(err, domain.ErrInvalidEventText) {
			h.sendError(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.sendError(w, "Failed to create event", http.StatusInternalServerError)
		return
	}

	// Запланировать напоминание, если указано
	if reminderTime != nil {
		task := &domain.ReminderTask{
			EventID: event.ID,
			UserID:  event.UserID,
			Text:    event.Text,
			Time:    *reminderTime,
		}
		select {
		case h.reminderChan <- task:
		default:
			h.logger.Log(logger.LevelError, "Reminder channel full", nil)
		}
	}

	h.sendSuccess(w, map[string]interface{}{
		"event_id": event.ID,
		"message":  "Event created successfully",
	})
}

// UpdateEvent handles POST /update_event
func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UpdateEventRequest
	if err := h.decodeRequest(r, &req); err != nil {
		h.sendError(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		h.sendError(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	var reminderTime *time.Time
	if req.ReminderTime != "" {
		rt, err := time.Parse(time.RFC3339, req.ReminderTime)
		if err != nil {
			h.sendError(w, "Invalid reminder time format. Use RFC3339", http.StatusBadRequest)
			return
		}
		reminderTime = &rt
	}

	event, err := h.service.UpdateEvent(req.UserID, req.EventID, req.Event, date, reminderTime)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			h.sendError(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		if errors.Is(err, domain.ErrInvalidDate) || errors.Is(err, domain.ErrInvalidUserID) || errors.Is(err, domain.ErrInvalidEventText) {
			h.sendError(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.sendError(w, "Failed to update event", http.StatusInternalServerError)
		return
	}

	// Переназначить напоминание, если указано
	if reminderTime != nil {
		task := &domain.ReminderTask{
			EventID: event.ID,
			UserID:  event.UserID,
			Text:    event.Text,
			Time:    *reminderTime,
		}
		select {
		case h.reminderChan <- task:
		default:
			h.logger.Log(logger.LevelError, "Reminder channel full", nil)
		}
	}

	h.sendSuccess(w, map[string]interface{}{
		"message": "Event updated successfully",
	})
}

// DeleteEvent handles POST /delete_event
func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DeleteEventRequest
	if err := h.decodeRequest(r, &req); err != nil {
		h.sendError(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	err := h.service.DeleteEvent(req.UserID, req.EventID)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			h.sendError(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		h.sendError(w, "Failed to delete event", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, map[string]interface{}{
		"message": "Event deleted successfully",
	})
}

// GetEventsForDay handles GET /events_for_day
func (h *EventHandler) GetEventsForDay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	dateStr := r.URL.Query().Get("date")

	if userID == "" || dateStr == "" {
		h.sendError(w, "user_id and date are required", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.sendError(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	events, err := h.service.GetEventsForDay(userID, date)
	if err != nil {
		h.sendError(w, "Failed to get events", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, map[string]interface{}{
		"events": eventsToDTO(events),
	})
}

// GetEventsForWeek handles GET /events_for_week
func (h *EventHandler) GetEventsForWeek(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	dateStr := r.URL.Query().Get("date")

	if userID == "" || dateStr == "" {
		h.sendError(w, "user_id and date are required", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.sendError(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	events, err := h.service.GetEventsForWeek(userID, date)
	if err != nil {
		h.sendError(w, "Failed to get events", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, map[string]interface{}{
		"events": eventsToDTO(events),
	})
}

// GetEventsForMonth handles GET /events_for_month
func (h *EventHandler) GetEventsForMonth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	dateStr := r.URL.Query().Get("date")

	if userID == "" || dateStr == "" {
		h.sendError(w, "user_id and date are required", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.sendError(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	events, err := h.service.GetEventsForMonth(userID, date)
	if err != nil {
		h.sendError(w, "Failed to get events", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, map[string]interface{}{
		"events": eventsToDTO(events),
	})
}

// Request/Response types
type CreateEventRequest struct {
	UserID       string `json:"user_id" form:"user_id"`
	Date         string `json:"date" form:"date"`
	Event        string `json:"event" form:"event"`
	ReminderTime string `json:"reminder_time,omitempty" form:"reminder_time"`
}

type UpdateEventRequest struct {
	UserID       string `json:"user_id" form:"user_id"`
	EventID      string `json:"event_id" form:"event_id"`
	Date         string `json:"date" form:"date"`
	Event        string `json:"event" form:"event"`
	ReminderTime string `json:"reminder_time,omitempty" form:"reminder_time"`
}

type DeleteEventRequest struct {
	UserID  string `json:"user_id" form:"user_id"`
	EventID string `json:"event_id" form:"event_id"`
}

type EventDTO struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	Date         string     `json:"date"`
	Text         string     `json:"text"`
	ReminderTime *string    `json:"reminder_time,omitempty"`
	CreatedAt    string     `json:"created_at"`
	UpdatedAt    string     `json:"updated_at"`
}

// decodeRequest is implemented in form_decoder.go

func (h *EventHandler) sendSuccess(w http.ResponseWriter, data map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"result": data,
	})
}

func (h *EventHandler) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
	})
}

func eventsToDTO(events []*domain.Event) []EventDTO {
	dtos := make([]EventDTO, len(events))
	for i, e := range events {
		dtos[i] = EventDTO{
			ID:        e.ID,
			UserID:    e.UserID,
			Date:      e.Date.Format("2006-01-02"),
			Text:      e.Text,
			CreatedAt: e.CreatedAt.Format(time.RFC3339),
			UpdatedAt: e.UpdatedAt.Format(time.RFC3339),
		}
		if e.ReminderTime != nil {
			rt := e.ReminderTime.Format(time.RFC3339)
			dtos[i].ReminderTime = &rt
		}
	}
	return dtos
}

