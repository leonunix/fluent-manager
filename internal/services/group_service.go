package services

import (
	"github.com/fluent-manager/fluent-manager/internal/models"
	"gorm.io/gorm"
)

// GroupService manages user groups.
type GroupService interface {
	List(search string, page, pageSize int) ([]models.Group, int64, error)
	Get(id uint) (*models.Group, error)
	Create(name, description string, roleIDs []uint, scopes []ScopeInput) (*models.Group, error)
	Update(id uint, name, description string, roleIDs []uint, scopes []ScopeInput) (*models.Group, error)
	Delete(id uint) error
	SetGroupUsers(groupID uint, userIDs []uint) error
	AddUserToGroups(userID uint, groupIDs []uint) error
}

type groupService struct {
	db *gorm.DB
}

func NewGroupService(db *gorm.DB) GroupService {
	return &groupService{db: db}
}

func (s *groupService) List(search string, page, pageSize int) ([]models.Group, int64, error) {
	var groups []models.Group
	var total int64

	q := s.db.Model(&models.Group{})
	if search != "" {
		q = q.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	q.Count(&total)

	offset := (page - 1) * pageSize
	if err := q.Preload("Roles").Preload("Scopes").
		Offset(offset).Limit(pageSize).Order("id ASC").
		Find(&groups).Error; err != nil {
		return nil, 0, err
	}

	// Load user counts
	for i := range groups {
		groups[i].MemberCount = s.db.Model(&groups[i]).Association("Users").Count()
	}

	return groups, total, nil
}

func (s *groupService) Get(id uint) (*models.Group, error) {
	var group models.Group
	if err := s.db.Preload("Roles.Permissions").Preload("Scopes").Preload("Users").
		First(&group, id).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (s *groupService) Create(name, description string, roleIDs []uint, scopes []ScopeInput) (*models.Group, error) {
	group := models.Group{
		Name:        name,
		Description: description,
	}
	if err := s.db.Create(&group).Error; err != nil {
		return nil, err
	}

	if len(roleIDs) > 0 {
		var roles []models.Role
		s.db.Where("id IN ?", roleIDs).Find(&roles)
		s.db.Model(&group).Association("Roles").Replace(roles)
	}

	if len(scopes) > 0 {
		for _, sc := range scopes {
			gs := models.GroupScope{
				GroupID:   group.ID,
				ScopeType: sc.ScopeType,
				ScopeID:   sc.ScopeID,
				ScopeName: resolveScopeName(s.db, sc.ScopeType, sc.ScopeID),
			}
			s.db.Create(&gs)
		}
	}

	return s.Get(group.ID)
}

func (s *groupService) Update(id uint, name, description string, roleIDs []uint, scopes []ScopeInput) (*models.Group, error) {
	var group models.Group
	if err := s.db.First(&group, id).Error; err != nil {
		return nil, err
	}

	s.db.Model(&group).Updates(map[string]interface{}{
		"name":        name,
		"description": description,
	})

	// Update roles
	var roles []models.Role
	if len(roleIDs) > 0 {
		s.db.Where("id IN ?", roleIDs).Find(&roles)
	}
	s.db.Model(&group).Association("Roles").Replace(roles)

	// Replace scopes: delete old, create new
	s.db.Where("group_id = ?", id).Delete(&models.GroupScope{})
	for _, sc := range scopes {
		gs := models.GroupScope{
			GroupID:   id,
			ScopeType: sc.ScopeType,
			ScopeID:   sc.ScopeID,
			ScopeName: resolveScopeName(s.db, sc.ScopeType, sc.ScopeID),
		}
		s.db.Create(&gs)
	}

	return s.Get(id)
}

func (s *groupService) Delete(id uint) error {
	var group models.Group
	if err := s.db.First(&group, id).Error; err != nil {
		return err
	}

	s.db.Model(&group).Association("Roles").Clear()
	s.db.Model(&group).Association("Users").Clear()
	s.db.Where("group_id = ?", id).Delete(&models.GroupScope{})
	s.db.Where("group_id = ?", id).Delete(&models.ExternalGroupMapping{})

	return s.db.Delete(&group).Error
}

func (s *groupService) SetGroupUsers(groupID uint, userIDs []uint) error {
	var group models.Group
	if err := s.db.First(&group, groupID).Error; err != nil {
		return err
	}

	var users []models.User
	if len(userIDs) > 0 {
		s.db.Where("id IN ?", userIDs).Find(&users)
	}
	return s.db.Model(&group).Association("Users").Replace(users)
}

func resolveScopeName(db *gorm.DB, scopeType string, scopeID uint) string {
	switch scopeType {
	case "datacenter":
		var dc models.DataCenter
		if db.First(&dc, scopeID).Error == nil {
			if dc.Alias != "" {
				return dc.Alias
			}
			return dc.Name
		}
	case "region":
		var r models.Region
		if db.First(&r, scopeID).Error == nil {
			if r.Alias != "" {
				return r.Alias
			}
			return r.Name
		}
	case "cluster":
		var cl models.Cluster
		if db.First(&cl, scopeID).Error == nil {
			if cl.Alias != "" {
				return cl.Alias
			}
			return cl.Name
		}
	}
	return ""
}

func (s *groupService) AddUserToGroups(userID uint, groupIDs []uint) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}

	var groups []models.Group
	if len(groupIDs) > 0 {
		s.db.Where("id IN ?", groupIDs).Find(&groups)
	}
	return s.db.Model(&user).Association("Groups").Replace(groups)
}
