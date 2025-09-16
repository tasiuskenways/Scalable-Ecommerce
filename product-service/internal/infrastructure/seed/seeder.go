package seed

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/tasiuskenways/scalable-ecommerce/product-service/internal/domain/entities"
	"gorm.io/gorm"
)

type Seeder struct {
	db *gorm.DB
}

func NewSeeder(db *gorm.DB) *Seeder {
	return &Seeder{db: db}
}

func (s *Seeder) Seed() error {
	log.Println("Starting database seeding...")
	start := time.Now()

	// Clear existing data
	if err := s.clearData(); err != nil {
		return fmt.Errorf("failed to clear existing data: %w", err)
	}

	// Seed categories
	categories, err := s.seedCategories(5) // Seed 5 categories
	if err != nil {
		return fmt.Errorf("failed to seed categories: %w", err)
	}

	// Seed products
	if err := s.seedProducts(categories, 20); err != nil { // Seed 20 products
		return fmt.Errorf("failed to seed products: %w", err)
	}

	log.Printf("Database seeding completed in %v\n", time.Since(start))
	return nil
}

func (s *Seeder) clearData() error {
	// Disable foreign key checks (PostgreSQL specific)
	if err := s.db.Exec("SET session_replication_role = 'replica';").Error; err != nil {
		return err
	}

	// Clear tables
	tables := []string{"products", "categories"}
	for _, table := range tables {
		if err := s.db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", table)).Error; err != nil {
			// Re-enable foreign key checks if there's an error
			s.db.Exec("SET session_replication_role = 'origin';")
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	// Re-enable foreign key checks
	if err := s.db.Exec("SET session_replication_role = 'origin';").Error; err != nil {
		return err
	}

	return nil
}

// generateRandomString generates a random string of length n
func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (s *Seeder) seedCategories(count int) ([]*entities.Category, error) {
	var categories []*entities.Category
	categoryNames := []string{
		"Electronics",
		"Clothing",
		"Home & Kitchen",
		"Books",
		"Sports & Outdoors",
	}

	descriptions := []string{
		"Latest gadgets and electronic devices",
		"Fashionable clothing for all occasions",
		"Everything you need for your home",
		"Books for all ages and interests",
		"Equipment for sports and outdoor activities",
	}

	for i := 0; i < count && i < len(categoryNames); i++ {
		category := &entities.Category{
			Name:        categoryNames[i],
			Description: descriptions[i],
			IsActive:    true,
		}

		if err := s.db.Create(category).Error; err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	return categories, nil
}

func (s *Seeder) seedProducts(categories []*entities.Category, count int) error {
	products := []*entities.Product{
		{
			Name:        "Smartphone X",
			Description: "Latest smartphone with advanced features",
			Price:       999.99,
			Stock:       100,
			SKU:         "SMARTPHONE-X",
			IsActive:    true,
			CategoryID:  categories[0].ID,
		},
		{
			Name:        "Laptop Pro",
			Description: "High-performance laptop for professionals",
			Price:       1499.99,
			Stock:       50,
			SKU:         "LAPTOP-PRO",
			IsActive:    true,
			CategoryID:  categories[0].ID,
		},
		{
			Name:        "Wireless Earbuds",
			Description: "Noise-canceling wireless earbuds",
			Price:       199.99,
			Stock:       200,
			SKU:         "EARBUDS-WL",
			IsActive:    true,
			CategoryID:  categories[0].ID,
		},
		{
			Name:        "Casual T-Shirt",
			Description: "Comfortable cotton t-shirt",
			Price:       29.99,
			Stock:       500,
			SKU:         "TSHIRT-001",
			IsActive:    true,
			CategoryID:  categories[1].ID,
		},
		{
			Name:        "Denim Jeans",
			Description: "Classic fit denim jeans",
			Price:       59.99,
			Stock:       300,
			SKU:         "JEANS-001",
			IsActive:    true,
			CategoryID:  categories[1].ID,
		},
	}

	// Add more random products if needed
	for i := len(products); i < count; i++ {
		category := categories[rand.Intn(len(categories))]
		price := 10.0 + rand.Float64()*990 // Random price between 10 and 1000
		stock := rand.Intn(500) + 1        // Random stock between 1 and 500

		product := &entities.Product{
			Name:        fmt.Sprintf("Product %s", generateRandomString(5)),
			Description: fmt.Sprintf("Description for product %d", i+1),
			Price:       price,
			Stock:       stock,
			SKU:         fmt.Sprintf("SKU-%s-%04d", generateRandomString(3), i+1),
			IsActive:    rand.Float32() > 0.1, // 90% chance of being active
			CategoryID:  category.ID,
		}

		products = append(products, product)
	}

	// Batch insert products
	return s.db.CreateInBatches(products, 100).Error
}

// SeedData is a convenience function to seed the database with sample data
func SeedData(db *gorm.DB) error {
	rand.Seed(time.Now().UnixNano())
	return NewSeeder(db).Seed()
}
