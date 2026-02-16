package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"orderfc/models"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type OrderRepository struct {
	Database    *gorm.DB
	Redis       *redis.Client
	ProductHost string
}

func NewOrderRepository(db *gorm.DB, redis *redis.Client, productHost string) *OrderRepository {
	return &OrderRepository{
		Database:    db,
		Redis:       redis,
		ProductHost: productHost,
	}
}

func (r *OrderRepository) GetProductInfo(ctx context.Context, productID int64) (models.Product, error) {

	url := fmt.Sprintf("%s/v1/products/%d", r.ProductHost, productID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return models.Product{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return models.Product{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.Product{}, errors.New("failed to get product info")
	}

	var product models.Product
	err = json.NewDecoder(resp.Body).Decode(&product)
	if err != nil {
		return models.Product{}, err
	}

	return product, nil
}

