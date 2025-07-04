package domainservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Ian-zy0329/go-mall/common/app"
	"github.com/Ian-zy0329/go-mall/common/enum"
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/common/logger"
	"github.com/Ian-zy0329/go-mall/common/util"
	"github.com/Ian-zy0329/go-mall/dal/cache"
	"github.com/Ian-zy0329/go-mall/dal/dao"
	"github.com/Ian-zy0329/go-mall/logic/do"
	"github.com/Ian-zy0329/go-mall/resources"
	"sort"
	"time"
)

type CommodityDomainSvc struct {
	ctx          context.Context
	commodityDao *dao.CommodityDao
}

func NewCommodityDomainSvc(ctx context.Context) *CommodityDomainSvc {
	return &CommodityDomainSvc{
		ctx:          ctx,
		commodityDao: dao.NewCommodityDao(ctx),
	}
}

func (cds *CommodityDomainSvc) InitCategoryData() error {
	categories, err := cds.commodityDao.GetAllCategories()
	if err != nil {
		return errcode.Wrap("初始化商品分类错误", err)
	}
	if len(categories) > 1 {
		return errcode.Wrap("重复初始化商品分类", errors.New("不能重复初始化商品分类"))
	}
	cateInitFileHandler, _ := resources.LoadResourceFile("category_init_data.json")
	categoryDos := make([]*do.CommodityCategory, 0, len(categories))
	decoder := json.NewDecoder(cateInitFileHandler)
	decoder.Decode(&categoryDos)
	err = cds.commodityDao.InitCategoryData(categoryDos)
	if err != nil {
		return errcode.Wrap("初始化商品分类错误", err)
	}
	return nil
}

func (cds *CommodityDomainSvc) InitRedisStock() error {
	commodityModels, _ := cds.commodityDao.GetAllCommodity()
	pipeline := cache.Redis().Pipeline()
	stockItems := make([]*do.StockItem, 0)
	for _, commodity := range commodityModels {
		stockItems = append(stockItems, &do.StockItem{
			InitStock: commodity.StockNum,
			Stock:     commodity.StockNum,
			Modified:  time.Now(),
			Version:   1,
			ItemID:    commodity.ID,
		})
	}
	for _, stockItem := range stockItems {
		stockKey := fmt.Sprintf("%s%d", enum.STOCK_KEY_PREFIX, stockItem.ItemID)
		pipeline.HSet(cds.ctx, stockKey, map[string]interface{}{
			"id":        stockItem.ItemID,
			"stock":     stockItem.Stock,
			"version":   stockItem.Version,
			"modified":  stockItem.Modified.Format(time.RFC3339),
			"initStock": stockItem.InitStock,
		})
		pipeline.SAdd(cds.ctx, enum.STOCK_INIT_SETKEY, stockItem.ItemID)
		pipeline.Expire(cds.ctx, stockKey, 30*24*time.Hour) // 30天过期
	}
	if _, err := pipeline.Exec(cds.ctx); err != nil {
		return errcode.Wrap("初始化商品库存错误", err)
	}
	return nil
}

func (cds *CommodityDomainSvc) GetHierarchicCommodityCategories() []*do.HierarchicCommodityCategory {
	categoryModels, _ := cds.commodityDao.GetAllCategories()
	FlatCategories := make([]*do.HierarchicCommodityCategory, 0, len(categoryModels))
	util.CopyProperties(&FlatCategories, &categoryModels)
	sort.SliceStable(FlatCategories, func(i, j int) bool {
		if FlatCategories[i].Level != FlatCategories[j].Level {
			return FlatCategories[i].Level < FlatCategories[j].Level
		}
		if FlatCategories[i].Rank != FlatCategories[j].Rank {
			return FlatCategories[i].Rank < FlatCategories[j].Rank
		}
		return FlatCategories[i].ID < FlatCategories[j].ID
	})
	categoryTempMap := make(map[int64]*do.HierarchicCommodityCategory)
	for _, category := range FlatCategories {
		if category.ParentId == 0 {
			categoryTempMap[category.ID] = category
		} else if category.ParentId != 0 && category.Level == 2 {
			categoryTempMap[category.ID] = category
			categoryTempMap[category.ParentId].SubCategories = append(categoryTempMap[category.ParentId].SubCategories, category)
		} else if category.ParentId != 0 && category.Level == 3 {
			categoryTempMap[category.ParentId].SubCategories = append(categoryTempMap[category.ParentId].SubCategories, category)
		}
	}

	var hierarchicCategories []*do.HierarchicCommodityCategory
	for _, category := range FlatCategories {
		if category.ParentId != 0 {
			continue
		}
		category.SubCategories = categoryTempMap[category.ID].SubCategories
		for _, subCategory := range category.SubCategories {
			subCategory.SubCategories = categoryTempMap[subCategory.ID].SubCategories
		}
		hierarchicCategories = append(hierarchicCategories, category)
	}
	return hierarchicCategories
}

