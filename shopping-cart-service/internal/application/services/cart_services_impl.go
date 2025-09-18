package services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"
	"github.com/tasiuskenways/scalable-ecommerce/shopping-cart-service/internal/application/dto"
	"github.com/tasiuskenways/scalable-ecommerce/shopping-cart-service/internal/config"
	"github.com/tasiuskenways/scalable-ecommerce/shopping-cart-service/internal/domain/entities"
	"github.com/tasiuskenways/scalable-ecommerce/shopping-cart-service/internal/domain/repositories"
	"github.com/tasiuskenways/scalable-ecommerce/shopping-cart-service/internal/domain/services"
	"github.com/tasiuskenways/scalable-ecommerce/shopping-cart-service/internal/infrastructure/external"
)

type cartService struct {
	cartRepo       repositories.CartRepository
	cartItemRepo   repositories.CartItemRepository
	productService *external.ProductServiceClient
	config         *config.Config
}

func NewCartService(
	cartRepo repositories.CartRepository,
	cartItemRepo repositories.CartItemRepository,
	productService *external.ProductServiceClient,
	config *config.Config,
) services.CartService {
	return &cartService{
		cartRepo:       cartRepo,
		cartItemRepo:   cartItemRepo,
		productService: productService,
		config:         config,
	}
}

func (s *cartService) GetCart(ctx *fiber.Ctx, userID string) (*dto.CartResponse, error) {
	// Get cart from database
	cart, err := s.cartRepo.GetByUserID(ctx.Context(), userID)
	if err != nil {
		return nil, err
	}

	if cart == nil {
		// Return empty cart if not found
		return &dto.CartResponse{
			UserID:     userID,
			Items:      []dto.CartItemResponse{},
			Stores:     []dto.StoreCartItems{},
			TotalItems: 0,
			TotalPrice: decimal.NewFromFloat(0),
		}, nil
	}

	// Get cart items
	items, err := s.cartItemRepo.GetByCartID(ctx.Context(), cart.ID)
	if err != nil {
		return nil, err
	}

	// Fetch product details for all items
	cartResponse := &dto.CartResponse{
		ID:         cart.ID,
		UserID:     cart.UserID,
		Items:      []dto.CartItemResponse{},
		Stores:     []dto.StoreCartItems{},
		TotalItems: 0,
		TotalPrice: decimal.NewFromFloat(0),
		CreatedAt:  cart.CreatedAt,
		UpdatedAt:  cart.UpdatedAt,
	}

	productIDs := make([]string, len(items))
	for i, item := range items {
		productIDs[i] = item.ProductID
	}

	if len(productIDs) == 0 {
		return cartResponse, nil
	}

	products, err := s.productService.GetProducts(ctx.Context(), productIDs)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		// Find product details
		var product *external.ProductResponse
		for _, p := range products {
			if p.ID == item.ProductID {
				product = p
				break
			}
		}

		// If product not found, mark as unavailable
		if product == nil {
			cartItem := dto.CartItemResponse{
				ID:          item.ID,
				ProductID:   item.ProductID,
				Quantity:    item.Quantity,
				PriceAtTime: item.PriceAtTime,
				Subtotal:    item.GetSubtotal(),
				Available:   false,
				StockStatus: "Product not found",
			}
			cartResponse.Items = append(cartResponse.Items, cartItem)
			continue
		}

		// Check stock availability
		available := product.Stock >= item.Quantity
		stockStatus := "In stock"
		if !available {
			if product.Stock > 0 {
				stockStatus = fmt.Sprintf("Only %d available", product.Stock)
			} else {
				stockStatus = "Out of stock"
			}
		}
		if !product.IsActive {
			available = false
			stockStatus = "Product unavailable"
		}

		cartItem := dto.CartItemResponse{
			ID:          item.ID,
			ProductID:   item.ProductID,
			Quantity:    item.Quantity,
			PriceAtTime: item.PriceAtTime,
			Subtotal:    item.GetSubtotal(),
			Product: &dto.ProductInfo{
				Name:     product.Name,
				Price:    decimal.NewFromFloat(product.Price),
				Category: product.Category.Name,
				SKU:      product.SKU,
				IsActive: product.IsActive,
				Stock:    product.Stock,
				StoreID:  product.StoreID,
			},
			Available:   available,
			StockStatus: stockStatus,
		}

		cartResponse.Items = append(cartResponse.Items, cartItem)
		cartResponse.TotalItems += item.Quantity
		cartResponse.TotalPrice = cartResponse.TotalPrice.Add(item.GetSubtotal())
	}

	// Group items by store
	storeGroups := make(map[string]*dto.StoreCartItems)
	for _, item := range cartResponse.Items {
		if item.Product != nil {
			storeID := item.Product.StoreID
			if storeGroup, exists := storeGroups[storeID]; exists {
				storeGroup.Items = append(storeGroup.Items, item)
				storeGroup.ItemCount += item.Quantity
				storeGroup.StoreTotal = storeGroup.StoreTotal.Add(item.Subtotal)
			} else {
				storeGroups[storeID] = &dto.StoreCartItems{
					StoreID:    storeID,
					Items:      []dto.CartItemResponse{item},
					ItemCount:  item.Quantity,
					StoreTotal: item.Subtotal,
				}
			}
		}
	}

	// Convert map to slice
	for _, storeGroup := range storeGroups {
		cartResponse.Stores = append(cartResponse.Stores, *storeGroup)
	}

	return cartResponse, nil
}

