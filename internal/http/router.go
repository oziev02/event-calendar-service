package http

import (
	"net/http"

	"github.com/oziev02/event-calendar-service/internal/http/handlers"
	"github.com/oziev02/event-calendar-service/pkg/logger"
)

// Router настраивает маршруты HTTP сервера
func Router(eventHandler *handlers.EventHandler, log logger.Logger) http.Handler {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/create_event", eventHandler.CreateEvent)
	mux.HandleFunc("/update_event", eventHandler.UpdateEvent)
	mux.HandleFunc("/delete_event", eventHandler.DeleteEvent)
	mux.HandleFunc("/events_for_day", eventHandler.GetEventsForDay)
	mux.HandleFunc("/events_for_week", eventHandler.GetEventsForWeek)
	mux.HandleFunc("/events_for_month", eventHandler.GetEventsForMonth)

	// Применить middleware
	loggingMiddleware := NewLoggingMiddleware(log)
	return loggingMiddleware.Handler(mux)
}

