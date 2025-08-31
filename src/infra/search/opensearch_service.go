package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"go.uber.org/zap"

	"api/domain/entities"
	"api/infra/logger"
)

type OpenSearchService struct {
	client    *opensearch.Client
	indexName string
}

type SearchConfig struct {
	Host      string
	Username  string
	Password  string
	IndexName string
}

// ApartmentDocument represents the apartment document structure in OpenSearch
type ApartmentDocument struct {
	ID            string  `json:"id"`
	Number        string  `json:"number"`
	FloorID       string  `json:"floor_id"`
	TowerID       string  `json:"tower_id"`
	TowerName     string  `json:"tower_name"`
	Area          string  `json:"area,omitempty"`
	Suites        int     `json:"suites,omitempty"`
	Bedrooms      int     `json:"bedrooms,omitempty"`
	ParkingSpots  int     `json:"parking_spots,omitempty"`
	Status        string  `json:"status"`
	Price         float64 `json:"price,omitempty"`
	Available     bool    `json:"available"`
	SolarPosition string  `json:"solar_position,omitempty"`
	Description   string  `json:"description,omitempty"`
	IndexedAt     string  `json:"indexed_at"`
}

// SearchQuery represents a search query structure
type SearchQuery struct {
	Query           string            `json:"query,omitempty"`
	Filters         map[string]string `json:"filters,omitempty"`
	PriceRange      *PriceRange       `json:"price_range,omitempty"`
	BedroomsRange   *IntRange         `json:"bedrooms_range,omitempty"`
	SuitesRange     *IntRange         `json:"suites_range,omitempty"`
	Available       *bool             `json:"available,omitempty"`
	TowerIDs        []string          `json:"tower_ids,omitempty"`
	SolarPositions  []string          `json:"solar_positions,omitempty"`
	Limit           int               `json:"limit,omitempty"`
	Offset          int               `json:"offset,omitempty"`
}

