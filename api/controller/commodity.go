package controller

import (
	"errors"
	"github.com/Ian-zy0329/go-mall/api/request"
	"github.com/Ian-zy0329/go-mall/common/app"
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/logic/appservice"
	"github.com/gin-gonic/gin"
	"strconv"
)

func GetCategoryHierarchy(c *gin.Context) {
	svc := appservice.NewCommodityAppSvc(c)
	replyData := svc.GetHierarchicCommodityCategories()
	app.NewResponse(c).Success(replyData)
}

func GetCategoriesWithParentId(c *gin.Context) {
	parentId, _ := strconv.ParseInt(c.Query("parent_id"), 10, 64)
	svc := appservice.NewCommodityAppSvc(c)
	replyData := svc.GetSubCategories(parentId)
	app.NewResponse(c).Success(replyData)
}

func CommoditiesInCategory(c *gin.Context) {
	categoryId, _ := strconv.ParseInt(c.Query("category_id"), 10, 64)
	pagination := app.NewPagination(c)
	svc := appservice.NewCommodityAppSvc(c)
	commodityList, err := svc.GetCategoryCommodityList(categoryId, pagination)
	if err != nil {
		if errors.Is(err, errcode.ErrParams) {
			app.NewResponse(c).Error(errcode.ErrParams)
		} else {
			app.NewResponse(c).Error(errcode.ErrServer.WithCause(err))
		}
		return
	}
	app.NewResponse(c).SetPagination(pagination).Success(commodityList)
}

func CommoditySearch(c *gin.Context) {
	searchQuery := new(request.CommoditySearch)
	if err := c.ShouldBindQuery(searchQuery); err != nil {
		app.NewResponse(c).Error(errcode.ErrParams.WithCause(err))
		return
	}

	pagination := app.NewPagination(c)
	svc := appservice.NewCommodityAppSvc(c)
	commodityList, err := svc.SearchCommodity(searchQuery.Keyword, pagination)
	if err != nil {
		app.NewResponse(c).Error(errcode.ErrServer.WithCause(err))
		return
	}

	app.NewResponse(c).SetPagination(pagination).Success(commodityList)
}

func CommodityInfo(c *gin.Context) {
	commodityId, _ := strconv.ParseInt(c.Param("commodity_id"), 10, 64)
	if commodityId <= 0 {
		app.NewResponse(c).Error(errcode.ErrParams)
		return
	}

	svc := appservice.NewCommodityAppSvc(c)
	commodityInfo := svc.CommodityInfo(commodityId)
	if commodityInfo == nil {
		app.NewResponse(c).Error(errcode.ErrParams)
		return
	}

	app.NewResponse(c).Success(commodityInfo)
}
