package dao

import (
	"context"
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/common/util"
	"github.com/Ian-zy0329/go-mall/dal/model"
	"github.com/Ian-zy0329/go-mall/logic/do"
	"gorm.io/gorm"
)

type CartDao struct {
	ctx context.Context
}

func NewCartDao(ctx context.Context) *CartDao {
	return &CartDao{
		ctx: ctx,
	}
}

func (cd *CartDao) GetUserCartItemWithCommodityId(userId, commodityId int64) (*model.ShoppingCartItem, error) {
	cartItemModel := new(model.ShoppingCartItem)
	err := DB().WithContext(cd.ctx).Where(model.ShoppingCartItem{UserId: userId, CommodityId: commodityId},
		"UserId", "CommodityId").Find(cartItemModel).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return cartItemModel, nil
}

func (cd *CartDao) UpdateCartItem(cartItem *model.ShoppingCartItem) error {
	return DB().WithContext(cd.ctx).Model(cartItem).Updates(cartItem).Error
}

func (cd *CartDao) AddCartItem(cartItem *do.ShoppingCartItem) error {
	cartItemModel := new(model.ShoppingCartItem)
	err := util.CopyProperties(&cartItemModel, cartItem)
	if err != nil {
		return errcode.ErrCoverData.WithCause(err)
	}
	return DB().WithContext(cd.ctx).Create(cartItemModel).Error
}

func (cd *CartDao) FindCartItems(cartItemIds []int64) ([]*model.ShoppingCartItem, error) {
	items := make([]*model.ShoppingCartItem, 0)
	err := DB().WithContext(cd.ctx).Find(&items, cartItemIds).Error
	return items, err
}

func (cd *CartDao) GetCartItemById(cartItemId int64) (*model.ShoppingCartItem, error) {
	cartItemModel := new(model.ShoppingCartItem)
	err := DB().WithContext(cd.ctx).Where(model.ShoppingCartItem{CartItemId: cartItemId}, "cart_item_id").Find(cartItemModel).Error
	return cartItemModel, err
}

func (cd *CartDao) GetUserCartItems(userId int64) ([]*model.ShoppingCartItem, error) {
	cartItemModels := make([]*model.ShoppingCartItem, 0)
	err := DB().WithContext(cd.ctx).Where(model.ShoppingCartItem{UserId: userId}, "UserId").Find(&cartItemModels).Error
	return cartItemModels, err
}

func (cd *CartDao) DeleteCartItem(cartItem *model.ShoppingCartItem) error {
	return DB().WithContext(cd.ctx).Delete(cartItem).Error
}

func (cd *CartDao) DeleteMultiCartItemInTx(tx *gorm.DB, cartIdList []int64) error {
	return tx.WithContext(cd.ctx).Delete(&model.ShoppingCartItem{}, cartIdList).Error
}
