package appservice

import (
	"context"
	"github.com/Ian-zy0329/go-mall/api/reply"
	"github.com/Ian-zy0329/go-mall/common/app"
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/common/logger"
	"github.com/Ian-zy0329/go-mall/common/util"
	"github.com/Ian-zy0329/go-mall/logic/domainservice"
)

type CommodityAppSvc struct {
	ctx                context.Context
	commodityDomainSvc *domainservice.CommodityDomainSvc
}

func NewCommodityAppSvc(ctx context.Context) *CommodityAppSvc {
	return &CommodityAppSvc{
		ctx:                ctx,
		commodityDomainSvc: domainservice.NewCommodityDomainSvc(ctx),
	}
}

func (cas *CommodityAppSvc) GetHierarchicCommodityCategories() []*reply.HierarchicCommodityCategory {
	hierarchicCategories := make([]*reply.HierarchicCommodityCategory, 0)
	hierarchicCategoriesDomain := cas.commodityDomainSvc.GetHierarchicCommodityCategories()
	util.CopyProperties(&hierarchicCategories, &hierarchicCategoriesDomain)
	return hierarchicCategories
}

func (cas *CommodityAppSvc) GetSubCategories(parentId int64) []*reply.CommodityCategory {
	categories, err := cas.commodityDomainSvc.GetSubCategories(parentId)
	replyData := make([]*reply.CommodityCategory, 0, len(categories))
	log := logger.New(cas.ctx)
	if err != nil {
		// 有错误返回空列表, 不阻塞前端
		log.Error("CommodityAppSvcGetSubCategoriesError", "err", err)
		return replyData
	}

	err = util.CopyProperties(&replyData, &categories)
	if err != nil {
		log.Error(errcode.ErrCoverData.Msg(), "err", err)
		return replyData
	}
	return replyData
}

func (cas *CommodityAppSvc) GetCategoryCommodityList(categoryId int64, pagination *app.Pagination) ([]*reply.CommodityListElem, error) {
	categoryInfo := cas.commodityDomainSvc.GetCategoryInfo(categoryId)
	if categoryInfo == nil || categoryInfo.ID == 0 {
		return nil, errcode.ErrParams
	}
	commodityList, err := cas.commodityDomainSvc.GetCommodityListInCategory(categoryInfo, pagination)
	if err != nil {
		return nil, err
	}
	replyCommodityList := make([]*reply.CommodityListElem, 0, len(commodityList))
	err = util.CopyProperties(&replyCommodityList, &commodityList)
	if err != nil {
		return nil, errcode.ErrCoverData.WithCause(err)
	}
	return replyCommodityList, nil
}

func (cas *CommodityAppSvc) SearchCommodity(keyword string, pagination *app.Pagination) ([]*reply.CommodityListElem, error) {
	commodityList, err := cas.commodityDomainSvc.SearchCommodity(keyword, pagination)
	if err != nil {
		return nil, err
	}
	replyCommodityList := make([]*reply.CommodityListElem, 0, len(commodityList))
	err = util.CopyProperties(&replyCommodityList, &commodityList)
	if err != nil {
		return nil, errcode.ErrCoverData.WithCause(err)
	}

	return replyCommodityList, nil
}
func (cas *CommodityAppSvc) CommodityInfo(commodityId int64) *reply.Commodity {
	commodityDO := cas.commodityDomainSvc.GetCommodityInfo(commodityId)
	if commodityDO == nil || commodityDO.ID == 0 {
		return nil
	}

	commodityInfo := new(reply.Commodity)
	util.CopyProperties(commodityInfo, commodityDO)
	return commodityInfo
}
