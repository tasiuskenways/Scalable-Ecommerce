package handlers

import (
	"errors"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/application/dto"
	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/domain/services"
	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/utils"
)

type StoreHandler struct {
	storeService services.StoreService
	validator    *validator.Validate
}

func NewStoreHandler(storeService services.StoreService) *StoreHandler {
	return &StoreHandler{
		storeService: storeService,
		validator:    validator.New(),
	}
}

func (h *StoreHandler) CreateStore(c *fiber.Ctx) error {
	userID := c.Get("X-User-Id")
	if userID == "" {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "User ID is required")
	}

	var req dto.CreateStoreRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Struct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	store, err := h.storeService.CreateStore(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Store created successfully", store)
}

func (h *StoreHandler) GetStore(c *fiber.Ctx) error {
	userID := c.Get("X-User-Id")
	if userID == "" {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "User ID is required")
	}

	storeID := c.Params("id")
	if storeID == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Store ID is required")
	}

	store, err := h.storeService.GetStore(storeID, userID)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Store not found")
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Store retrieved successfully", store)
}

func (h *StoreHandler) GetStoreBySlug(c *fiber.Ctx) error {
	userID := c.Get("X-User-Id")
	if userID == "" {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "User ID is required")
	}

	slug := c.Params("slug")
	if slug == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Store slug is required")
	}

	store, err := h.storeService.GetStoreBySlug(slug, userID)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Store not found")
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Store retrieved successfully", store)
}

func (h *StoreHandler) GetUserStores(c *fiber.Ctx) error {
	userID := c.Get("X-User-Id")
	if userID == "" {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "User ID is required")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "10"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	stores, err := h.storeService.GetUserStores(userID, page, perPage)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Stores retrieved successfully", stores)
}

func (h *StoreHandler) UpdateStore(c *fiber.Ctx) error {
	userID := c.Get("X-User-Id")
	if userID == "" {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "User ID is required")
	}

	storeID := c.Params("id")
	if storeID == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Store ID is required")
	}

	var req dto.UpdateStoreRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Struct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	store, err := h.storeService.UpdateStore(storeID, userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Store updated successfully", store)
}

func (h *StoreHandler) DeleteStore(c *fiber.Ctx) error {
	userID := c.Get("X-User-Id")
	if userID == "" {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "User ID is required")
	}

	storeID := c.Params("id")
	if storeID == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Store ID is required")
	}

	err := h.storeService.DeleteStore(storeID, userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Store deleted successfully", nil)
}

// Member management endpoints

func (h *StoreHandler) InviteMember(c *fiber.Ctx) error {
	userID := c.Get("X-User-Id")
	if userID == "" {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "User ID is required")
	}

	storeID := c.Params("id")
	if storeID == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Store ID is required")
	}

	var req dto.InviteMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Struct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	invitation, err := h.storeService.InviteMember(storeID, userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Member invited successfully", invitation)
}

func (h *StoreHandler) AcceptInvitation(c *fiber.Ctx) error {
	userID := c.Get("X-User-Id")
	if userID == "" {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "User ID is required")
	}

	var req dto.AcceptInvitationRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Struct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	err := h.storeService.AcceptInvitation(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Invitation accepted successfully", nil)
}

func (h *StoreHandler) GetStoreMembers(c *fiber.Ctx) error {
	userID := c.Get("X-User-Id")
	if userID == "" {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "User ID is required")
	}

	storeID := c.Params("id")
	if storeID == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Store ID is required")
	}

	members, err := h.storeService.GetStoreMembers(storeID, userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Store members retrieved successfully", members)
}

func (h *StoreHandler) GetStoreInvitations(c *fiber.Ctx) error {
	userID := c.Get("X-User-Id")
	if userID == "" {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "User ID is required")
	}

	storeID := c.Params("id")
	if storeID == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Store ID is required")
	}

	invitations, err := h.storeService.GetStoreInvitations(storeID, userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Store invitations retrieved successfully", invitations)
}

func (h *StoreHandler) UpdateMemberRole(c *fiber.Ctx) error {
	userID := c.Get("X-User-Id")
	if userID == "" {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "User ID is required")
	}

	storeID := c.Params("id")
	if storeID == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Store ID is required")
	}

	memberID := c.Params("memberId")
	if memberID == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Member ID is required")
	}

	var req dto.UpdateMemberRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Struct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	err := h.storeService.UpdateMemberRole(storeID, memberID, userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Member role updated successfully", nil)
}

func (h *StoreHandler) RemoveMember(c *fiber.Ctx) error {
	userID := c.Get("X-User-Id")
	if userID == "" {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "User ID is required")
	}

	storeID := c.Params("id")
	if storeID == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Store ID is required")
	}

	memberID := c.Params("memberId")
	if memberID == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Member ID is required")
	}

	err := h.storeService.RemoveMember(storeID, memberID, userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Member removed successfully", nil)
}

func (h *StoreHandler) GetUserInvitations(c *fiber.Ctx) error {
	userEmail := c.Get("X-User-Email")
	if userEmail == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Email parameter is required")
	}

	invitations, err := h.storeService.GetUserInvitations(userEmail)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "User invitations retrieved successfully", invitations)
}
