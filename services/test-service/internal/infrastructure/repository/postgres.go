package repository

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"BootDevClone/services/test-service/internal/domain"

	_ "github.com/lib/pq"
)

var ErrNotFound = errors.New("not found")

//go:embed migrations/*.sql
var migrationsFS embed.FS

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

const testSelectCols = `id, course_id, title, description, is_visible, status, open_date, language, starter_code, solution_code, test_code, expected_output, group_id`

func scanTest(row interface{ Scan(dest ...any) error }) (*domain.Test, error) {
	t := &domain.Test{}
	var openDate, lang, starter, solution, testC, expected, groupID sql.NullString
	err := row.Scan(&t.ID, &t.CourseID, &t.Title, &t.Description, &t.IsVisible, &t.Status,
		&openDate, &lang, &starter, &solution, &testC, &expected, &groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan test: %w", err)
	}
	if openDate.Valid { t.OpenDate = &openDate.String }
	if lang.Valid { t.Language = lang.String }
	if starter.Valid { t.StarterCode = starter.String }
	if solution.Valid { t.SolutionCode = solution.String }
	if testC.Valid { t.TestCode = testC.String }
	if expected.Valid { t.ExpectedOutput = expected.String }
	if groupID.Valid { t.GroupID = &groupID.String }
	return t, nil
}

func (r *PostgresRepository) GetTests() ([]*domain.Test, error) {
	query := fmt.Sprintf(`SELECT %s FROM tests`, testSelectCols)
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("get tests: %w", err)
	}
	defer rows.Close()

	var tests []*domain.Test
	for rows.Next() {
		t := &domain.Test{}
		var openDate, lang, starter, solution, testC, expected, groupID sql.NullString
		if err := rows.Scan(&t.ID, &t.CourseID, &t.Title, &t.Description, &t.IsVisible, &t.Status,
			&openDate, &lang, &starter, &solution, &testC, &expected, &groupID); err != nil {
			return nil, fmt.Errorf("scan test: %w", err)
		}
		if openDate.Valid { t.OpenDate = &openDate.String }
		if lang.Valid { t.Language = lang.String }
		if starter.Valid { t.StarterCode = starter.String }
		if solution.Valid { t.SolutionCode = solution.String }
		if testC.Valid { t.TestCode = testC.String }
		if expected.Valid { t.ExpectedOutput = expected.String }
		if groupID.Valid { t.GroupID = &groupID.String }
		tests = append(tests, t)
	}
	return tests, rows.Err()
}

func (r *PostgresRepository) GetTestByID(id string) (*domain.Test, error) {
	query := fmt.Sprintf(`SELECT %s FROM tests WHERE id=$1`, testSelectCols)
	row := r.db.QueryRow(query, id)
	return scanTest(row)
}

func (r *PostgresRepository) GetTestsByGroupID(groupID string) ([]*domain.Test, error) {
	query := fmt.Sprintf(`SELECT %s FROM tests WHERE group_id=$1`, testSelectCols)
	rows, err := r.db.Query(query, groupID)
	if err != nil {
		return nil, fmt.Errorf("get tests by group: %w", err)
	}
	defer rows.Close()

	var tests []*domain.Test
	for rows.Next() {
		t := &domain.Test{}
		var openDate, lang, starter, solution, testC, expected, gid sql.NullString
		if err := rows.Scan(&t.ID, &t.CourseID, &t.Title, &t.Description, &t.IsVisible, &t.Status,
			&openDate, &lang, &starter, &solution, &testC, &expected, &gid); err != nil {
			return nil, fmt.Errorf("scan test: %w", err)
		}
		if openDate.Valid { t.OpenDate = &openDate.String }
		if lang.Valid { t.Language = lang.String }
		if starter.Valid { t.StarterCode = starter.String }
		if solution.Valid { t.SolutionCode = solution.String }
		if testC.Valid { t.TestCode = testC.String }
		if expected.Valid { t.ExpectedOutput = expected.String }
		if gid.Valid { t.GroupID = &gid.String }
		tests = append(tests, t)
	}
	return tests, rows.Err()
}

func (r *PostgresRepository) CreateTest(t *domain.Test) error {
	query := `INSERT INTO tests (id, course_id, title, description, is_visible, status, open_date, language, starter_code, solution_code, test_code, expected_output, group_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	var openDate interface{}
	if t.OpenDate != nil {
		openDate = *t.OpenDate
	}
	var groupID interface{}
	if t.GroupID != nil {
		groupID = *t.GroupID
	}
	_, err := r.db.Exec(query, t.ID, t.CourseID, t.Title, t.Description, t.IsVisible, t.Status, openDate, t.Language, t.StarterCode, t.SolutionCode, t.TestCode, t.ExpectedOutput, groupID)
	if err != nil {
		return fmt.Errorf("create test: %w", err)
	}
	return nil
}

func (r *PostgresRepository) UpdateTest(t *domain.Test) error {
	query := `UPDATE tests SET title=$1, description=$2, is_visible=$3, status=$4, open_date=$5, language=$6, starter_code=$7, solution_code=$8, test_code=$9, expected_output=$10, group_id=$11 WHERE id=$12`
	var openDate interface{}
	if t.OpenDate != nil {
		openDate = *t.OpenDate
	}
	var groupID interface{}
	if t.GroupID != nil {
		groupID = *t.GroupID
	}
	_, err := r.db.Exec(query, t.Title, t.Description, t.IsVisible, t.Status, openDate, t.Language, t.StarterCode, t.SolutionCode, t.TestCode, t.ExpectedOutput, groupID, t.ID)
	if err != nil {
		return fmt.Errorf("update test: %w", err)
	}
	return nil
}

func (r *PostgresRepository) DeleteTest(id string) error {
	_, err := r.db.Exec(`DELETE FROM tests WHERE id=$1`, id)
	return err
}
