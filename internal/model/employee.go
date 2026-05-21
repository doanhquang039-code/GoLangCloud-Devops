package model

import "time"

type Employee struct {
    ID         string    `json:"id"`
    Name       string    `json:"name"`
    Email      string    `json:"email"`
    Department string    `json:"department"`
    Title      string    `json:"title"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}

type CreateEmployeeRequest struct {
    Name       string `json:"name"`
    Email      string `json:"email"`
    Department string `json:"department"`
    Title      string `json:"title"`
}
