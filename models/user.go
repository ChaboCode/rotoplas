package models

import (
	"time"

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

type File struct {
	// ID        bson.ObjectID `bson:"_id"`
	Name      string    `json:"name" validate:"required"`
	Size      int64     `json:"size" validate:"required"`
	CreatedAt time.Time `bson:"created_at"`
	UploadIP  string    `json:"upload_ip" validate:"required"`
}
