package domainservice

import (
	"context"
	"fmt"
	"github.com/Ian-zy0329/go-mall/common/app"
	"github.com/Ian-zy0329/go-mall/common/enum"
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/common/util"
	"github.com/Ian-zy0329/go-mall/dal/dao"
	"github.com/Ian-zy0329/go-mall/dal/model"
	"github.com/Ian-zy0329/go-mall/library"
	"github.com/Ian-zy0329/go-mall/logic/do"
	"github.com/samber/lo"
	"time"
)

type OrderDomainSvc struct {
	ctx      context.Context
	orderDao *dao.OrderDao
}

func NewOrderDomainSvc(ctx context.Context) *OrderDomainSvc {
	return &OrderDomainSvc{
		ctx:      ctx,
		orderDao: dao.NewOrderDao(ctx),
	}
}

func (ods *OrderDomainSvc) CreateOrder(items []*do.ShoppingCartItem, userAddressInfo *do.UserAddressInfo) (*do.Order, error) {
	billInfo, err := NewCartBillChecker(items, userAddressInfo.UserId).GetBill()
	if err != nil {
		return nil, errcode.Wrap("CreateOrderError", err)
	}
	if billInfo.OriginalTotalPrice <= 0 {
		return nil, errcode.ErrCartItemParam
	}
	order := do.OrderNew()
	order.UserId = userAddressInfo.UserId
	order.OrderNo = util.GenOrderNo(order.UserId)
	order.BillMoney = billInfo.OriginalTotalPrice
	order.PayMoney = billInfo.TotalPrice
	order.OrderStatus = enum.OrderStatusCreated
	if err = util.CopyProperties(&order.Items, &items); err != nil {
		return nil, errcode.ErrCoverData.WithCause(err)
	}
	if err = util.CopyProperties(&order.Address, &userAddressInfo); err != nil {
		return nil, errcode.ErrCoverData.WithCause(err)
	}

	tx := dao.DBMaster().Begin()
	panicked := true
	defer func() {
		if err != nil && panicked {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	err = ods.orderDao.CreateOrder(tx, order)
	if err != nil {
		return nil, err
	}
	cartDao := dao.NewCartDao(ods.ctx)
	cartItems := lo.Map(items, func(item *do.ShoppingCartItem, index int) int64 {
		return item.CartItemId
	})
	err = cartDao.DeleteMultiCartItemInTx(tx, cartItems)
	if err != nil {
		return nil, err
	}
	if billInfo.Coupon.CouponId > 0 {
		//couponDao.LockCoupon(tx,coupon)
	}
	if billInfo.Discount.DiscountId > 0 {
		//discountDao.recordDiscount(tx,discount)
	}
	commodityDao := dao.NewCommodityDao(ods.ctx)
	err = commodityDao.ReduceStuckInOrderCreate(tx, order.Items)
	if err != nil {
		return nil, err
	}
	panicked = false
	return order, nil
}

func (ods *OrderDomainSvc) GetUserOrders(userId int64, pagination *app.Pagination) ([]*do.Order, error) {
	offset := pagination.Offset()
	size := pagination.GetPageSize()
	orderModels, totalRow, err := ods.orderDao.GetUserOrders(userId, offset, size)
	if err != nil {
		return nil, errcode.Wrap("GetUserOrdersError", err)
	}
	pagination.SetTotalRows(int(totalRow))
	orders := make([]*do.Order, 0, len(orderModels))
	if err = util.CopyProperties(&orders, &orderModels); err != nil {
		return nil, errcode.ErrCoverData.WithCause(err)
	}
	orderIds := lo.Map(orders, func(order *do.Order, index int) int64 {
		return order.ID
	})
	ordersAddressMap, err := ods.orderDao.GetMultiOrdersAddress(orderIds)
	if err != nil {
		return nil, errcode.Wrap("GetMultiOrdersAddressError", err)
	}
	ordersItemMap, err := ods.orderDao.GetMultiOrdersItems(orderIds)
	if err != nil {
		return nil, errcode.Wrap("GetMultiOrdersItemsError", err)
	}
	for _, order := range orders {
		order.Address = new(do.OrderAddress)
		if err = util.CopyProperties(order.Address, ordersAddressMap[order.ID]); err != nil {
			return nil, errcode.ErrCoverData.WithCause(err)
		}
		orderItems := ordersItemMap[order.ID]
		if err = util.CopyProperties(&order.Items, &orderItems); err != nil {
			return nil, errcode.ErrCoverData.WithCause(err)
		}
	}
	return orders, nil
}

func (ods *OrderDomainSvc) GetSpecifiedUserOrder(orderNo string, userId int64) (*do.Order, error) {
	orderModel, err := ods.orderDao.GetOrderByNo(orderNo)
	if err != nil {
		return nil, errcode.Wrap("GetSpecifiedUserOrderError", err)
	}
	if orderModel == nil || orderModel.UserId != userId {
		return nil, errcode.ErrOrderParams
	}
	order := do.OrderNew()
	if err = util.CopyProperties(order, orderModel); err != nil {
		return nil, errcode.ErrCoverData.WithCause(err)
	}
	// 订单地址信息
	orderAddress, err := ods.orderDao.GetOrderAddress(orderModel.ID)
	if err != nil {
		return nil, errcode.Wrap("GetSpecifiedUserOrderError", err)
	}
	if err = util.CopyProperties(order.Address, orderAddress); err != nil {
		return nil, errcode.ErrCoverData.WithCause(err)
	}
	// 订单购物明细
	orderItems, err := ods.orderDao.GetOrderItems(orderModel.ID)
	if err != nil {
		return nil, errcode.Wrap("GetSpecifiedUserOrderError", err)
	}
	if err = util.CopyProperties(&order.Items, &orderItems); err != nil {
		return nil, errcode.ErrCoverData.WithCause(err)
	}

	return order, nil
}

func (ods *OrderDomainSvc) CancelUserOrder(orderNo string, userId int64) error {
	order, err := ods.GetSpecifiedUserOrder(orderNo, userId)
	if err != nil {
		return err
	}
	if order.OrderStatus >= enum.OrderStatusPaid {
		return errcode.ErrOrderCanNotBeChanged
	}

	err = ods.orderDao.UpdateOrderStatus(order.ID, enum.OrderStatusUserQuit)
	if err != nil {
		return errcode.Wrap("CancelOrderError", err)
	}
	commodityDao := dao.NewCommodityDao(ods.ctx)
	err = commodityDao.RecoverOrderCommodityStuck(order.Items)
	return err
}

func (ods *OrderDomainSvc) CreateOrderWxPay(orderNo string, userId int64) (payInfo *library.WxPayInvokeInfo, err error) {
	order, err := ods.GetSpecifiedUserOrder(orderNo, userId)
	if err != nil {
		return
	}
	if order.OrderStatus != enum.OrderStatusCreated {
		err = errcode.ErrOrderParams
		return
	}
	order.PayType = enum.PayTypeWxPay
	order.OrderStatus = enum.OrderStatusUnPaid
	order.PayState = enum.PayStateUnPaid
	orderModel := new(model.Order)
	if err = util.CopyProperties(orderModel, order); err != nil {
		err = errcode.Wrap("CreateOrderWxPayError", err)
		return
	}
	if err = ods.orderDao.UpdateOrder(orderModel); err != nil {
		err = errcode.Wrap("CreateOrderWxPayError", err)
		return
	}

	payInfo = &library.WxPayInvokeInfo{
		AppId:     "123456",
		TimeStamp: fmt.Sprintf("%v", time.Now().Unix()),
		NonceStr:  util.RandomString(32),
		Package:   "prepay_id=wx21201855730335ac86f8c43d18891123400",
		SignType:  "RSA",
		PaySign:   "oR9d8PuhnIc+YZ8cBHFCwfgpaK9gd7vaRvkYD7rthRAZ/X+QBhcCYL21N7cHCTUxbQ+EAt6Uy+lwSN22f5YZvI45MLko8Pfso0jm46v5hqcVwrk6uddkGuT+Cdvu4WBqDzaDjnNa5UK3GfE1Wfl2gHxIIY5lLdUgWFts17D4WuolLLkiFZV+JSHMvH7eaLdT9N5GBovBwu5yYKUR7skR8Fu+LozcSqQixnlEZUfyE55feLOQTUYzLmR9pNtPbPsu6WVhbNHMS3Ss2+AehHvz+n64GDmXxbX++IOBvm2olHu3PsOUGRwhudhVf7UcGcunXt8cqNjKNqZLhLw4jq/xDg==",
	}
	return
}
