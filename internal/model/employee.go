package model

import "time"

type Employee struct {
	ID         string    `json:"id" bson:"id"`
	Name       string    `json:"name" bson:"name"`
	Email      string    `json:"email" bson:"email"`
	Department string    `json:"department" bson:"department"`
	Title      string    `json:"title" bson:"title"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at"`
}

type CreateEmployeeRequest struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Department string `json:"department"`
	Title      string `json:"title"`
}

type UpdateEmployeeRequest struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Department string `json:"department"`
	Title      string `json:"title"`
}