func (cds *CommodityDomainSvc) GetSubCategories(parentId int64) ([]*do.CommodityCategory, error) {
	categoryModels, err := cds.commodityDao.GetSubCategories(parentId)
	if err != nil {
		return nil, errcode.Wrap("GetSubCategoriesError", err)
	}
	categories := make([]*do.CommodityCategory, 0, len(categoryModels))
	util.CopyProperties(&categories, &categoryModels)

	return categories, nil
}

func (cds *CommodityDomainSvc) InitCommodityData() error {
	commodity, err := cds.commodityDao.GetOneCommodity()
	if err != nil {
		return errcode.Wrap("初始化商品错误", err)
	}
	if commodity.ID > 0 {
		return errcode.Wrap("重复初始化商品", errors.New("不能重复初始化商品"))
	}
	initDataFileReader, err := resources.LoadResourceFile("commodity_init_data.json")
	if err != nil {
		return errcode.Wrap("加载商品初始化数据错误", err)
	}
	commodityDos := make([]*do.Commodity, 0)
	decoder := json.NewDecoder(initDataFileReader)
	decoder.Decode(&commodityDos)
	err = cds.commodityDao.InitCommodityData(commodityDos)
	if err != nil {
		return errcode.Wrap("初始化商品错误", err)
	}
	return nil
}

func (cds *CommodityDomainSvc) GetCommodityListInCategory(cagegoryInfo *do.CommodityCategory, pagination *app.Pagination) ([]*do.Commodity, error) {
	offset := pagination.Offset()
	size := pagination.GetPageSize()
	thirdLevelCategoryIds, err := cds.commodityDao.GetThirdLevelCategories(cagegoryInfo)
	if err != nil {
		return nil, errcode.Wrap("获取三级分类错误", err)
	}
	commodityModelList, totalRows, err := cds.commodityDao.GetCommoditiesListInCategory(thirdLevelCategoryIds, offset, size)
	if err != nil {
		return nil, errcode.Wrap("获取商品列表错误", err)
	}
	pagination.SetTotalRows(int(totalRows))
	commodityList := make([]*do.Commodity, 0, len(commodityModelList))
	err = util.CopyProperties(&commodityList, &commodityModelList)
	if err != nil {
		return nil, errcode.ErrCoverData.WithCause(err)
	}
	return commodityList, nil
}

func (cds *CommodityDomainSvc) GetCategoryInfo(categoryId int64) *do.CommodityCategory {
	categoryModel, err := cds.commodityDao.GetCategoryById(categoryId)
	if err != nil {
		logger.New(cds.ctx).Error("GetCategoryInfoError", err)
		return nil
	}

	categoryInfo := new(do.CommodityCategory)
	util.CopyProperties(&categoryInfo, &categoryModel)
	return categoryInfo
}

func (cds *CommodityDomainSvc) SearchCommodity(keyword string, pagination *app.Pagination) ([]*do.Commodity, error) {
	offset := pagination.Offset()
	size := pagination.GetPageSize()

	commodityModelList, totalRows, err := cds.commodityDao.FindCommodityWithNameKeyword(keyword, offset, size)
	if err != nil {
		return nil, errcode.Wrap("SearchCommodityError", err)
	}
	pagination.SetTotalRows(int(totalRows))

	commodityList := make([]*do.Commodity, 0, len(commodityModelList))
	err = util.CopyProperties(&commodityList, &commodityModelList)
	if err != nil {
		return nil, errcode.ErrCoverData.WithCause(err)
	}

	return commodityList, nil
}

func (cds *CommodityDomainSvc) GetCommodityInfo(commodityId int64) *do.Commodity {
	commodityModel, err := cds.commodityDao.FindCommodityById(commodityId)
	log := logger.New(cds.ctx)
	if err != nil {
		log.Error("GetCommodityInfoError", "err", err)
		return nil
	}

	commodity := new(do.Commodity)
	err = util.CopyProperties(commodity, commodityModel)
	if err != nil {
		log.Error(errcode.ErrCoverData.Error(), "err", err)
		return nil
	}
	return commodity
}
