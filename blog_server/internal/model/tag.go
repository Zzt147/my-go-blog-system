package model

type Tag struct {
	Id    int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name  string `gorm:"column:name" json:"name"` // 你的数据库字段是 name 还是 tag_name？Java里是 name
	
	// 用于接收统计数量 (SQL: COUNT(...) as count)
	Count int    `gorm:"-" json:"count"` 
}

func (Tag) TableName() string {
	return "t_tag"
}