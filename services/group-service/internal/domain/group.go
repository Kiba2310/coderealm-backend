package domain

import (
	"time"

	"github.com/google/uuid"
)

type Group struct {
	ID          uuid.UUID
	Name        string
	Description string
	TeacherID   uuid.UUID
	InviteCode  string
	CreatedAt   time.Time
}

type GroupMember struct {
	GroupID  uuid.UUID
	UserID   uuid.UUID
	JoinedAt time.Time
}

type GroupRepository interface {
	CreateGroup(g *Group) error
	GetGroupByID(id uuid.UUID) (*Group, error)
	GetGroupByInviteCode(code string) (*Group, error)
	ListAllGroups() ([]*Group, error)
	ListTeacherGroups(teacherID uuid.UUID) ([]*Group, error)
	ListStudentGroups(studentID uuid.UUID) ([]*Group, error)
	DeleteGroup(id uuid.UUID) error
	AddMember(groupID, userID uuid.UUID) error
	RemoveMember(groupID, userID uuid.UUID) error
	ListGroupMembers(groupID uuid.UUID) ([]uuid.UUID, error)
	IsMember(groupID, userID uuid.UUID) (bool, error)
}
