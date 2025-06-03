package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id"`
	Email        *string            `json:"email" validate:"required,email"`
	Password     *string            `json:"password" validate:"required"`
	Token        *string            `json:"token"`
	RefreshToken *string            `json:"refresh_token"`
	UserID       string             `json:"user_id"`
}
