package model

import "time"

// User 对应数据库 t_user 表
type User struct {
	Id       int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Username string    `gorm:"column:username" json:"username"`
	Password string    `gorm:"column:password" json:"-"` // json:"-" 表示不返回给前端
	Email    string    `gorm:"column:email" json:"email"`
	Avatar   string    `gorm:"column:avatar" json:"avatar"`
	Created  time.Time `gorm:"column:created" json:"created"`
	Valid    int       `gorm:"column:valid" json:"valid"` // tinyint(1) 通常映射为 int 或 bool
	// [NEW] 新增权限字段 (为了骗过前端的路由生成器)
	// 对应 Java List<GrantedAuthority>，序列化后通常是对象数组
	Authorities []map[string]string `gorm:"-" json:"authorities"`
}

// TableName 强制指定表名为 t_user (非常重要，否则 GORM 会去找 users 表)
func (User) TableName() string {
	return "t_user"
}