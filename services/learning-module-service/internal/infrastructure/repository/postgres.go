package repository

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"BootDevClone/services/learning-module-service/internal/domain"

	_ "github.com/lib/pq"
)

var ErrNotFound = errors.New("not found")

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations executes all SQL files in the migrations directory in order.
func RunMigrations(db *sql.DB) error {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("run migration %s: %w", entry.Name(), err)
		}
	}
	return nil
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) ListCourses(lang string) ([]*domain.Course, error) {
	query := `
		SELECT 
			c.id, c.title, c.description, COALESCE(c.language, ''), COALESCE(c.category, ''), COALESCE(c.difficulty, ''),
			COUNT(DISTINCT ch.id) as chapters_count,
			COUNT(DISTINCT l.id) as tasks_count,
			COALESCE(string_agg(l.id::text, ','), '') as lesson_ids
		FROM courses c
		LEFT JOIN chapters ch ON ch.course_id = c.id
		LEFT JOIN lessons l ON l.chapter_id = ch.id
		GROUP BY c.id
	`
	if lang == "hu" {
		query = `
			SELECT 
				c.id, COALESCE(c.title_hu, c.title), COALESCE(c.description_hu, c.description), COALESCE(c.language, ''), COALESCE(c.category, ''), COALESCE(c.difficulty, ''),
				COUNT(DISTINCT ch.id) as chapters_count,
				COUNT(DISTINCT l.id) as tasks_count,
				COALESCE(string_agg(l.id::text, ','), '') as lesson_ids
			FROM courses c
			LEFT JOIN chapters ch ON ch.course_id = c.id
			LEFT JOIN lessons l ON l.chapter_id = ch.id
			GROUP BY c.id
		`
	}
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("list courses: %w", err)
	}
	defer rows.Close()

	var courses []*domain.Course
	for rows.Next() {
		c := &domain.Course{}
		var lessonIDsStr string
		if err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.Language, &c.Category, &c.Difficulty, &c.ChaptersCount, &c.TasksCount, &lessonIDsStr); err != nil {
			return nil, fmt.Errorf("scan course: %w", err)
		}

		if lessonIDsStr != "" {
			c.LessonIDs = strings.Split(lessonIDsStr, ",")
		} else {
			c.LessonIDs = []string{}
		}
		courses = append(courses, c)
	}
	return courses, rows.Err()
}

func (r *PostgresRepository) GetCourse(id string, lang string) (*domain.Course, error) {
	course := &domain.Course{}

	courseQuery := `SELECT id, title, description, COALESCE(language, ''), COALESCE(category, ''), COALESCE(difficulty, '') FROM courses WHERE id = $1`
	if lang == "hu" {
		courseQuery = `SELECT id, COALESCE(title_hu, title), COALESCE(description_hu, description), COALESCE(language, ''), COALESCE(category, ''), COALESCE(difficulty, '') FROM courses WHERE id = $1`
	}
	err := r.db.QueryRow(courseQuery, id).
		Scan(&course.ID, &course.Title, &course.Description, &course.Language, &course.Category, &course.Difficulty)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get course: %w", err)
	}

	chQuery := `SELECT id, course_id, order_num, title FROM chapters WHERE course_id = $1 ORDER BY order_num`
	if lang == "hu" {
		chQuery = `SELECT id, course_id, order_num, COALESCE(title_hu, title) FROM chapters WHERE course_id = $1 ORDER BY order_num`
	}
	chRows, err := r.db.Query(chQuery, id)
	if err != nil {
		return nil, fmt.Errorf("get chapters: %w", err)
	}
	defer chRows.Close()

	for chRows.Next() {
		ch := domain.Chapter{}
		if err := chRows.Scan(&ch.ID, &ch.CourseID, &ch.Order, &ch.Title); err != nil {
			return nil, fmt.Errorf("scan chapter: %w", err)
		}

		lQuery := `
			SELECT id, chapter_id, order_num, title, description, type,
			       COALESCE(language,''), COALESCE(starter_code,''), COALESCE(solution_code,''), COALESCE(test_code,''), COALESCE(expected_output,''), COALESCE(document_url,'')
			FROM lessons WHERE chapter_id = $1 ORDER BY order_num`
		if lang == "hu" {
			lQuery = `
				SELECT id, chapter_id, order_num, COALESCE(title_hu, title), COALESCE(description_hu, description), type,
				       COALESCE(language,''), COALESCE(starter_code,''), COALESCE(solution_code,''), COALESCE(test_code,''), COALESCE(expected_output,''), COALESCE(document_url,'')
				FROM lessons WHERE chapter_id = $1 ORDER BY order_num`
		}
		lRows, err := r.db.Query(lQuery, ch.ID)
		if err != nil {
			return nil, fmt.Errorf("get lessons: %w", err)
		}

		for lRows.Next() {
			l := domain.Lesson{}
			var lessonType string
			if err := lRows.Scan(&l.ID, &l.ChapterID, &l.Order, &l.Title, &l.Description,
				&lessonType, &l.Language, &l.StarterCode, &l.SolutionCode, &l.TestCode, &l.ExpectedOutput, &l.DocumentURL); err != nil {
				lRows.Close()
				return nil, fmt.Errorf("scan lesson: %w", err)
			}
			l.Type = domain.LessonType(lessonType)
			ch.Lessons = append(ch.Lessons, l)
		}
		lRows.Close()
		if err := lRows.Err(); err != nil {
			return nil, err
		}

		course.Chapters = append(course.Chapters, ch)
	}
	return course, chRows.Err()
}