type PriceRange struct {
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`
}

type IntRange struct {
	Min *int `json:"min,omitempty"`
	Max *int `json:"max,omitempty"`
}

// SearchResult represents search results
type SearchResult struct {
	Total      int64                  `json:"total"`
	Apartments []*ApartmentDocument   `json:"apartments"`
	Took       int                    `json:"took"`
	TimedOut   bool                   `json:"timed_out"`
}

func NewOpenSearchService(config SearchConfig) (*OpenSearchService, error) {
	cfg := opensearch.Config{
		Addresses: []string{config.Host},
		Username:  config.Username,
		Password:  config.Password,
	}

	client, err := opensearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenSearch client: %w", err)
	}

	service := &OpenSearchService{
		client:    client,
		indexName: config.IndexName,
	}

	return service, nil
}

// CreateIndex creates the apartments index with proper mapping
func (s *OpenSearchService) CreateIndex(ctx context.Context) error {
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id":             map[string]interface{}{"type": "keyword"},
				"number":         map[string]interface{}{"type": "text", "analyzer": "standard"},
				"floor_id":       map[string]interface{}{"type": "keyword"},
				"tower_id":       map[string]interface{}{"type": "keyword"},
				"tower_name":     map[string]interface{}{"type": "text", "analyzer": "standard"},
				"area":           map[string]interface{}{"type": "text"},
				"suites":         map[string]interface{}{"type": "integer"},
				"bedrooms":       map[string]interface{}{"type": "integer"},
				"parking_spots":  map[string]interface{}{"type": "integer"},
				"status":         map[string]interface{}{"type": "keyword"},
				"price":          map[string]interface{}{"type": "double"},
				"available":      map[string]interface{}{"type": "boolean"},
				"solar_position": map[string]interface{}{"type": "keyword"},
				"description":    map[string]interface{}{"type": "text", "analyzer": "standard"},
				"indexed_at":     map[string]interface{}{"type": "date", "format": "strict_date_time"},
			},
		},
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 1,
		},
	}

	body, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal index mapping: %w", err)
	}

	req := opensearchapi.IndicesCreateRequest{
		Index: s.indexName,
		Body:  bytes.NewReader(body),
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to create index: %s", res.Status())
	}

	logger.Info(ctx, "OpenSearch index created successfully", zap.String("index", s.indexName))
	return nil
}

// IndexApartment indexes a single apartment document
func (s *OpenSearchService) IndexApartment(ctx context.Context, apartment *entities.Apartment, towerName string) error {
	doc := s.apartmentToDocument(apartment, towerName)

	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal apartment document: %w", err)
	}

	req := opensearchapi.IndexRequest{
		Index:      s.indexName,
		DocumentID: apartment.ID,
		Body:       bytes.NewReader(body),
		Refresh:    "wait_for",
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to index apartment: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to index apartment: %s", res.Status())
	}

	logger.Debug(ctx, "Apartment indexed successfully", zap.String("apartment_id", apartment.ID))
	return nil
}

// BulkIndexApartments indexes multiple apartments in a single request
func (s *OpenSearchService) BulkIndexApartments(ctx context.Context, apartments []*entities.Apartment, towerNames map[string]string) error {
	if len(apartments) == 0 {
		return nil
	}

	var buf bytes.Buffer

	for _, apartment := range apartments {
		// Action header
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": s.indexName,
				"_id":    apartment.ID,
			},
		}
		actionBytes, _ := json.Marshal(action)
		buf.Write(actionBytes)
		buf.WriteString("\n")

		// Document
		towerName := towerNames[apartment.FloorID] // Map floor to tower name
		doc := s.apartmentToDocument(apartment, towerName)
		docBytes, _ := json.Marshal(doc)
		buf.Write(docBytes)
		buf.WriteString("\n")
	}

	req := opensearchapi.BulkRequest{
		Body:    &buf,
		Refresh: "wait_for",
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to bulk index apartments: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to bulk index apartments: %s", res.Status())
	}

	logger.Info(ctx, "Apartments bulk indexed successfully", zap.Int("count", len(apartments)))
	return nil
}

// SearchApartments performs a full-text search on apartments
func (s *OpenSearchService) SearchApartments(ctx context.Context, query SearchQuery) (*SearchResult, error) {
	searchQuery := s.buildSearchQuery(query)

	body, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search query: %w", err)
	}

	req := opensearchapi.SearchRequest{
		Index: []string{s.indexName},
		Body:  bytes.NewReader(body),
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search failed: %s", res.Status())
	}

	var searchResult map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	result := s.parseSearchResponse(searchResult)
	
	logger.Debug(ctx, "Search completed", 
		zap.Int64("total", result.Total),
		zap.Int("returned", len(result.Apartments)),
		zap.Int("took", result.Took),
	)

	return result, nil
}

// DeleteApartment removes an apartment from the search index
func (s *OpenSearchService) DeleteApartment(ctx context.Context, apartmentID string) error {
	req := opensearchapi.DeleteRequest{
		Index:      s.indexName,
		DocumentID: apartmentID,
		Refresh:    "wait_for",
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to delete apartment: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("failed to delete apartment: %s", res.Status())
	}

	logger.Debug(ctx, "Apartment deleted from search index", zap.String("apartment_id", apartmentID))
	return nil
}

// apartmentToDocument converts an apartment entity to a search document
func (s *OpenSearchService) apartmentToDocument(apartment *entities.Apartment, towerName string) *ApartmentDocument {
	doc := &ApartmentDocument{
		ID:        apartment.ID,
		Number:    apartment.Number,
		FloorID:   apartment.FloorID,
		TowerName: towerName,
		Status:    string(apartment.Status),
		Available: apartment.Available,
		IndexedAt: time.Now().UTC().Format(time.RFC3339),
	}

	if apartment.Area != nil {
		doc.Area = *apartment.Area
	}
	if apartment.Suites != nil {
		doc.Suites = *apartment.Suites
	}
	if apartment.Bedrooms != nil {
		doc.Bedrooms = *apartment.Bedrooms
	}
	if apartment.ParkingSpots != nil {
		doc.ParkingSpots = *apartment.ParkingSpots
	}
	if apartment.Price != nil {
		doc.Price = *apartment.Price
	}
	if apartment.SolarPosition != nil {
		doc.SolarPosition = *apartment.SolarPosition
	}

	return doc
}

// buildSearchQuery constructs an OpenSearch query from SearchQuery
func (s *OpenSearchService) buildSearchQuery(query SearchQuery) map[string]interface{} {
	searchQuery := map[string]interface{}{
		"size": 50, // Default size
		"from": 0,  // Default offset
	}

	if query.Limit > 0 {
		searchQuery["size"] = query.Limit
	}
	if query.Offset > 0 {
		searchQuery["from"] = query.Offset
	}

	// Build query clause
	boolQuery := map[string]interface{}{
		"bool": map[string]interface{}{},
	}

	// Text search
	if query.Query != "" {
		mustClauses := []interface{}{
			map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":  query.Query,
					"fields": []string{"number^2", "tower_name^1.5", "description"},
					"type":   "best_fields",
					"fuzziness": "AUTO",
				},
			},
		}
		boolQuery["bool"].(map[string]interface{})["must"] = mustClauses
	}

	// Filters
	var filterClauses []interface{}

	// Availability filter
	if query.Available != nil {
		filterClauses = append(filterClauses, map[string]interface{}{
			"term": map[string]interface{}{
				"available": *query.Available,
			},
		})
	}

	// Tower IDs filter
	if len(query.TowerIDs) > 0 {
		filterClauses = append(filterClauses, map[string]interface{}{
			"terms": map[string]interface{}{
				"tower_id": query.TowerIDs,
			},
		})
	}

	// Solar positions filter
	if len(query.SolarPositions) > 0 {
		filterClauses = append(filterClauses, map[string]interface{}{
			"terms": map[string]interface{}{
				"solar_position": query.SolarPositions,
			},
		})
	}

	// Price range filter
	if query.PriceRange != nil {
		rangeFilter := map[string]interface{}{}
		if query.PriceRange.Min != nil {
			rangeFilter["gte"] = *query.PriceRange.Min
		}
		if query.PriceRange.Max != nil {
			rangeFilter["lte"] = *query.PriceRange.Max
		}
		if len(rangeFilter) > 0 {
			filterClauses = append(filterClauses, map[string]interface{}{
				"range": map[string]interface{}{
					"price": rangeFilter,
				},
			})
		}
	}

	// Bedrooms range filter
	if query.BedroomsRange != nil {
		rangeFilter := map[string]interface{}{}
		if query.BedroomsRange.Min != nil {
			rangeFilter["gte"] = *query.BedroomsRange.Min
		}
		if query.BedroomsRange.Max != nil {
			rangeFilter["lte"] = *query.BedroomsRange.Max
		}
		if len(rangeFilter) > 0 {
			filterClauses = append(filterClauses, map[string]interface{}{
				"range": map[string]interface{}{
					"bedrooms": rangeFilter,
				},
			})
		}
	}

	// Suites range filter
	if query.SuitesRange != nil {
		rangeFilter := map[string]interface{}{}
		if query.SuitesRange.Min != nil {
			rangeFilter["gte"] = *query.SuitesRange.Min
		}
		if query.SuitesRange.Max != nil {
			rangeFilter["lte"] = *query.SuitesRange.Max
		}
		if len(rangeFilter) > 0 {
			filterClauses = append(filterClauses, map[string]interface{}{
				"range": map[string]interface{}{
					"suites": rangeFilter,
				},
			})
		}
	}

	// Add filters to bool query
	if len(filterClauses) > 0 {
		boolQuery["bool"].(map[string]interface{})["filter"] = filterClauses
	}

	// Only add query if there are conditions
	if len(boolQuery["bool"].(map[string]interface{})) > 0 {
		searchQuery["query"] = boolQuery
	} else {
		// Match all if no specific conditions
		searchQuery["query"] = map[string]interface{}{
			"match_all": map[string]interface{}{},
		}
	}

	// Sort by relevance score and price
	searchQuery["sort"] = []interface{}{
		map[string]interface{}{"_score": map[string]interface{}{"order": "desc"}},
		map[string]interface{}{"price": map[string]interface{}{"order": "asc", "missing": "_last"}},
	}

	return searchQuery
}

// parseSearchResponse parses OpenSearch response into SearchResult
func (s *OpenSearchService) parseSearchResponse(response map[string]interface{}) *SearchResult {
	result := &SearchResult{
		Apartments: make([]*ApartmentDocument, 0),
	}

	if took, ok := response["took"].(float64); ok {
		result.Took = int(took)
	}

	if timedOut, ok := response["timed_out"].(bool); ok {
		result.TimedOut = timedOut
	}

	hits, ok := response["hits"].(map[string]interface{})
	if !ok {
		return result
	}

	if total, ok := hits["total"].(map[string]interface{})["value"].(float64); ok {
		result.Total = int64(total)
	}

	if hitsList, ok := hits["hits"].([]interface{}); ok {
		for _, hit := range hitsList {
			if hitMap, ok := hit.(map[string]interface{}); ok {
				if source, ok := hitMap["_source"].(map[string]interface{}); ok {
					apartment := &ApartmentDocument{}
					
					// Parse source fields
					if id, ok := source["id"].(string); ok {
						apartment.ID = id
					}
					if number, ok := source["number"].(string); ok {
						apartment.Number = number
					}
					if floorID, ok := source["floor_id"].(string); ok {
						apartment.FloorID = floorID
					}
					if towerName, ok := source["tower_name"].(string); ok {
						apartment.TowerName = towerName
					}
					if area, ok := source["area"].(string); ok {
						apartment.Area = area
					}
					if suites, ok := source["suites"].(float64); ok {
						apartment.Suites = int(suites)
					}
					if bedrooms, ok := source["bedrooms"].(float64); ok {
						apartment.Bedrooms = int(bedrooms)
					}
					if parkingSpots, ok := source["parking_spots"].(float64); ok {
						apartment.ParkingSpots = int(parkingSpots)
					}
					if status, ok := source["status"].(string); ok {
						apartment.Status = status
					}
					if price, ok := source["price"].(float64); ok {
						apartment.Price = price
					}
					if available, ok := source["available"].(bool); ok {
						apartment.Available = available
					}
					if solarPosition, ok := source["solar_position"].(string); ok {
						apartment.SolarPosition = solarPosition
					}
					if description, ok := source["description"].(string); ok {
						apartment.Description = description
					}

					result.Apartments = append(result.Apartments, apartment)
				}
			}
		}
	}

	return result
}

// HealthCheck checks if OpenSearch is accessible
func (s *OpenSearchService) HealthCheck(ctx context.Context) error {
	req := opensearchapi.CatHealthRequest{}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("OpenSearch health check failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("OpenSearch health check failed: %s", res.Status())
	}

	return nil
}

// GetStats returns search statistics
func (s *OpenSearchService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	req := opensearchapi.IndicesStatsRequest{
		Index: []string{s.indexName},
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, fmt.Errorf("failed to get search stats: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("failed to get search stats: %s", res.Status())
	}

	var stats map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode stats response: %w", err)
	}

	return stats, nil
}