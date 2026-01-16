package utils

// PageParams 对应 Java 的 PageParams 类
type PageParams struct {
	Page    int    `json:"page"`    // 当前页码
	Rows    int    `json:"rows"`    // 每页条数
	Keyword string `json:"keyword"` // 搜索关键字 (对应 ArticleSearch)

	// [NEW] 新增 Total 字段
	// 前端分页组件通常需要后端把这个总数回传给它
	Total   int64  `json:"total"`
}

// GetOffset 计算数据库查询的偏移量 (Limit ?, ?)
func (p *PageParams) GetOffset() int {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Rows < 1 {
		p.Rows = 10
	}
	return (p.Page - 1) * p.Rows
}