package model

import "time"

type Lesson struct {
	ID          int64     `json:"id" db:"id"`
	CourseID    int64     `json:"course_id" db:"course_id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	Content     string    `json:"content" db:"content"`
	Position    int64     `json:"position" db:"position"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type LessonListItem struct {
	ID          int64     `json:"id"`
	CourseID    int64     `json:"course_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Position    int64     `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateLessonRequest struct {
	CourseID    int64  `json:"course_id" db:"course_id"`
	Title       string `json:"title" db:"title"`
	Description string `json:"description" db:"description"`
	Content     string `json:"content" db:"content"`
	Position    int64  `json:"position" db:"position"`
}
