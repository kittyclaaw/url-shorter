package service

import (
	"context"
	"log"
	"url-shortener/internal/models"
)

type WorkerService struct {
	analyticsService *AnalyticsService
	clickQueue       chan *models.Click
	workerCount      int
}

type ClickData struct {
	URLID     int
	IPAddress string
	UserAgent string
	Referer   string
}

func NewWorkerService(analyticsService *AnalyticsService, workerCount int) *WorkerService {
	ws := &WorkerService{
		analyticsService: analyticsService,
		clickQueue:       make(chan *models.Click, 1000),
		workerCount:      workerCount,
	}

	ws.startWorkers()
	return ws
}

func (ws *WorkerService) ProcessClickAsync(clickData *ClickData) {
	click := &models.Click{
		URLID:     clickData.URLID,
		IPAddress: clickData.IPAddress,
		UserAgent: clickData.UserAgent,
		Referer:   clickData.Referer,
	}

	select {
	case ws.clickQueue <- click:
	default:
		log.Printf("Click queue is full, dropping click for URL ID: %d", clickData.URLID)
	}
}

func (ws *WorkerService) startWorkers() {
	for i := 0; i < ws.workerCount; i++ {
		go ws.worker(i)
	}
}

func (ws *WorkerService) worker(id int) {
	for click := range ws.clickQueue {
		ctx := context.Background()
		if err := ws.analyticsService.SaveClick(ctx, click); err != nil {
			log.Printf("Worker %d: failed to save click: %v", id, err)
		} else {
			log.Printf("Worker %d: successfully processed click for URL ID: %d", id, click.URLID)
		}
	}
}

func (ws *WorkerService) Shutdown() {
	close(ws.clickQueue)
}
