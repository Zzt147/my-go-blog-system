package service

import (
	"errors"
	"fmt"
	"my-blog/internal/model"
	"my-blog/internal/repository"
	"my-blog/pkg/utils"
)

type CategoryService interface {
	GetTree() (*utils.Result, error)
	GetResources(id int) (*utils.Result, error)
	Add(category *model.Category) error
	Update(category *model.Category) error
	UpdateBatch(categories []model.Category) error
	Delete(id int, mode int) error
}

type categoryService struct {
	repo        repository.CategoryRepository
	articleRepo repository.ArticleRepository
}

func NewCategoryService(repo repository.CategoryRepository, articleRepo repository.ArticleRepository) CategoryService {
	return &categoryService{repo: repo, articleRepo: articleRepo}
}

// [NEW] 获取树形结构
func (s *categoryService) GetTree() (*utils.Result, error) {
	// 1. 查出所有
	all, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	// 2. 组装树 (Java stream groupingBy 逻辑的 Go 实现)
	tree := s.buildTree(all)

	res := utils.Ok()
	res.Put("data", tree)
	return res, nil
}

// [NEW] 获取右侧资源 (子文件夹 + 文章 + 当前路径)
func (s *categoryService) GetResources(id int) (*utils.Result, error) {
	// 1. 获取子分类 (Folders)
	folders, err := s.repo.FindByParentId(id)
	if err != nil {
		return nil, err
	}

	// 2. 获取该分类下的文章 (Articles)
	articles, err := s.articleRepo.FindByCategoryId(id)
	if err != nil {
		return nil, err
	}

	// 3. 构建当前路径字符串 (e.g., "技术 / 后端 / Go")
	currentPath, err := s.buildPath(id)
	if err != nil {
		currentPath = "根目录"
	}

	data := map[string]interface{}{
		"folders":     folders,
		"articles":    articles,
		"currentPath": currentPath,
	}

	res := utils.Ok()
	res.Put("data", data)
	return res, nil
}

// Add, Update, UpdateBatch 简单透传
func (s *categoryService) Add(category *model.Category) error {
	return s.repo.Create(category)
}
func (s *categoryService) Update(category *model.Category) error {
	return s.repo.Update(category)
}
func (s *categoryService) UpdateBatch(categories []model.Category) error {
	return s.repo.UpdateBatch(categories)
}

// [NEW] 删除分类 (复杂逻辑)
// mode=1: 仅删除分类，文章移至父级
// mode=2: 删除分类及文章
func (s *categoryService) Delete(id int, mode int) error {
	// 1. 查当前分类
	current, err := s.repo.FindById(id)
	if err != nil {
		return errors.New("分类不存在")
	}

	// 2. 查是否有子分类 (如果有子分类，不允许删除，或者需要递归删除，这里参考通常逻辑：有子分类不能直接删)
	children, _ := s.repo.FindByParentId(id)
	if len(children) > 0 {
		return errors.New("该分类下包含子分类，请先处理子分类")
	}

	// 3. 处理文章
	if mode == 1 {
		// 模式1: 移动文章到父级
		// 如果 parentId 是 0 (根目录), 则文章变成未分类(或者 category_id=0)
		err := s.articleRepo.UpdateCategoryId(id, current.ParentId)
		if err != nil {
			return err
		}
	} else if mode == 2 {
		// 模式2: 销毁所有文章
		err := s.articleRepo.DeleteByCategoryId(id)
		if err != nil {
			return err
		}
	} else {
		return errors.New("未知的删除模式")
	}

	// 4. 删除分类本身
	return s.repo.Delete(id)
}

// --- Helper Functions ---

// 递归构建树
func (s *categoryService) buildTree(all []*model.Category) []*model.Category {
	// Map 映射 ID -> Category
	nodeMap := make(map[int]*model.Category)
	for _, cat := range all {
		cat.Children = []*model.Category{} // 初始化空切片
		nodeMap[cat.Id] = cat
	}

	var roots []*model.Category
	for _, cat := range all {
		if cat.ParentId == 0 {
			roots = append(roots, cat)
		} else {
			if parent, ok := nodeMap[cat.ParentId]; ok {
				parent.Children = append(parent.Children, cat)
			} else {
				// 孤儿节点，也可以作为根或者忽略
				roots = append(roots, cat)
			}
		}
	}
	return roots
}

// 递归构建路径
func (s *categoryService) buildPath(id int) (string, error) {
	if id == 0 {
		return "", nil
	}
	cat, err := s.repo.FindById(id)
	if err != nil {
		return "", err
	}
	parentPath, _ := s.buildPath(cat.ParentId)
	if parentPath == "" {
		return cat.Name, nil
	}
	return fmt.Sprintf("%s / %s", parentPath, cat.Name), nil
}
