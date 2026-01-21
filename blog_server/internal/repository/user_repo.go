package repository

import (
	"my-blog/internal/model"

	"gorm.io/gorm"
)

// 接口定义
type UserRepository interface {
	FindAll() ([]model.User, error)
	FindById(id int) (*model.User, error)
	FindByUsername(username string) (*model.User, error)
	// [修复] 补上这两个缺失的方法定义
	FindByEmail(email string) (*model.User, error)
	Create(user *model.User) error

	// [NEW] 更新用户
	Update(user *model.User) error
}

// 结构体实现
type userRepository struct {
	db *gorm.DB
}

// 构造函数
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// --- 实现方法 ---

func (r *userRepository) FindAll() ([]model.User, error) {
	var users []model.User
	// SELECT * FROM t_user
	result := r.db.Find(&users)
	return users, result.Error
}

func (r *userRepository) FindById(id int) (*model.User, error) {
	var user model.User
	result := r.db.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// [NEW] 实现 Create
func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// [NEW] 实现 FindByUsername
func (r *userRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	result := r.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// [NEW] 实现 FindByEmail
func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// [NEW] 更新用户实现
func (r *userRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}
