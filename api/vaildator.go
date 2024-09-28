package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/techschool/simplebank/util"
)

// 创建自定义的验证函数
var vaildCurrency validator.Func = func(fieldlevel validator.FieldLevel) bool {
	// 尝试将字段的值断言为字符串类型
	if currency, ok := fieldlevel.Field().Interface().(string); ok {
		// 判断货币是否合法
		return util.IsSupportedCurrency(currency)
	}
	// 如果不是字符串类型，则验证失败
	return false
}
