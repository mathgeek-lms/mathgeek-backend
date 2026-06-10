-- +goose Up
CREATE TABLE IF NOT EXISTS lessons(
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL REFERENCES courses(id),
    title VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    content TEXT, 
    position BIGSERIAL NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT lessons_course_position_unique UNIQUE(course_id, position),
    CONSTRAINT lessons_position_positive CHECK (position > 0)
);

-- +goose Down
DROP TABLE IF EXISTS lessons;
