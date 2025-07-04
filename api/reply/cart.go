package reply

type CartItem struct {
	CartItemId            int64  `json:"cart_item_id"`
	UserId                int64  `json:"user_id"`
	CommodityId           int64  `json:"commodity_id"`
	CommodityNum          int    `json:"commodity_num"`
	CommodityName         string `json:"commodity_name"`                 // 商品名称
	CommodityImg          string `json:"commodity_img"`                  // 商品图片
	CommoditySellingPrice int    `json:"commodity_selling_price"`        // 商品售价
	AddCartAt             string `json:"add_cart_at" copier:"CreatedAt"` //购物车添加时间,  把Do的CreatedAt字段用copier映射到这里
}
type CheckedCartItemBill struct {
	Items      []*CartItem `json:"items"`
	TotalPrice int         `json:"total_price"`
}

type CheckedCartItemBillV2 struct {
	Items      []*CartItem `json:"items"`
	BillDetail struct {
		Coupon struct {
			CouponId      int64  `json:"coupon_id"`
			CouponName    string `json:"coupon_name"`
			DiscountMoney int    `json:"discount_money"`
		} `json:"coupon"`
		Discount struct {
			DiscountId    int64  `json:"discount_id"`
			DiscountName  string `json:"discount_name"`
			DiscountMoney int    `json:"discount_money"`
		} `json:"discount"`
		VipDiscountMoney   int `json:"vip_discount_money"`
		OriginalTotalPrice int `json:"original_total_price"`
		TotalPrice         int `json:"total_price"`
	} `json:"bill_detail"`
}
