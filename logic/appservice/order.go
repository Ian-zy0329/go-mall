package appservice

import (
	"context"
	"github.com/Ian-zy0329/go-mall/api/reply"
	"github.com/Ian-zy0329/go-mall/api/request"
	"github.com/Ian-zy0329/go-mall/common/app"
	"github.com/Ian-zy0329/go-mall/common/enum"
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/common/util"
	"github.com/Ian-zy0329/go-mall/logic/domainservice"
)

type OrderAppSvc struct {
	ctx            context.Context
	orderDomainSvc *domainservice.OrderDomainSvc
}

func NewOrderAppSvc(ctx context.Context) *OrderAppSvc {
	return &OrderAppSvc{
		ctx:            ctx,
		orderDomainSvc: domainservice.NewOrderDomainSvc(ctx),
	}
}

func (oas *OrderAppSvc) CreateOrder(request *request.OrderCreate, userId int64) (*reply.OrderCreateReply, error) {
	cartDomainSvc := domainservice.NewCartDomainSvc(oas.ctx)
	cartItems, err := cartDomainSvc.GetCheckedCartItems(request.CartItemIdList, userId)
	if err != nil {
		return nil, err
	}
	userDomainSvc := domainservice.NewUserDomainSvc(oas.ctx)
	userAddressInfo, err := userDomainSvc.GetSingleAddress(request.UserAddressId)
	if err != nil {
		return nil, err
	}
	order, err := oas.orderDomainSvc.CreateOrder(cartItems, userAddressInfo)
	if err != nil {
		return nil, err
	}
	orderReply := new(reply.OrderCreateReply)
	orderReply.OrderNo = order.OrderNo
	return orderReply, nil
}

func (oas *OrderAppSvc) GetUserOrders(userId int64, pagination *app.Pagination) ([]*reply.Order, error) {
	orders, err := oas.orderDomainSvc.GetUserOrders(userId, pagination)
	if err != nil {
		return nil, err
	}
	replyOrders := make([]*reply.Order, 0, len(orders))
	if err = util.CopyProperties(&replyOrders, &orders); err != nil {
		return nil, errcode.ErrCoverData.WithCause(err)
	}
	for _, replyOrder := range replyOrders {
		replyOrder.FrontStatus = enum.OrderFrontStatus[replyOrder.OrderStatus]
		replyOrder.Address.UserName = util.MaskRealName(replyOrder.Address.UserName)
		replyOrder.Address.UserPhone = util.MaskPhone(replyOrder.Address.UserPhone)
	}
	return replyOrders, nil
}

func (oas *OrderAppSvc) GetOrderInfo(orderNo string, userId int64) (*reply.Order, error) {
	order, err := oas.orderDomainSvc.GetSpecifiedUserOrder(orderNo, userId)
	if err != nil {
		return nil, err
	}

	replyOrder := new(reply.Order)
	if err = util.CopyProperties(replyOrder, order); err != nil {
		return nil, errcode.ErrCoverData.WithCause(err)
	}
	// 订单的前台状态
	replyOrder.FrontStatus = enum.OrderFrontStatus[replyOrder.OrderStatus]
	// 敏感信息脱敏
	replyOrder.Address.UserName = util.MaskRealName(replyOrder.Address.UserName)
	replyOrder.Address.UserPhone = util.MaskPhone(replyOrder.Address.UserPhone)

	return replyOrder, nil
}

func (oas *OrderAppSvc) CancelOrder(orderNo string, userId int64) error {
	return oas.orderDomainSvc.CancelUserOrder(orderNo, userId)
}
