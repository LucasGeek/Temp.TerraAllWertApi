package dataloader

import (
	"context"
	"fmt"
	"sync"
	"time"

	"api/domain/entities"
	"api/domain/interfaces"
	"api/infra/logger"
	"go.uber.org/zap"
)

// ApartmentLoader provides batched loading of apartments to solve N+1 query problem
type ApartmentLoader struct {
	repo         interfaces.ApartmentRepository
	cache        map[string]*entities.Apartment
	batchSize    int
	waitDuration time.Duration
	mu           sync.RWMutex
	batch        []apartmentRequest
	batchMu      sync.Mutex
}

type apartmentRequest struct {
	id       string
	resultCh chan apartmentResult
}

type apartmentResult struct {
	apartment *entities.Apartment
	err       error
}

func NewApartmentLoader(repo interfaces.ApartmentRepository) *ApartmentLoader {
	loader := &ApartmentLoader{
		repo:         repo,
		cache:        make(map[string]*entities.Apartment),
		batchSize:    50,  // Process up to 50 IDs at once
		waitDuration: 10 * time.Millisecond, // Wait 10ms to batch requests
	}
	
	// Start batch processor
	go loader.processBatches()
	
	return loader
}

// Load loads an apartment by ID, using batching to optimize database calls
func (l *ApartmentLoader) Load(ctx context.Context, id string) (*entities.Apartment, error) {
	// Check cache first
	l.mu.RLock()
	if apartment, exists := l.cache[id]; exists {
		l.mu.RUnlock()
		logger.Debug(ctx, "Apartment loaded from cache", zap.String("apartment_id", id))
		return apartment, nil
	}
	l.mu.RUnlock()
	
	// Create request channel
	resultCh := make(chan apartmentResult, 1)
	request := apartmentRequest{
		id:       id,
		resultCh: resultCh,
	}
	
	// Add to batch
	l.batchMu.Lock()
	l.batch = append(l.batch, request)
	shouldProcess := len(l.batch) >= l.batchSize
	l.batchMu.Unlock()
	
	// Process immediately if batch is full
	if shouldProcess {
		go l.processBatch()
	}
	
	// Wait for result
	select {
	case result := <-resultCh:
		return result.apartment, result.err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Second): // Timeout
		return nil, context.DeadlineExceeded
	}
}

// LoadMany loads multiple apartments by IDs efficiently
func (l *ApartmentLoader) LoadMany(ctx context.Context, ids []string) ([]*entities.Apartment, error) {
	results := make([]*entities.Apartment, len(ids))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstError error
	
	for i, id := range ids {
		wg.Add(1)
		go func(index int, apartmentID string) {
			defer wg.Done()
			
			apartment, err := l.Load(ctx, apartmentID)
			
			mu.Lock()
			defer mu.Unlock()
			
			if err != nil && firstError == nil {
				firstError = err
			}
			results[index] = apartment
		}(i, id)
	}
	
	wg.Wait()
	
	if firstError != nil {
		return nil, firstError
	}
	
	return results, nil
}

// ClearCache clears the loader's cache
func (l *ApartmentLoader) ClearCache() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache = make(map[string]*entities.Apartment)
}

// InvalidateCache removes a specific apartment from cache
func (l *ApartmentLoader) InvalidateCache(id string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.cache, id)
}

// processBatches runs in a goroutine to process batches periodically
func (l *ApartmentLoader) processBatches() {
	ticker := time.NewTicker(l.waitDuration)
	defer ticker.Stop()
	
	for range ticker.C {
		l.processBatch()
	}
}

// processBatch processes the current batch of requests
func (l *ApartmentLoader) processBatch() {
	l.batchMu.Lock()
	if len(l.batch) == 0 {
		l.batchMu.Unlock()
		return
	}
	
	// Take the current batch
	batch := l.batch
	l.batch = nil
	l.batchMu.Unlock()
	
	// Extract unique IDs
	idMap := make(map[string][]apartmentRequest)
	for _, req := range batch {
		idMap[req.id] = append(idMap[req.id], req)
	}
	
	ids := make([]string, 0, len(idMap))
	for id := range idMap {
		ids = append(ids, id)
	}
	
	ctx := context.Background()
	logger.Debug(ctx, "Processing apartment batch", zap.Int("batch_size", len(ids)))
	
	// Load apartments that are not in cache
	toLoad := make([]string, 0)
	l.mu.RLock()
	for _, id := range ids {
		if _, exists := l.cache[id]; !exists {
			toLoad = append(toLoad, id)
		}
	}
	l.mu.RUnlock()
	
	// Batch load from database
	if len(toLoad) > 0 {
		apartments, err := l.batchLoadFromDB(ctx, toLoad)
		if err != nil {
			// Send error to all waiting requests
			for _, requests := range idMap {
				for _, req := range requests {
					req.resultCh <- apartmentResult{nil, err}
				}
			}
			return
		}
		
		// Update cache
		l.mu.Lock()
		for _, apartment := range apartments {
			if apartment != nil {
				l.cache[apartment.ID] = apartment
			}
		}
		l.mu.Unlock()
	}
	
	// Send results to waiting requests
	l.mu.RLock()
	for id, requests := range idMap {
		apartment := l.cache[id]
		var err error
		if apartment == nil {
			err = ErrApartmentNotFound
		}
		
		for _, req := range requests {
			req.resultCh <- apartmentResult{apartment, err}
		}
	}
	l.mu.RUnlock()
}

