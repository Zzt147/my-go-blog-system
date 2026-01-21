package model

type Category struct {
	Id       int    `gorm:"primaryKey;autoIncrement" json:"id"`
	ParentId int    `gorm:"column:parent_id" json:"parentId"`
	Name     string `gorm:"column:name" json:"name"`
	Sort     int    `gorm:"column:sort" json:"sort"`

	// gorm:"-" 表示不对应数据库列，用于生成树形结构时嵌套子节点
	Children []*Category `gorm:"-" json:"children"`
}

func (Category) TableName() string {
	return "t_category"
}
