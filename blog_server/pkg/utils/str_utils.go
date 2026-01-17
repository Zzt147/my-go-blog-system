package utils

// SubString 截取字符串，防止数组越界
// str: 原字符串, length: 截取长度
func SubString(str string, length int) string {
	// 将字符串转为 rune 切片，以支持中文
	rs := []rune(str)
	rl := len(rs)
	
	if length < 0 {
		return ""
	}
	
	if rl > length {
		return string(rs[:length]) + "..."
	}
	
	return string(rs)
}