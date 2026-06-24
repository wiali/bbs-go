package services

import (
	"errors"

	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/locales"

	"bbs-go/internal/pkg/params"

	"github.com/mlogclub/simple/sqls"

	"bbs-go/internal/models"
	"bbs-go/internal/repositories"

	"gorm.io/gorm"
)

var CategoryService = newCategoryService()

func newCategoryService() *categoryService {
	return &categoryService{}
}

type categoryService struct {
}

func (s *categoryService) CanViewCategory(user *models.User, category *models.Category) bool {
	if category == nil || category.Status != constants.StatusOk {
		return false
	}
	switch category.Visibility {
	case constants.CategoryVisibilityPublic:
		return true
	case constants.CategoryVisibilityLogin:
		return user != nil
	case constants.CategoryVisibilityOwner:
		return user != nil && user.IsOwner()
	default:
		return false
	}
}

func (s *categoryService) CanViewTopic(user *models.User, topic *models.Topic) bool {
	if topic == nil || topic.Status == constants.StatusDeleted {
		return false
	}
	if topic.CategoryId <= 0 {
		return true
	}
	return s.CanViewCategory(user, s.Get(topic.CategoryId))
}

func (s *categoryService) FilterVisibleCategories(user *models.User, categories []models.Category) []models.Category {
	if len(categories) == 0 {
		return nil
	}
	visible := make([]models.Category, 0, len(categories))
	for _, category := range categories {
		if s.CanViewCategory(user, &category) {
			visible = append(visible, category)
		}
	}
	return visible
}

func (s *categoryService) FilterVisibleTopics(user *models.User, topics []models.Topic) []models.Topic {
	if len(topics) == 0 {
		return nil
	}
	visible := make([]models.Topic, 0, len(topics))
	for _, topic := range topics {
		if s.CanViewTopic(user, &topic) {
			visible = append(visible, topic)
		}
	}
	return visible
}

func (s *categoryService) HiddenCategoryIds(user *models.User) []int64 {
	if user != nil && user.IsOwner() {
		return nil
	}
	categories := s.GetCategories()
	hidden := make([]int64, 0)
	for _, category := range categories {
		if !s.CanViewCategory(user, &category) {
			hidden = append(hidden, category.Id)
		}
	}
	return hidden
}

func (s *categoryService) Get(id int64) *models.Category {
	return repositories.CategoryRepository.Get(sqls.DB(), id)
}

func (s *categoryService) Take(where ...interface{}) *models.Category {
	return repositories.CategoryRepository.Take(sqls.DB(), where...)
}

func (s *categoryService) Find(cnd *sqls.Cnd) []models.Category {
	return repositories.CategoryRepository.Find(sqls.DB(), cnd)
}

func (s *categoryService) FindOne(cnd *sqls.Cnd) *models.Category {
	return repositories.CategoryRepository.FindOne(sqls.DB(), cnd)
}

func (s *categoryService) FindPageByParams(params *params.QueryParams) (list []models.Category, paging *sqls.Paging) {
	return repositories.CategoryRepository.FindPageByParams(sqls.DB(), params)
}

func (s *categoryService) FindPageByCnd(cnd *sqls.Cnd) (list []models.Category, paging *sqls.Paging) {
	return repositories.CategoryRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *categoryService) Create(t *models.Category) error {
	return repositories.CategoryRepository.Create(sqls.DB(), t)
}

func (s *categoryService) Update(t *models.Category) error {
	return repositories.CategoryRepository.Update(sqls.DB(), t)
}

func (s *categoryService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.CategoryRepository.Updates(sqls.DB(), id, columns)
}

func (s *categoryService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.CategoryRepository.UpdateColumn(sqls.DB(), id, name, value)
}

// DeleteWithCheck 删除节点，若为一级且有子节点则返回错误
func (s *categoryService) DeleteWithCheck(id int64) error {
	category := s.Get(id)
	if category == nil {
		return nil
	}
	if category.ParentId == 0 {
		children := s.GetChildren(id)
		if len(children) > 0 {
			return errors.New(locales.Get("topic.category.has_children"))
		}
	}
	return repositories.CategoryRepository.Updates(sqls.DB(), id, map[string]interface{}{
		"status": constants.StatusDeleted,
	})
}

// GetTopLevelCategories 仅一级节点（parent_id=0），用于导航
func (s *categoryService) GetTopLevelCategories() []models.Category {
	return repositories.CategoryRepository.Find(sqls.DB(), sqls.NewCnd().
		Eq("status", constants.StatusOk).
		Eq("parent_id", 0).
		Asc("sort_no").Desc("id"))
}

// GetChildren 获取某一级下的二级节点
func (s *categoryService) GetChildren(parentId int64) []models.Category {
	return repositories.CategoryRepository.Find(sqls.DB(), sqls.NewCnd().
		Eq("status", constants.StatusOk).
		Eq("parent_id", parentId).
		Asc("sort_no").Desc("id"))
}

// GetCategoryIdsForList 用于帖子列表筛选：一级返回 [自身+子节点id]，二级返回 [自身]
func (s *categoryService) GetCategoryIdsForList(categoryId int64) []int64 {
	category := s.Get(categoryId)
	if category == nil {
		return nil
	}
	if category.ParentId == 0 {
		ids := []int64{categoryId}
		for _, c := range s.GetChildren(categoryId) {
			ids = append(ids, c.Id)
		}
		return ids
	}
	return []int64{categoryId}
}

func (s *categoryService) GetCategories() []models.Category {
	return repositories.CategoryRepository.Find(sqls.DB(), sqls.NewCnd().Eq("status", constants.StatusOk).Asc("sort_no").Desc("id"))
}

func (s *categoryService) GetCategoriesByType(categoryType constants.CategoryType) []models.Category {
	return repositories.CategoryRepository.Find(sqls.DB(), sqls.NewCnd().
		Eq("status", constants.StatusOk).
		Eq("type", categoryType).
		Asc("sort_no").Desc("id"))
}

func (s *categoryService) GetCategoriesByTopicType(topicType constants.TopicType) []models.Category {
	if topicType == constants.TopicTypeQA {
		return s.GetCategoriesByType(constants.CategoryTypeQA)
	}
	return s.GetCategoriesByType(constants.CategoryTypeNormal)
}

func (s *categoryService) GetNextSortNo() int {
	if max := s.FindOne(sqls.NewCnd().Eq("status", constants.StatusOk).Desc("sort_no")); max != nil {
		return max.SortNo + 1
	}
	return 0
}

func (s *categoryService) UpdateSort(ids []int64) error {
	return sqls.DB().Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := repositories.CategoryRepository.UpdateColumn(tx, id, "sort_no", i); err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateChildrenType 将父节点下所有子节点的 type 更新为指定值（父节点编辑类型时联动）
func (s *categoryService) UpdateChildrenType(parentId int64, categoryType constants.CategoryType) error {
	children := s.GetChildren(parentId)
	if len(children) == 0 {
		return nil
	}
	return sqls.DB().Transaction(func(tx *gorm.DB) error {
		for _, c := range children {
			if err := repositories.CategoryRepository.UpdateColumn(tx, c.Id, "type", categoryType); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *categoryService) UpdateChildrenVisibility(parentId int64, visibility constants.CategoryVisibility) error {
	children := s.GetChildren(parentId)
	if len(children) == 0 {
		return nil
	}
	return sqls.DB().Transaction(func(tx *gorm.DB) error {
		for _, c := range children {
			if err := repositories.CategoryRepository.UpdateColumn(tx, c.Id, "visibility", visibility); err != nil {
				return err
			}
		}
		return nil
	})
}
