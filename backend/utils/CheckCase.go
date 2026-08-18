package utils

// IsAllLowercase 判断字符串是否全部由小写字母(a-z)组成
func IsAllLowercase(s string) bool {
	for _, ch := range s {
		if ch < 'a' || ch > 'z' {
			return false
		}
	}
	return len(s) > 0
}
