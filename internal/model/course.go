package model

import "time"

type Course struct {
	ID             int64     `json:"id" db:"id"`
	Title          string    `json:"title" db:"title"`
	Description    string    `json:"description" db:"description"`
	DurationMonths int       `json:"duration_months" db:"duration_months"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type CreateCourseRequest struct {
	Title          string  `json:"title"`
	Description    *string `json:"description"`
	DurationMonths int     `json:"duration_months"`
}
