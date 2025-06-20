package domainservice

import (
	"context"
	"github.com/Ian-zy0329/go-mall/api/request"
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/common/logger"
	"github.com/Ian-zy0329/go-mall/common/util"
	"github.com/Ian-zy0329/go-mall/dal/dao"
	"github.com/Ian-zy0329/go-mall/dal/model"
	"github.com/Ian-zy0329/go-mall/logic/do"
	"github.com/samber/lo"
)

type CartDomainSvc struct {
	ctx     context.Context
	cartDao *dao.CartDao
}

func NewCartDomainSvc(ctx context.Context) *CartDomainSvc {
	return &CartDomainSvc{
		ctx:     ctx,
		cartDao: dao.NewCartDao(ctx),
	}
}

func (cds *CartDomainSvc) CartAddItem(cartItem *do.ShoppingCartItem) error {
	cartItemModel, err := cds.cartDao.GetUserCartItemWithCommodityId(cartItem.UserId, cartItem.CommodityId)
	if err != nil {
		return errcode.Wrap("CartAddItemError", err)
	}
	if cartItemModel != nil && cartItemModel.CartItemId != 0 {
		cartItemModel.CommodityNum += cartItem.CommodityNum
		return cds.cartDao.UpdateCartItem(cartItemModel)
	}
	err = cds.cartDao.AddCartItem(cartItem)
	if err != nil {
		err = errcode.Wrap("CartAddItemError", err)
	}
	return err
}

func (cds *CartDomainSvc) GetCheckedCartItems(cartItemIds []int64, userId int64) ([]*do.ShoppingCartItem, error) {
	cartItemModels, err := cds.cartDao.FindCartItems(cartItemIds)
	if err != nil {
		err = errcode.Wrap("GetCheckedCartItemsError", err)
		return nil, err
	}
	userCartItemModels := lo.Filter(cartItemModels, func(item *model.ShoppingCartItem, index int) bool {
		return item.UserId == userId
	})
	if len(userCartItemModels) != len(cartItemIds) {
		return nil, errcode.ErrCartWrongUser
	}
	userCartItems := make([]*do.ShoppingCartItem, 0, len(userCartItemModels))
	err = util.CopyProperties(&userCartItems, &cartItemModels)
	if err != nil {
		return nil, errcode.ErrCoverData.WithCause(err)
	}
	cds.fillInCommodityInfo(userCartItems)
	return userCartItems, nil
}

func (cds *CartDomainSvc) fillInCommodityInfo(cartItems []*do.ShoppingCartItem) error {
	// 获取购物项中的商品信息
	commodityDao := dao.NewCommodityDao(cds.ctx)
	commodityIdList := lo.Map(cartItems, func(item *do.ShoppingCartItem, index int) int64 {
		return item.CommodityId
	})
	commodities, err := commodityDao.FindCommodities(commodityIdList)
	if err != nil {
		return errcode.Wrap("CartItemFillInCommodityInfoError", err)
	}
	if len(commodities) != len(cartItems) {
		logger.New(cds.ctx).Error("fillInCommodityError", "err", "商品信息不匹配", "commodityIdList", commodityIdList,
			"fetchedCommodities", commodities)
		return errcode.ErrCartItemParam
	}
	// 转换成以ID为Key的商品Map
	commodityMap := lo.SliceToMap(commodities, func(item *model.Commodity) (int64, *model.Commodity) {
		return item.ID, item
	})
	for _, cartItem := range cartItems {
		cartItem.CommodityName = commodityMap[cartItem.CommodityId].Name
		cartItem.CommodityImg = commodityMap[cartItem.CommodityId].CoverImg
		cartItem.CommoditySellingPrice = commodityMap[cartItem.CommodityId].SellingPrice
	}

	return nil
}

func (cds *CartDomainSvc) UpdateCartItem(request *request.CartItemUpdate, userId int64) error {
	cartItemModel, err := cds.cartDao.GetCartItemById(request.CartItemId)
	if err != nil {
		err = errcode.Wrap("CartUpdateItemError", err)
		return err
	}
	if cartItemModel == nil || cartItemModel.UserId != userId {
		logger.New(cds.ctx).Error("DataMatchError", "cartItem", cartItemModel, "request", request, "requestUserId", userId)
		return errcode.ErrParams
	}

	cartItemModel.CommodityNum = request.CommodityNum
	err = cds.cartDao.UpdateCartItem(cartItemModel)
	if err != nil {
		err = errcode.Wrap("CartUpdateItemError", err)
	}

	return err
}

func (cds *CartDomainSvc) GetUserCartItems(userId int64) (cartItems []*do.ShoppingCartItem, err error) {
	cartItemModels, err := cds.cartDao.GetUserCartItems(userId)
	if err != nil {
		err = errcode.Wrap("CartGetUserCartItemsError", err)
		return
	}
	cartItems = make([]*do.ShoppingCartItem, len(cartItemModels))
	util.CopyProperties(&cartItems, cartItemModels)
	if err != nil {
		return nil, errcode.ErrCoverData.WithCause(err)
	}
	err = cds.fillInCommodityInfo(cartItems)
	if err != nil {
		return nil, err
	}
	return cartItems, err
}

func (cds *CartDomainSvc) DeleteCartItem(userId, itemId int64) error {
	cartItemModel, _ := cds.cartDao.GetCartItemById(itemId)
	if cartItemModel == nil || cartItemModel.UserId != userId {
		logger.New(cds.ctx).Error("DataMatchError", "cartItem", cartItemModel, "cartItemId", itemId, "userId", userId)
		return errcode.ErrParams
	}
	err := cds.cartDao.DeleteCartItem(cartItemModel)
	if err != nil {
		err = errcode.Wrap("DeleteCartItemError", err)
	}
	return err
}
