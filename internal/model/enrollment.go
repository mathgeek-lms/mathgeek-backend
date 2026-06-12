package model

import "time"

type Enrollment struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	GroupID   int64     `json:"group_id" db:"group_id"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateEnrollmentRequest struct {
	GroupID int64 `json:"group_id"`
}

type EnrollmentWithDetails struct {
	ID          int64  `json:"id"`
	Status      string `json:"status"`
	GroupID     int64  `json:"group_id"`
	GroupTitle  string `json:"group_title"`
	CourseID    int64  `json:"course_id"`
	CourseTitle string `json:"course_title"`
}
