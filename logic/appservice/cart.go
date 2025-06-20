package appservice

import (
	"context"
	"github.com/Ian-zy0329/go-mall/api/reply"
	"github.com/Ian-zy0329/go-mall/api/request"
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/common/util"
	"github.com/Ian-zy0329/go-mall/logic/do"
	"github.com/Ian-zy0329/go-mall/logic/domainservice"
	"github.com/samber/lo"
)

type CartAppSvc struct {
	ctx           context.Context
	cartDomainSvc *domainservice.CartDomainSvc
}

func NewCartAppSvc(ctx context.Context) *CartAppSvc {
	return &CartAppSvc{
		ctx:           ctx,
		cartDomainSvc: domainservice.NewCartDomainSvc(ctx),
	}
}

func (cas *CartAppSvc) AddCartItem(request *request.AddCartItem, userId int64) error {
	commodityDomainSvc := domainservice.NewCommodityDomainSvc(cas.ctx)
	commodityInfo := commodityDomainSvc.GetCommodityInfo(request.CommodityId)
	if commodityInfo == nil || commodityInfo.ID == 0 { // 商品不存在
		return errcode.ErrCommodityNotExists
	}
	if commodityInfo.StockNum < request.CommodityNum {
		// 先初步判断库存是否充足, 下单时需要重新用当前读判断库存
		return errcode.ErrCommodityStockOut
	}

	cartItem := new(do.ShoppingCartItem)
	err := util.CopyProperties(&cartItem, request)
	if err != nil {
		return errcode.ErrCoverData.WithCause(err)
	}
	cartItem.UserId = userId
	return cas.cartDomainSvc.CartAddItem(cartItem)
}

func (cas *CartAppSvc) CheckCartItemBill(itemIds []int64, userId int64) (*reply.CheckedCartItemBill, error) {
	checkedCartItems, err := cas.cartDomainSvc.GetCheckedCartItems(itemIds, userId)
	if err != nil {
		return nil, err
	}
	totalPrice := lo.Reduce(checkedCartItems, func(agg int, item *do.ShoppingCartItem, index int) int {
		return agg + item.CommoditySellingPrice*item.CommodityNum
	}, 0)
	replyBill := new(reply.CheckedCartItemBill)
	err = util.CopyProperties(&replyBill.Items, checkedCartItems)
	if err != nil {
		return nil, errcode.ErrCoverData.WithCause(err)
	}
	replyBill.TotalPrice = totalPrice
	return replyBill, nil
}

func (cas *CartAppSvc) UpdateCartItem(request *request.CartItemUpdate, userId int64) error {
	return cas.cartDomainSvc.UpdateCartItem(request, userId)
}

func (cas *CartAppSvc) GetUserCartItems(userId int64) ([]*reply.CartItem, error) {
	cartItemDomains, err := cas.cartDomainSvc.GetUserCartItems(userId)
	if err != nil {
		return nil, err
	}
	cartItems := make([]*reply.CartItem, 0, len(cartItemDomains))
	err = util.CopyProperties(&cartItems, cartItemDomains)
	if err != nil {
		return nil, errcode.ErrCoverData.WithCause(err)
	}
	return cartItems, nil
}

func (cas *CartAppSvc) DeleteCartItem(userId, cartItemId int64) error {
	return cas.cartDomainSvc.DeleteCartItem(userId, cartItemId)
}

func (cas *CartAppSvc) CheckCartItemBillV2(cartItemIds []int64, userId int64) (*reply.CheckedCartItemBillV2, error) {
	checkedCartItems, err := cas.cartDomainSvc.GetCheckedCartItems(cartItemIds, userId)
	if err != nil {
		return nil, err
	}
	billChecker := domainservice.NewCartBillChecker(checkedCartItems, userId)
	billInfo := billChecker.GetBill()
	replyBillInfo := new(reply.CheckedCartItemBillV2)
	if err = util.CopyProperties(&replyBillInfo.Items, checkedCartItems); err != nil {
		return nil, errcode.ErrCoverData.WithCause(err)
	}
	if err = util.CopyProperties(&replyBillInfo.BillDetail, &billInfo); err != nil {
		return nil, errcode.ErrCoverData.WithCause(err)
	}
	return replyBillInfo, nil
}
