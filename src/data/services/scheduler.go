package services

import (
	"log"
	"time"

	"api/infra/config"
)

type SchedulerService struct {
	userService  *UserService
	cacheService *CacheService
	config       *config.Config
	stopChan     chan struct{}
	running      bool
}

func NewScheduler(
	userService *UserService,
	cacheService *CacheService,
	cfg *config.Config,
) *SchedulerService {
	return &SchedulerService{
		userService:  userService,
		cacheService: cacheService,
		config:       cfg,
		stopChan:     make(chan struct{}),
	}
}

func (s *SchedulerService) Start() {
	if s.running {
		return
	}

	s.running = true
	go s.run()
	log.Println("Scheduler started")
}

func (s *SchedulerService) Stop() {
	if !s.running {
		return
	}

	close(s.stopChan)
	s.running = false
	log.Println("Scheduler stopped")
}

func (s *SchedulerService) run() {
	ticker := time.NewTicker(1 * time.Hour) // Default interval
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.runScheduledTasks()
		case <-s.stopChan:
			return
		}
	}
}

func (s *SchedulerService) runScheduledTasks() {
	log.Println("Running scheduled tasks...")

	// Add your scheduled tasks here
	// Example:
	// ctx := context.Background()
	// if err := s.cleanupExpiredData(ctx); err != nil {
	//     log.Printf("Error cleaning up expired data: %v", err)
	// }

	log.Println("Scheduled tasks completed")
}

// Add scheduled task methods here
// func (s *SchedulerService) cleanupExpiredData(ctx context.Context) error {
//     return nil
// }