func (s *cartService) AddItemToCart(ctx *fiber.Ctx, userID string, req *dto.AddItemRequest) (*dto.CartResponse, error) {
	// Validate product and check stock
	product, err := s.productService.GetProduct(ctx.Context(), req.ProductID)
	if err != nil {
		return nil, errors.New("product not found")
	}

	if !product.IsActive {
		return nil, errors.New("product is not available")
	}

	if product.Stock < req.Quantity {
		return nil, fmt.Errorf("insufficient stock. Only %d available", product.Stock)
	}

	// Get or create cart
	cart, err := s.cartRepo.GetByUserID(ctx.Context(), userID)
	if err != nil {
		return nil, err
	}

	if cart == nil {
		// Create new cart
		cart = &entities.Cart{
			UserID: userID,
		}
		if err := s.cartRepo.Create(ctx.Context(), cart); err != nil {
			return nil, err
		}
	}

	// Check if item already exists in cart
	existingItem, err := s.cartItemRepo.GetByCartAndProduct(ctx.Context(), cart.ID, req.ProductID)
	if err != nil {
		return nil, err
	}

	if existingItem != nil {
		// Update quantity
		newQuantity := existingItem.Quantity + req.Quantity
		if product.Stock < newQuantity {
			return nil, fmt.Errorf("insufficient stock. Only %d available", product.Stock)
		}
		existingItem.Quantity = newQuantity
		existingItem.PriceAtTime = decimal.NewFromFloat(product.Price)
		if err := s.cartItemRepo.Update(ctx.Context(), existingItem); err != nil {
			return nil, err
		}
	} else {
		// Create new cart item
		cartItem := &entities.CartItem{
			CartID:      cart.ID,
			ProductID:   req.ProductID,
			Quantity:    req.Quantity,
			PriceAtTime: decimal.NewFromFloat(product.Price),
		}
		if err := s.cartItemRepo.Create(ctx.Context(), cartItem); err != nil {
			return nil, err
		}
	}

	// Return updated cart
	return s.GetCart(ctx, userID)
}

func (s *cartService) UpdateCartItem(ctx *fiber.Ctx, userID string, itemID string, req *dto.UpdateItemRequest) (*dto.CartResponse, error) {
	// Get cart
	cart, err := s.cartRepo.GetByUserID(ctx.Context(), userID)
	if err != nil {
		return nil, err
	}
	if cart == nil {
		return nil, errors.New("cart not found")
	}

	// Get cart item
	item, err := s.cartItemRepo.GetByID(ctx.Context(), itemID)
	if err != nil {
		return nil, err
	}
	if item == nil || item.CartID != cart.ID {
		return nil, errors.New("cart item not found")
	}

	// Validate product and check stock
	product, err := s.productService.GetProduct(ctx.Context(), item.ProductID)
	if err != nil {
		return nil, errors.New("product not found")
	}

	if !product.IsActive {
		return nil, errors.New("product is not available")
	}

	if product.Stock < req.Quantity {
		return nil, fmt.Errorf("insufficient stock. Only %d available", product.Stock)
	}

	// Update quantity and price
	item.Quantity = req.Quantity
	item.PriceAtTime = decimal.NewFromFloat(product.Price)
	if err := s.cartItemRepo.Update(ctx.Context(), item); err != nil {
		return nil, err
	}

	// Return updated cart
	return s.GetCart(ctx, userID)
}

