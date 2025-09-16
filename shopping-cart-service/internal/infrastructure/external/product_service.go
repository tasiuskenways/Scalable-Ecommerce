package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type ProductServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

type ProductResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Stock       int      `json:"stock"`
	CategoryID  string   `json:"category_id"`
	Category    Category `json:"category"`
	SKU         string   `json:"sku"`
	IsActive    bool     `json:"is_active"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

type Category struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

type ServiceResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Error   string          `json:"error,omitempty"`
}

func NewProductServiceClient(baseURL string) *ProductServiceClient {
	return &ProductServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *ProductServiceClient) GetProducts(ctx context.Context, productIDs []string) ([]*ProductResponse, error) {
	url := fmt.Sprintf("%s/api/products/ids", c.baseURL)

	payload, err := json.Marshal(map[string][]string{"ids": productIDs})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	log.Printf("Payload: %s", string(payload))

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch product: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("product service returned status %d", resp.StatusCode)
	}

	var serviceResp ServiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&serviceResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !serviceResp.Success {
		return nil, fmt.Errorf("product service error: %s", serviceResp.Error)
	}

	var product []*ProductResponse
	if err := json.Unmarshal(serviceResp.Data, &product); err != nil {
		return nil, fmt.Errorf("failed to decode product data: %w", err)
	}

	return product, nil
}

func (c *ProductServiceClient) GetProduct(ctx context.Context, productID string) (*ProductResponse, error) {
	url := fmt.Sprintf("%s/api/products/%s", c.baseURL, productID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch product: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("product service returned status %d", resp.StatusCode)
	}

	var serviceResp ServiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&serviceResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !serviceResp.Success {
		return nil, fmt.Errorf("product service error: %s", serviceResp.Error)
	}

	var product ProductResponse
	if err := json.Unmarshal(serviceResp.Data, &product); err != nil {
		return nil, fmt.Errorf("failed to decode product data: %w", err)
	}

	return &product, nil
}

func (c *ProductServiceClient) CheckStock(ctx context.Context, productID string, quantity int) (bool, error) {
	product, err := c.GetProduct(ctx, productID)
	if err != nil {
		return false, err
	}

	return product.Stock >= quantity && product.IsActive, nil
}
