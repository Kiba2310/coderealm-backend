package repository

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"BootDevClone/services/group-service/internal/domain"

	"github.com/google/uuid"
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

func scanGroup(row interface{ Scan(dest ...any) error }) (*domain.Group, error) {
	g := &domain.Group{}
	err := row.Scan(&g.ID, &g.Name, &g.Description, &g.TeacherID, &g.InviteCode, &g.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan group: %w", err)
	}
	return g, nil
}

func (r *PostgresRepository) CreateGroup(g *domain.Group) error {
	_, err := r.db.Exec(`
		INSERT INTO groups (id, name, description, teacher_id, invite_code, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		g.ID, g.Name, g.Description, g.TeacherID, g.InviteCode, g.CreatedAt,
	)
	return err
}

func (r *PostgresRepository) GetGroupByID(id uuid.UUID) (*domain.Group, error) {
	row := r.db.QueryRow(`SELECT id, name, description, teacher_id, invite_code, created_at FROM groups WHERE id=$1`, id)
	return scanGroup(row)
}

func (r *PostgresRepository) GetGroupByInviteCode(code string) (*domain.Group, error) {
	row := r.db.QueryRow(`SELECT id, name, description, teacher_id, invite_code, created_at FROM groups WHERE invite_code=$1`, code)
	return scanGroup(row)
}

func (r *PostgresRepository) ListAllGroups() ([]*domain.Group, error) {
	rows, err := r.db.Query(`SELECT id, name, description, teacher_id, invite_code, created_at FROM groups ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*domain.Group
	for rows.Next() {
		g, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

func (r *PostgresRepository) ListTeacherGroups(teacherID uuid.UUID) ([]*domain.Group, error) {
	rows, err := r.db.Query(`SELECT id, name, description, teacher_id, invite_code, created_at FROM groups WHERE teacher_id=$1 ORDER BY created_at DESC`, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*domain.Group
	for rows.Next() {
		g, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

func (r *PostgresRepository) ListStudentGroups(studentID uuid.UUID) ([]*domain.Group, error) {
	rows, err := r.db.Query(`
		SELECT g.id, g.name, g.description, g.teacher_id, g.invite_code, g.created_at
		FROM groups g
		JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = $1
		ORDER BY gm.joined_at DESC`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*domain.Group
	for rows.Next() {
		g, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

func (r *PostgresRepository) DeleteGroup(id uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM groups WHERE id=$1`, id)
	return err
}

func (r *PostgresRepository) AddMember(groupID, userID uuid.UUID) error {
	_, err := r.db.Exec(`
		INSERT INTO group_members (group_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (group_id, user_id) DO NOTHING`,
		groupID, userID,
	)
	return err
}

func (r *PostgresRepository) RemoveMember(groupID, userID uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM group_members WHERE group_id=$1 AND user_id=$2`, groupID, userID)
	return err
}

func (r *PostgresRepository) ListGroupMembers(groupID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(`SELECT user_id FROM group_members WHERE group_id=$1`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var uid uuid.UUID
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, uid)
	}
	return userIDs, nil
}

func (r *PostgresRepository) IsMember(groupID, userID uuid.UUID) (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM group_members WHERE group_id=$1 AND user_id=$2`, groupID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
