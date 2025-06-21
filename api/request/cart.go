package request

type AddCartItem struct {
	CommodityId  int64 `json:"commodity_id" binding:"required"`
	CommodityNum int   `json:"commodity_num" binding:"required" binding:"required,min=1,max=5"`
}

type CartItemUpdate struct {
	CartItemId   int64 `json:"item_id" binding:"required"`
	CommodityNum int   `json:"commodity_num" binding:"required" binding:"required,min=1,max=5"`
}

type OrderCreate struct {
	CartItemIdList []int64 `json:"cart_item_id_list" binding:"required"`
	UserAddressId  int64   `json:"user_address_id" binding:"required"`
}
