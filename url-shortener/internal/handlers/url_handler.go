// internal/handlers/url_handler.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"url-shortener/internal/service"

	"github.com/gorilla/mux"
)

type URLHandler struct {
	urlService    *service.URLService
	workerService *service.WorkerService
}

func NewURLHandler(urlService *service.URLService, workerService *service.WorkerService) *URLHandler {
	return &URLHandler{
		urlService:    urlService,
		workerService: workerService,
	}
}

func (h *URLHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		json.NewEncoder(w).Encode(map[string]string{"error": "Content-Type must be application/json"})
		return
	}

	var request struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	if request.URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "URL is required"})
		return
	}

	url, err := h.urlService.CreateShortURL(r.Context(), request.URL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"short_url":    "http://localhost:8080/" + url.ShortCode,
		"original_url": url.OriginalURL,
		"short_code":   url.ShortCode,
	})
}

func (h *URLHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	if shortCode == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Short code is required"})
		return
	}

	url, err := h.urlService.GetURL(r.Context(), shortCode)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "URL not found"})
		return
	}

	// Асинхронная обработка клика через воркер
	clickData := &service.ClickData{
		URLID:     url.ID,
		IPAddress: getIPAddress(r),
		UserAgent: r.UserAgent(),
		Referer:   r.Referer(),
	}
	h.workerService.ProcessClickAsync(clickData)

	http.Redirect(w, r, url.OriginalURL, http.StatusFound)
}

func (h *URLHandler) GetURLInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	if shortCode == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Short code is required"})
		return
	}

	url, err := h.urlService.GetURL(r.Context(), shortCode)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "URL not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"short_code":   url.ShortCode,
		"original_url": url.OriginalURL,
		"created_at":   url.CreatedAt,
		"click_count":  url.ClickCount,
	})
}

func getIPAddress(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}
	return r.RemoteAddr
}