func (r *PostgresRepository) GetLesson(courseID, chapterID, lessonID string, lang string) (*domain.Lesson, error) {
	l := &domain.Lesson{}
	var lessonType string

	lQuery := `
		SELECT l.id, l.chapter_id, l.order_num, l.title, l.description, l.type,
		       COALESCE(l.language,''), COALESCE(l.starter_code,''), COALESCE(l.solution_code,''), COALESCE(l.test_code,''), COALESCE(l.expected_output,''), COALESCE(l.document_url,'')
		FROM lessons l
		JOIN chapters ch ON l.chapter_id = ch.id
		WHERE l.id = $1 AND ch.id = $2 AND ch.course_id = $3`
	if lang == "hu" {
		lQuery = `
			SELECT l.id, l.chapter_id, l.order_num, COALESCE(l.title_hu, l.title), COALESCE(l.description_hu, l.description), l.type,
			       COALESCE(l.language,''), COALESCE(l.starter_code,''), COALESCE(l.solution_code,''), COALESCE(l.test_code,''), COALESCE(l.expected_output,''), COALESCE(l.document_url,'')
			FROM lessons l
			JOIN chapters ch ON l.chapter_id = ch.id
			WHERE l.id = $1 AND ch.id = $2 AND ch.course_id = $3`
	}
	err := r.db.QueryRow(lQuery, lessonID, chapterID, courseID).Scan(&l.ID, &l.ChapterID, &l.Order, &l.Title, &l.Description,
		&lessonType, &l.Language, &l.StarterCode, &l.SolutionCode, &l.TestCode, &l.ExpectedOutput, &l.DocumentURL)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get lesson: %w", err)
	}
	l.Type = domain.LessonType(lessonType)

	if l.Type == domain.LessonTypeMultipleChoice {
		optQuery := `SELECT id, text, is_correct FROM answer_options WHERE lesson_id = $1`
		if lang == "hu" {
			optQuery = `SELECT id, COALESCE(text_hu, text), is_correct FROM answer_options WHERE lesson_id = $1`
		}
		optRows, err := r.db.Query(optQuery, l.ID)
		if err != nil {
			return nil, fmt.Errorf("get answer options: %w", err)
		}
		defer optRows.Close()

		for optRows.Next() {
			opt := domain.AnswerOption{}
			if err := optRows.Scan(&opt.ID, &opt.Text, &opt.IsCorrect); err != nil {
				return nil, fmt.Errorf("scan answer option: %w", err)
			}
			l.Options = append(l.Options, opt)
		}
		if err := optRows.Err(); err != nil {
			return nil, err
		}
	}
	return l, nil
}

func (r *PostgresRepository) CreateCourse(course *domain.Course) error {
	query := `INSERT INTO courses (id, title, description, language, category, difficulty) VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''))`
	_, err := r.db.Exec(query, course.ID, course.Title, course.Description, course.Language, course.Category, course.Difficulty)
	if err != nil {
		return fmt.Errorf("create course: %w", err)
	}
	return nil
}

