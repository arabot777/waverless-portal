package handler

// PriceUnit 价格单位: 1 USD = 1000000
const PriceUnit = 1000000

// ToUSD 内部价格转 USD (用于返回给前端)
func ToUSD(amount int64) float64 {
	return float64(amount) / PriceUnit
}

// FromUSD USD 转内部价格 (用于接收前端输入)
func FromUSD(usd float64) int64 {
	return int64(usd * PriceUnit)
}
