package dao

import (
	"context"
	"errors"
	"fmt"
	"github.com/Ian-zy0329/go-mall/common/enum"
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/common/util"
	"github.com/Ian-zy0329/go-mall/dal/cache"
	"github.com/Ian-zy0329/go-mall/dal/model"
	"github.com/Ian-zy0329/go-mall/logic/do"
	"github.com/Ian-zy0329/go-mall/resources"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io"
	"strconv"
	"time"
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

func (cd *CommodityDao) GetAllCommodity() (commodityModels []*model.Commodity, err error) {
	// 查询所有商品信息，不带条件查询全表数据
	err = DB().WithContext(cd.ctx).
		Model(&model.Commodity{}). // 明确指定操作的模型
		Find(&commodityModels).Error
	if err != nil {
		// 如果发生错误，返回空结果和错误信息
		return nil, err
	}
	return commodityModels, nil
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

func (cd *CommodityDao) FindCommodities(commodityIdList []int64) ([]*model.Commodity, error) {
	commodities := make([]*model.Commodity, 0)
	err := DB().WithContext(cd.ctx).Find(&commodities, commodityIdList).Error
	return commodities, err
}

func (cd *CommodityDao) ReduceStuckInOrderCreate(tx *gorm.DB, orderItems []*do.OrderItem) error {
	for _, orderItem := range orderItems {
		commodity := new(model.Commodity)
		tx.Clauses(clause.Locking{Strength: "UPDATE"}).WithContext(cd.ctx).
			Find(commodity, orderItem.CommodityId)
		newStock := commodity.StockNum - orderItem.CommodityNum
		if newStock < 0 {
			return errcode.ErrCommodityStockOut.WithCause(errors.New("商品缺少库存，商品ID:" + strconv.FormatInt(commodity.ID, 10)))
		}
		commodity.StockNum = newStock
		err := tx.WithContext(cd.ctx).Model(commodity).Update("stock_num", newStock).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// ReduceStuckInOrderCreateByLua 库存原子扣减
func (cd *CommodityDao) ReduceStuckInOrderCreateByLua(tx *gorm.DB, orderItems []*do.OrderItem, userId int64) error {
	redisStockService := cache.RedisStockService()
	deductStockLuaHandler, _ := resources.LoadResourceFile("deduct_stock.lua")
	scriptContent, err := io.ReadAll(deductStockLuaHandler)
	if err != nil {
		panic(err)
	}
	// 使用 redis.NewScript 来封装 Lua 脚本
	script := redis.NewScript(string(scriptContent))
	sha, err := script.Load(cd.ctx, redisStockService).Result()
	if err != nil {
		return errcode.Wrap("failed to load lua script: %w", err)
	}
	for _, orderItem := range orderItems {
		stockKey := fmt.Sprintf("%s%d", enum.STOCK_KEY_PREFIX, orderItem.CommodityId)
		logKey := fmt.Sprintf("%s%d", enum.STOCK_LOG_KEY_PREFIX, orderItem.CommodityId)
		lockKey := fmt.Sprintf("%s%d", enum.STOCK_LOCK_KEY_PREFIX, orderItem.CommodityId)
		// 准备参数
		args := []interface{}{
			orderItem.CommodityNum,
			orderItem.OrderId,
			userId,
			500, // 500ms 锁超时
			time.Now().Format(time.RFC3339),
		}
		//执行脚本
		result, err := redisStockService.EvalSha(cd.ctx, sha, []string{stockKey, logKey, lockKey}, args...).Result()
		if err != nil {
			return errcode.Wrap("deduction failed: %w", err)
		}
		// 解析结果
		if resultMap, ok := result.([]interface{}); ok {
			if len(resultMap) > 0 && resultMap[0] == "err" {
				return errors.New(resultMap[2].(string))
			}

			// 成功返回
			status := resultMap[0].(string)
			if status == "SUCCESS" {
				return nil
			}
		}
	}
	return errors.New("unknown deduction error")
}

func (cd *CommodityDao) RecoverOrderCommodityStuck(orderItems []*do.OrderItem) error {
	err := DBMaster().Transaction(func(tx *gorm.DB) error {
		for _, orderItem := range orderItems {
			commodity := new(model.Commodity)
			tx.Clauses(clause.Locking{Strength: "UPDATE"}).WithContext(cd.ctx).Find(commodity, orderItem.CommodityId)
			if commodity.ID == 0 {
				return errcode.ErrNotFound.WithCause(errors.New(fmt.Sprintf("商品未找到，ID：%d", orderItem.CommodityId)))
			}
			newStock := commodity.StockNum + orderItem.CommodityNum
			err := tx.WithContext(cd.ctx).Model(commodity).Update("stock_num", newStock).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
