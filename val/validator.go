package val

/*
	使用gRPC时，检查用户提交的字段格式是否要求
*/
import (
	"fmt"
	"net/mail"
	"regexp"
)

var (
	isValidateUsername = regexp.MustCompile(`^[a-z0-9_]+$`).MatchString // 正则表达式用于验证用户名只能包含小写字母、数字和下划线
	isValidateFullname = regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString // 检查fullname的正则表达式
)

// 检查字符串长度是否在范围内
func ValidateString(value string, minlength, maxlength int) error {
	n := len(value)
	if n < minlength || n > maxlength {
		return fmt.Errorf("must contain from %d-%d characters", minlength, maxlength)
	}
	return nil
}

// 检查用户名是否合法
func ValidateUsername(value string) error {
	if err := ValidateString(value, 3, 100); err != nil {
		return err
	}
	if !isValidateUsername(value) {
		return fmt.Errorf("must contain only lowercase letters, digits, or underscore")
	}
	return nil
}
func ValidatePassword(value string) error {
	return ValidateString(value, 6, 100)
}
func ValidateEmail(value string) error {
	if err := ValidateString(value, 3, 100); err != nil {
		return err
	}
	// 使用mail.ParseAddress检查邮箱格式是否正确
	if _, err := mail.ParseAddress(value); err != nil {
		return fmt.Errorf("is a invalid email")
	}
	return nil
}
func ValidateFullname(value string) error {
	if err := ValidateString(value, 3, 100); err != nil {
		return err
	}
	if !isValidateFullname(value) {
		return fmt.Errorf("must contain only letters or space")
	}
	return nil
}
func ValidateEmailID(value int64) error{
	if value<=0{
		return fmt.Errorf("must be greater than 0")
	}
	return nil
}
func ValidateSecretCode(value string) error{
	if err := ValidateString(value, 32, 128); err != nil {
		return err
	}
	return nil
}