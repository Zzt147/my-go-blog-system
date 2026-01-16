package utils

// Result 对应 Java 的 Result 类
type Result struct {
	Success bool                   `json:"success"` // 对应 private boolean success
	Msg     string                 `json:"msg"`     // 对应 private String msg
	Map     map[string]interface{} `json:"map"`     // 对应 private Map<String, Object> map
}

// Ok 成功时的快捷方法
func Ok() *Result {
	return &Result{
		Success: true,
		Msg:     "",
		Map:     make(map[string]interface{}),
	}
}

// Error 失败时的快捷方法
func Error(msg string) *Result {
	return &Result{
		Success: false,
		Msg:     msg,
		Map:     make(map[string]interface{}),
	}
}

// Put 模仿 Java 的 result.getMap().put("key", value)
// 这是一个“方法”，绑定在 Result 结构体上
func (r *Result) Put(key string, value interface{}) *Result {
	if r.Map == nil {
		r.Map = make(map[string]interface{})
	}
	r.Map[key] = value
	return r
}