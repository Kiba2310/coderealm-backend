package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"BootDevClone/services/group-service/internal/domain"
	"BootDevClone/services/group-service/pkg/types"

	"github.com/google/uuid"
)

type GroupService struct {
	repo domain.GroupRepository
}

func NewGroupService(repo domain.GroupRepository) *GroupService {
	return &GroupService{repo: repo}
}

func generateInviteCode() string {
	bytes := make([]byte, 4) // 8 characters
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func domainToModel(g *domain.Group) types.GroupModel {
	return types.GroupModel{
		ID:          g.ID.String(),
		Name:        g.Name,
		Description: g.Description,
		TeacherID:   g.TeacherID.String(),
		InviteCode:  g.InviteCode,
		CreatedAt:   g.CreatedAt,
	}
}

func (s *GroupService) CreateGroup(teacherIDStr string, req types.CreateGroupRequest) (*types.GroupModel, error) {
	teacherID, err := uuid.Parse(teacherIDStr)
	if err != nil {
		return nil, errors.New("invalid teacher ID")
	}

	group := &domain.Group{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		TeacherID:   teacherID,
		InviteCode:  generateInviteCode(),
		CreatedAt:   time.Now(),
	}

	if err := s.repo.CreateGroup(group); err != nil {
		return nil, err
	}

	model := domainToModel(group)
	return &model, nil
}

func (s *GroupService) JoinGroup(studentIDStr string, req types.JoinGroupRequest) (*types.GroupModel, error) {
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		return nil, errors.New("invalid student ID")
	}

	group, err := s.repo.GetGroupByInviteCode(req.InviteCode)
	if err != nil {
		return nil, errors.New("invalid invite code or group not found")
	}

	if err := s.repo.AddMember(group.ID, studentID); err != nil {
		return nil, err
	}

	model := domainToModel(group)
	return &model, nil
}

func (s *GroupService) ListTeacherGroups(teacherIDStr string) ([]types.GroupModel, error) {
	teacherID, err := uuid.Parse(teacherIDStr)
	if err != nil {
		return nil, errors.New("invalid teacher ID")
	}

	groups, err := s.repo.ListTeacherGroups(teacherID)
	if err != nil {
		return nil, err
	}

	models := make([]types.GroupModel, 0, len(groups))
	for _, g := range groups {
		models = append(models, domainToModel(g))
	}
	return models, nil
}

func (s *GroupService) ListStudentGroups(studentIDStr string) ([]types.GroupModel, error) {
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		return nil, errors.New("invalid student ID")
	}

	groups, err := s.repo.ListStudentGroups(studentID)
	if err != nil {
		return nil, err
	}

	models := make([]types.GroupModel, 0, len(groups))
	for _, g := range groups {
		models = append(models, domainToModel(g))
	}
	return models, nil
}

func (s *GroupService) AdminListAllGroups() ([]types.GroupModel, error) {
	groups, err := s.repo.ListAllGroups()
	if err != nil {
		return nil, err
	}

	models := make([]types.GroupModel, 0, len(groups))
	for _, g := range groups {
		models = append(models, domainToModel(g))
	}
	return models, nil
}

func (s *GroupService) GetGroupByID(groupIDStr string) (*types.GroupModel, error) {
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		return nil, errors.New("invalid group ID")
	}

	group, err := s.repo.GetGroupByID(groupID)
	if err != nil {
		return nil, err
	}

	model := domainToModel(group)
	return &model, nil
}

func (s *GroupService) GetGroupMembers(groupIDStr string) ([]string, error) {
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		return nil, errors.New("invalid group ID")
	}
	memberIDs, err := s.repo.ListGroupMembers(groupID)
	if err != nil {
		return nil, err
	}

	var ids []string
	for _, id := range memberIDs {
		ids = append(ids, id.String())
	}
	return ids, nil
}

func (s *GroupService) RemoveMember(groupIDStr, studentIDStr, requestorIDStr string, requestorRoles string) error {
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		return errors.New("invalid group ID")
	}
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		return errors.New("invalid student ID")
	}
	requestorID, err := uuid.Parse(requestorIDStr)
	if err != nil {
		return errors.New("invalid requestor ID")
	}

	// Only teacher or admin can remove a member, or the student themselves (leaving)
	group, err := s.repo.GetGroupByID(groupID)
	if err != nil {
		return err
	}

	// This logic handles a student leaving or a teacher removing a student
	if requestorID != group.TeacherID && requestorID != studentID {
		return errors.New("unauthorized to remove member")
	}

	return s.repo.RemoveMember(groupID, studentID)
}

func (s *GroupService) DeleteGroup(groupIDStr, teacherIDStr string) error {
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		return errors.New("invalid group ID")
	}
	teacherID, err := uuid.Parse(teacherIDStr)
	if err != nil {
		return errors.New("invalid teacher ID")
	}

	group, err := s.repo.GetGroupByID(groupID)
	if err != nil {
		return err
	}

	if group.TeacherID != teacherID {
		return errors.New("unauthorized to delete this group")
	}

	return s.repo.DeleteGroup(groupID)
}
