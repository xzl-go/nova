package utils

import (
	"strings"
	"unicode"
)

// IsEmpty 检查字符串是否为空
func IsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// IsNotEmpty 检查字符串是否非空
func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}

// IsBlank 检查字符串是否为空白
func IsBlank(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// IsNotBlank 检查字符串是否非空白
func IsNotBlank(s string) bool {
	return !IsBlank(s)
}

// IsNumeric 检查字符串是否为数字
func IsNumeric(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// IsAlpha 检查字符串是否为字母
func IsAlpha(s string) bool {
	for _, c := range s {
		if !unicode.IsLetter(c) {
			return false
		}
	}
	return true
}

// IsAlphanumeric 检查字符串是否为字母和数字
func IsAlphanumeric(s string) bool {
	for _, c := range s {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// IsEmail 检查字符串是否为邮箱
func IsEmail(s string) bool {
	return strings.Contains(s, "@") && strings.Contains(s, ".")
}

// IsURL 检查字符串是否为 URL
func IsURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// IsIP 检查字符串是否为 IP 地址
func IsIP(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if !IsNumeric(part) {
			return false
		}
		num := 0
		for _, c := range part {
			num = num*10 + int(c-'0')
		}
		if num < 0 || num > 255 {
			return false
		}
	}
	return true
}

// IsMobile 检查字符串是否为手机号
func IsMobile(s string) bool {
	if len(s) != 11 {
		return false
	}
	if !strings.HasPrefix(s, "1") {
		return false
	}
	return IsNumeric(s)
}

// IsIDCard 检查字符串是否为身份证号
func IsIDCard(s string) bool {
	if len(s) != 18 {
		return false
	}
	for i := 0; i < 17; i++ {
		if !unicode.IsDigit(rune(s[i])) {
			return false
		}
	}
	last := s[17]
	return unicode.IsDigit(rune(last)) || last == 'X' || last == 'x'
}

// IsPassword 检查字符串是否为强密码
func IsPassword(s string) bool {
	if len(s) < 8 {
		return false
	}
	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, c := range s {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasNumber = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasNumber && hasSpecial
}
