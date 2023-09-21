package user

import "time"

type User struct {
	ID         string    `json:"id" bson:"_id,omitempty"`
	Name       string    `json:"name" bson:"name"`
	Email      string    `json:"email" bson:"email"`
	IsVerified bool      `json:"is_verified" bson:"is_verified"`
	Role       Role      `json:"role" bson:"role"`
	Picture    string    `json:"picture" bson:"picture"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at"`
}

type Role string

const (
	Admin  Role = "admin"
	Reader Role = "user"
)