func (r *PostgresRepository) CreateChapter(chapter *domain.Chapter) error {
	query := `INSERT INTO chapters (id, course_id, order_num, title) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(query, chapter.ID, chapter.CourseID, chapter.Order, chapter.Title)
	if err != nil {
		return fmt.Errorf("create chapter: %w", err)
	}
	return nil
}

func (r *PostgresRepository) CreateLesson(lesson *domain.Lesson) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("create lesson begin tx: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO lessons (id, chapter_id, order_num, title, description, type, language, starter_code, solution_code, test_code, expected_output, document_url)
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''), NULLIF($11, ''), NULLIF($12, ''))`

	_, err = tx.Exec(query, lesson.ID, lesson.ChapterID, lesson.Order, lesson.Title, lesson.Description,
		lesson.Type, lesson.Language, lesson.StarterCode, lesson.SolutionCode, lesson.TestCode, lesson.ExpectedOutput, lesson.DocumentURL)
	if err != nil {
		return fmt.Errorf("create lesson: %w", err)
	}

	if lesson.Type == domain.LessonTypeMultipleChoice && len(lesson.Options) > 0 {
		optQuery := `INSERT INTO answer_options (id, lesson_id, text, is_correct) VALUES ($1, $2, $3, $4)`
		for _, opt := range lesson.Options {
			_, err = tx.Exec(optQuery, opt.ID, lesson.ID, opt.Text, opt.IsCorrect)
			if err != nil {
				return fmt.Errorf("create answer option: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("create lesson commit tx: %w", err)
	}

	return nil
}

// ==================== Update/Delete Methods ====================

func (r *PostgresRepository) UpdateCourse(course *domain.Course) error {
	_, err := r.db.Exec(`UPDATE courses SET title=$1, description=$2, language=NULLIF($3, ''), category=NULLIF($4, ''), difficulty=NULLIF($5, '') WHERE id=$6`,
		course.Title, course.Description, course.Language, course.Category, course.Difficulty, course.ID)
	if err != nil {
		return fmt.Errorf("update course: %w", err)
	}
	return nil
}

func (r *PostgresRepository) DeleteCourse(id string) error {
	_, err := r.db.Exec(`DELETE FROM courses WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete course: %w", err)
	}
	return nil
}

func (r *PostgresRepository) UpdateChapter(chapter *domain.Chapter) error {
	_, err := r.db.Exec(`UPDATE chapters SET title=$1, order_num=$2 WHERE id=$3 AND course_id=$4`,
		chapter.Title, chapter.Order, chapter.ID, chapter.CourseID)
	if err != nil {
		return fmt.Errorf("update chapter: %w", err)
	}
	return nil
}

func (r *PostgresRepository) DeleteChapter(courseID, chapterID string) error {
	_, err := r.db.Exec(`DELETE FROM chapters WHERE id=$1 AND course_id=$2`, chapterID, courseID)
	if err != nil {
		return fmt.Errorf("delete chapter: %w", err)
	}
	return nil
}

func (r *PostgresRepository) UpdateLesson(lesson *domain.Lesson) error {
	_, err := r.db.Exec(`UPDATE lessons SET title=$1, description=$2, type=$3, order_num=$4, 
		language=NULLIF($5, ''), starter_code=NULLIF($6, ''), solution_code=NULLIF($7, ''), 
		test_code=NULLIF($8, ''), expected_output=NULLIF($9, ''), document_url=NULLIF($10, '')
		WHERE id=$11 AND chapter_id=$12`,
		lesson.Title, lesson.Description, lesson.Type, lesson.Order,
		lesson.Language, lesson.StarterCode, lesson.SolutionCode,
		lesson.TestCode, lesson.ExpectedOutput, lesson.DocumentURL,
		lesson.ID, lesson.ChapterID)
	if err != nil {
		return fmt.Errorf("update lesson: %w", err)
	}
	return nil
}

func (r *PostgresRepository) DeleteLesson(courseID, chapterID, lessonID string) error {
	_, err := r.db.Exec(`DELETE FROM lessons WHERE id=$1 AND chapter_id=$2`, lessonID, chapterID)
	if err != nil {
		return fmt.Errorf("delete lesson: %w", err)
	}
	return nil
}

func (r *PostgresRepository) IncrementLessonHelp(userID, lessonID string) (int, error) {
	var count int
	err := r.db.QueryRow(`
		INSERT INTO lesson_helps (user_id, lesson_id, help_count)
		VALUES ($1, $2, 1)
		ON CONFLICT (user_id, lesson_id)
		DO UPDATE SET help_count = lesson_helps.help_count + 1, updated_at = CURRENT_TIMESTAMP
		RETURNING help_count
	`, userID, lessonID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("increment lesson help: %w", err)
	}
	return count, nil
}

func (r *PostgresRepository) GetLessonHelpCount(userID, lessonID string) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT help_count FROM lesson_helps
		WHERE user_id = $1 AND lesson_id = $2
	`, userID, lessonID).Scan(&count)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("get lesson help count: %w", err)
	}
	return count, nil
}
