package do

import "time"

type ShoppingCartItem struct {
	CartItemId            int64
	UserId                int64
	CommodityId           int64
	CommodityName         string
	CommodityImg          string
	CommoditySellingPrice int
	CommodityNum          int
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type CartBillInfo struct {
	Coupon struct {
		CouponId      int64
		CouponName    string
		DiscountMoney int
		Threshold     int
	}
	Discount struct {
		DiscountId    int64
		DiscountName  string
		DiscountMoney int
		Threshold     int
	}
	VipDiscountMoney   int
	OriginalTotalPrice int
	TotalPrice         int
}
