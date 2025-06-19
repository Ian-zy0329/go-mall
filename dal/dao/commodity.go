package dao

import (
	"context"
	"github.com/Ian-zy0329/go-mall/common/util"
	"github.com/Ian-zy0329/go-mall/dal/model"
	"github.com/Ian-zy0329/go-mall/logic/do"
)

type CommodityDao struct {
	ctx context.Context
}

func NewCommodityDao(ctx context.Context) *CommodityDao {
	return &CommodityDao{ctx: ctx}
}

func (cd *CommodityDao) BulkCreateCommodityCategories(categories []*model.CommodityCategory) error {
	return DBMaster().WithContext(cd.ctx).Create(categories).Error
}

func (cd *CommodityDao) GetAllCategories() ([]*model.CommodityCategory, error) {
	commodityCategories := make([]*model.CommodityCategory, 0)
	err := DB().WithContext(cd.ctx).Find(&commodityCategories).Error
	return commodityCategories, err
}

func (cd *CommodityDao) InitCategoryData(categoryDos []*do.CommodityCategory) error {
	categoryModels := make([]*model.CommodityCategory, 0, len(categoryDos))
	util.CopyProperties(&categoryModels, &categoryDos)

	return cd.BulkCreateCommodityCategories(categoryModels)
}

func (cd *CommodityDao) GetSubCategories(parentId int64) ([]*model.CommodityCategory, error) {
	categories := make([]*model.CommodityCategory, 0)
	err := DB().WithContext(cd.ctx).
		Where("parent_id = ?", parentId).
		Order("rank DESC").Find(&categories).Error
	return categories, err
}

func (cd *CommodityDao) GetOneCommodity() (*model.Commodity, error) {
	commodity := new(model.Commodity)
	err := DB().WithContext(cd.ctx).
		Find(commodity).Error
	return commodity, err
}

func (cd *CommodityDao) InitCommodityData(commodityDos []*do.Commodity) error {
	commodityModels := make([]*model.Commodity, 0, len(commodityDos))
	util.CopyProperties(&commodityModels, &commodityDos)
	return cd.BulkCreateCommodity(commodityModels)
}

func (cd *CommodityDao) BulkCreateCommodity(commodityModels []*model.Commodity) error {
	return DBMaster().WithContext(cd.ctx).Create(commodityModels).Error
}

func (cd *CommodityDao) GetThirdLevelCategories(categoryInfo *do.CommodityCategory) (categoryIds []int64, err error) {
	if categoryInfo.Level == 3 {
		return []int64{categoryInfo.ID}, nil
	} else if categoryInfo.Level == 2 {
		categoryIds, err = cd.getSubCategoryIdList([]int64{categoryInfo.ID})
		return
	} else if categoryInfo.Level == 1 {
		var secondCategoryId []int64
		secondCategoryId, err = cd.getSubCategoryIdList([]int64{categoryInfo.ID})
		if err != nil {
			return
		}
		categoryIds, err = cd.getSubCategoryIdList(secondCategoryId)
		return
	}
	return
}

func (cd *CommodityDao) GetCategoryById(categoryId int64) (*model.CommodityCategory, error) {
	category := new(model.CommodityCategory)
	err := DB().WithContext(cd.ctx).Where("id = ?", categoryId).Find(category).Error
	return category, err
}

func (cd *CommodityDao) getSubCategoryIdList(parentCategoryIds []int64) (categoryIds []int64, err error) {
	err = DB().WithContext(cd.ctx).Model(model.CommodityCategory{}).
		Where("parent_id IN (?)", parentCategoryIds).
		Order("rank DESC").Pluck("id", &categoryIds).Error

	return
}

func (cd *CommodityDao) GetCommoditiesListInCategory(categoryIds []int64, offset, size int) (commodityModels []*model.Commodity, totalRows int64, err error) {
	err = DB().WithContext(cd.ctx).Omit("detail_content").
		Where("category_id IN (?)", categoryIds).
		Offset(offset).
		Limit(size).
		Find(&commodityModels).Error

	DB().WithContext(cd.ctx).Model(model.Commodity{}).Where("category_id IN (?)", categoryIds).Count(&totalRows)
	return
}

func (cd *CommodityDao) FindCommodityWithNameKeyword(keyword string, offset, returnSize int) (commodityList []*model.Commodity, totalRows int64, err error) {
	err = DB().WithContext(cd.ctx).Omit("detail_content").
		Where("name LIKE ?", "%"+keyword+"%").
		Offset(offset).Limit(returnSize).
		Find(&commodityList).Error
	DB().WithContext(cd.ctx).Model(model.Commodity{}).Where("name LIKE ?", "%"+keyword+"%").Count(&totalRows)

	return
}

func (cd *CommodityDao) FindCommodityById(commodityId int64) (*model.Commodity, error) {
	commodity := new(model.Commodity)
	err := DB().WithContext(cd.ctx).Where("id = ?", commodityId).Find(commodity).Error
	return commodity, err
}
