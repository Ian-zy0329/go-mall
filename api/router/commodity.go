package router

import (
	"github.com/Ian-zy0329/go-mall/api/controller"
	"github.com/gin-gonic/gin"
)

func registerCommodityRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/commodity/")
	g.GET("category-hierarchy/", controller.GetCategoryHierarchy)
	// 按ParentID 查询商品分类列表
	g.GET("category/", controller.GetCategoriesWithParentId)
	g.GET("commodity-in-cate", controller.CommoditiesInCategory)
	g.GET("search", controller.CommoditySearch)
	g.GET(":commodity_id/info", controller.CommodityInfo)
}
