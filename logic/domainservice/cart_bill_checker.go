package domainservice

import (
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/logic/do"
	"github.com/samber/lo"
	"math"
)

type CartBillChecker struct {
	UserId        int64
	checkingItems []*do.ShoppingCartItem
	Coupon        struct {
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
	VipOffRate int
	handler    cartBillCheckHandler
}

func NewCartBillChecker(items []*do.ShoppingCartItem, userId int64) *CartBillChecker {
	checker := new(CartBillChecker)
	checker.UserId = userId
	checker.checkingItems = items
	checker.handler = &checkerStarter{}
	checker.handler.SetNext(&couponChecker{}).
		SetNext(&discountChecker{}).
		SetNext(&vipChecker{})
	return checker
}

type cartBillCheckHandler interface {
	RunChecker(*CartBillChecker) error
	SetNext(cartBillCheckHandler) cartBillCheckHandler
	Check(*CartBillChecker) error
}

type cartCommonChecker struct {
	nextHandler cartBillCheckHandler
}

func (n *cartCommonChecker) SetNext(handler cartBillCheckHandler) cartBillCheckHandler {
	n.nextHandler = handler
	return handler
}

func (n *cartCommonChecker) RunChecker(billChecker *CartBillChecker) (err error) {
	if n.nextHandler != nil {
		if err = n.nextHandler.Check(billChecker); err != nil {
			err = errcode.Wrap("CartBillCheckerError", err)
			return
		}
		return n.nextHandler.RunChecker(billChecker)
	}
	return
}

type couponChecker struct {
	cartCommonChecker
}

func (cc *couponChecker) Check(cbc *CartBillChecker) error {
	cbc.Coupon = struct {
		CouponId      int64
		CouponName    string
		DiscountMoney int
		Threshold     int
	}{
		CouponId:      1,
		DiscountMoney: 100,
		Threshold:     100,
	}
	return nil
}

type discountChecker struct {
	cartCommonChecker
}

func (dc *discountChecker) Check(cbc *CartBillChecker) error {
	cbc.Discount = struct {
		DiscountId    int64
		DiscountName  string
		DiscountMoney int
		Threshold     int
	}{
		DiscountId:    1,
		DiscountMoney: 100,
		Threshold:     1000,
	}
	return nil
}

type vipChecker struct {
	cartCommonChecker
}

func (vc *vipChecker) Check(cbc *CartBillChecker) error {
	cbc.VipOffRate = 0
	return nil
}

type checkerStarter struct {
	cartCommonChecker
}

func (cs *checkerStarter) Check(cbc *CartBillChecker) (err error) {
	return
}

func (cbc *CartBillChecker) GetBill() *do.CartBillInfo {
	cbc.handler.RunChecker(cbc)
	originalTotalPrice := lo.Reduce(cbc.checkingItems, func(agg int, item *do.ShoppingCartItem, index int) int {
		return agg + item.CommoditySellingPrice
	}, 0)
	vipDiscountMoney := int(math.Round(float64(originalTotalPrice) * float64(cbc.VipOffRate) / 100))
	totalPrice := originalTotalPrice - vipDiscountMoney
	if cbc.Coupon.Threshold != 0 && originalTotalPrice > cbc.Coupon.Threshold {
		totalPrice -= cbc.Coupon.DiscountMoney
	}
	if cbc.Discount.Threshold != 0 && originalTotalPrice > cbc.Discount.Threshold {
		totalPrice -= cbc.Discount.DiscountMoney
	}
	billInfo := new(do.CartBillInfo)
	billInfo.Coupon = cbc.Coupon
	billInfo.Discount = cbc.Discount
	billInfo.OriginalTotalPrice = originalTotalPrice
	billInfo.TotalPrice = totalPrice
	billInfo.VipDiscountMoney = vipDiscountMoney
	return billInfo
}
