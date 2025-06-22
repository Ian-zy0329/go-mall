package request

type OrderPayCreate struct {
	OrderNo string `json:"order_no" binding:"required"`
	PayType int    `json:"pay_type" binding:"required,oneof=1 2"`
}
