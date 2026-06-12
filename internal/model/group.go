package model

import "time"

type Group struct {
	ID        int64      `json:"id" db:"id"`
	CourseID  int64      `json:"course_id" db:"course_id"`
	Title     string     `json:"title" db:"title"`
	StartDate *time.Time `json:"start_date" db:"start_date"`
	EndDate   *time.Time `json:"end_date" db:"end_date"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}
