package dto

import (
	"github.com/tasiuskenways/Scalable-Ecommerce/user-service/internal/domain/entities"
)

type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func NewUserResponse(user *entities.User) *UserResponse {
	return &UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}
}