// batchLoadFromDB loads multiple apartments from database
func (l *ApartmentLoader) batchLoadFromDB(ctx context.Context, ids []string) ([]*entities.Apartment, error) {
	apartments := make([]*entities.Apartment, 0, len(ids))
	
	// Load apartments individually (could be optimized with batch query)
	for _, id := range ids {
		apartment, err := l.repo.GetByID(ctx, id)
		if err != nil {
			logger.Warn(ctx, "Failed to load apartment", zap.String("apartment_id", id), zap.Error(err))
			continue // Skip failed loads
		}
		apartments = append(apartments, apartment)
	}
	
	logger.Debug(ctx, "Loaded apartments from database", 
		zap.Int("requested", len(ids)),
		zap.Int("loaded", len(apartments)),
	)
	
	return apartments, nil
}

// Prime adds an apartment to the cache
func (l *ApartmentLoader) Prime(apartment *entities.Apartment) {
	if apartment == nil {
		return
	}
	
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache[apartment.ID] = apartment
}

// TowerApartmentLoader loads apartments for specific towers
type TowerApartmentLoader struct {
	repo         interfaces.ApartmentRepository
	cache        map[string][]*entities.Apartment
	batchSize    int
	waitDuration time.Duration
	mu           sync.RWMutex
	batch        []towerApartmentRequest
	batchMu      sync.Mutex
}

type towerApartmentRequest struct {
	floorID  string
	resultCh chan towerApartmentResult
}

type towerApartmentResult struct {
	apartments []*entities.Apartment
	err        error
}

func NewTowerApartmentLoader(repo interfaces.ApartmentRepository) *TowerApartmentLoader {
	loader := &TowerApartmentLoader{
		repo:         repo,
		cache:        make(map[string][]*entities.Apartment),
		batchSize:    20,
		waitDuration: 15 * time.Millisecond,
	}
	
	go loader.processBatches()
	return loader
}

// LoadByFloorID loads apartments for a specific floor
func (l *TowerApartmentLoader) LoadByFloorID(ctx context.Context, floorID string) ([]*entities.Apartment, error) {
	// Check cache first
	l.mu.RLock()
	if apartments, exists := l.cache[floorID]; exists {
		l.mu.RUnlock()
		logger.Debug(ctx, "Apartments loaded from cache", zap.String("floor_id", floorID))
		return apartments, nil
	}
	l.mu.RUnlock()
	
	// Create request
	resultCh := make(chan towerApartmentResult, 1)
	request := towerApartmentRequest{
		floorID:  floorID,
		resultCh: resultCh,
	}
	
	// Add to batch
	l.batchMu.Lock()
	l.batch = append(l.batch, request)
	shouldProcess := len(l.batch) >= l.batchSize
	l.batchMu.Unlock()
	
	if shouldProcess {
		go l.processBatch()
	}
	
	// Wait for result
	select {
	case result := <-resultCh:
		return result.apartments, result.err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Second):
		return nil, context.DeadlineExceeded
	}
}

func (l *TowerApartmentLoader) processBatches() {
	ticker := time.NewTicker(l.waitDuration)
	defer ticker.Stop()
	
	for range ticker.C {
		l.processBatch()
	}
}

func (l *TowerApartmentLoader) processBatch() {
	l.batchMu.Lock()
	if len(l.batch) == 0 {
		l.batchMu.Unlock()
		return
	}
	
	batch := l.batch
	l.batch = nil
	l.batchMu.Unlock()
	
	// Group requests by floor ID
	floorMap := make(map[string][]towerApartmentRequest)
	for _, req := range batch {
		floorMap[req.floorID] = append(floorMap[req.floorID], req)
	}
	
	ctx := context.Background()
	logger.Debug(ctx, "Processing tower apartment batch", zap.Int("floors", len(floorMap)))
	
	// Load apartments for each floor
	for floorID, requests := range floorMap {
		// Check cache
		l.mu.RLock()
		apartments, exists := l.cache[floorID]
		l.mu.RUnlock()
		
		var err error
		if !exists {
			// Load from database
			apartments, err = l.repo.GetByFloorID(ctx, floorID)
			if err == nil {
				// Update cache
				l.mu.Lock()
				l.cache[floorID] = apartments
				l.mu.Unlock()
			}
		}
		
		// Send results
		for _, req := range requests {
			req.resultCh <- towerApartmentResult{apartments, err}
		}
	}
}

// ClearFloorCache clears cache for a specific floor
func (l *TowerApartmentLoader) ClearFloorCache(floorID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.cache, floorID)
}

// Errors
var (
	ErrApartmentNotFound = fmt.Errorf("apartment not found")
)