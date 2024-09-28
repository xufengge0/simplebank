package util

const (
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
)
// 判断货币是否合法
func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, CAD:
		return true
	default:
		return false
	}
}