func (s *cartService) RemoveItemFromCart(ctx *fiber.Ctx, userID string, itemID string) (*dto.CartResponse, error) {
	// Get cart
	cart, err := s.cartRepo.GetByUserID(ctx.Context(), userID)
	if err != nil {
		return nil, err
	}
	if cart == nil {
		return nil, errors.New("cart not found")
	}

	// Get cart item
	item, err := s.cartItemRepo.GetByID(ctx.Context(), itemID)
	if err != nil {
		return nil, err
	}
	if item == nil || item.CartID != cart.ID {
		return nil, errors.New("cart item not found")
	}

	// Delete item
	if err := s.cartItemRepo.Delete(ctx.Context(), itemID); err != nil {
		return nil, err
	}

	// Return updated cart
	return s.GetCart(ctx, userID)
}

func (s *cartService) ClearCart(ctx *fiber.Ctx, userID string) error {
	// Get cart
	cart, err := s.cartRepo.GetByUserID(ctx.Context(), userID)
	if err != nil {
		return err
	}
	if cart == nil {
		return nil // No cart to clear
	}

	// Delete all cart items
	return s.cartItemRepo.DeleteByCartID(ctx.Context(), cart.ID)
}

func (s *cartService) ValidateCart(ctx *fiber.Ctx, userID string) (*dto.CartValidationResponse, error) {
	// Get cart
	cart, err := s.cartRepo.GetByUserID(ctx.Context(), userID)
	if err != nil {
		return nil, err
	}

	if cart == nil {
		return &dto.CartValidationResponse{
			Valid:      true,
			TotalItems: 0,
			TotalPrice: decimal.NewFromFloat(0),
		}, nil
	}

	// Get cart items
	items, err := s.cartItemRepo.GetByCartID(ctx.Context(), cart.ID)
	if err != nil {
		return nil, err
	}

	response := &dto.CartValidationResponse{
		Valid:         true,
		TotalItems:    0,
		TotalPrice:    decimal.NewFromFloat(0),
		InvalidItems:  []dto.InvalidItemResponse{},
		UpdatedPrices: []dto.PriceUpdateResponse{},
	}

	for _, item := range items {
		// Fetch product details
		product, err := s.productService.GetProduct(ctx.Context(), item.ProductID)
		if err != nil {
			response.Valid = false
			response.InvalidItems = append(response.InvalidItems, dto.InvalidItemResponse{
				ProductID: item.ProductID,
				Reason:    "Product not found",
			})
			continue
		}

		// Check if product is active
		if !product.IsActive {
			response.Valid = false
			response.InvalidItems = append(response.InvalidItems, dto.InvalidItemResponse{
				ProductID: item.ProductID,
				Reason:    "Product is no longer available",
			})
			continue
		}

		// Check stock
		if product.Stock < item.Quantity {
			response.Valid = false
			reason := fmt.Sprintf("Insufficient stock. Only %d available", product.Stock)
			response.InvalidItems = append(response.InvalidItems, dto.InvalidItemResponse{
				ProductID: item.ProductID,
				Reason:    reason,
			})
			continue
		}

		// Check for price changes
		currentPrice := decimal.NewFromFloat(product.Price)
		if !item.PriceAtTime.Equal(currentPrice) {
			response.UpdatedPrices = append(response.UpdatedPrices, dto.PriceUpdateResponse{
				ProductID: item.ProductID,
				OldPrice:  item.PriceAtTime,
				NewPrice:  currentPrice,
			})
			// Update price in cart
			item.PriceAtTime = currentPrice
			s.cartItemRepo.Update(ctx.Context(), item)
		}

		response.TotalItems += item.Quantity
		response.TotalPrice = response.TotalPrice.Add(item.GetSubtotal())
	}

	return response, nil
}
