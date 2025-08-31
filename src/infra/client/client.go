package client

import (
	"net/http"
	"time"

	"api/infra/config"
)

type ExternalClient struct {
	httpClient *http.Client
	config     *config.Config
}

func NewExternalClient(cfg *config.Config) *ExternalClient {
	return &ExternalClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: cfg,
	}
}

// Add your external API calls here
// Example:
// func (c *ExternalClient) GetData(ctx context.Context) (*Data, error) {
//     // Implementation here
//     return nil, nil
// }