package service

import "context"

type CourseChecker interface {
	IsCourseExistsByID(ctx context.Context, id int64) (bool, error)
}
