package do

import "time"

type CommodityCategory struct {
	ID        int64     `json:"id"`
	Level     int64     `json:"level"`
	ParentId  int64     `json:"parent_id"`
	Name      string    `json:"name"`
	IconImg   string    `json:"icon_img"`
	Rank      int       `json:"rank"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type HierarchicCommodityCategory struct {
	ID            int64                          `json:"id"`
	Level         int                            `json:"level"`
	ParentId      int64                          `json:"parent_id"`
	Name          string                         `json:"name"`
	IconImg       string                         `json:"icon_img"`
	Rank          int                            `json:"rank"`
	SubCategories []*HierarchicCommodityCategory `json:"sub_categories"`
	CreatedAt     time.Time                      `json:"created_at"`
	UpdatedAt     time.Time                      `json:"updated_at"`
}

type Commodity struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Intro         string    `json:"intro"`
	CategoryId    int64     `json:"category_id"`
	CoverImg      string    `json:"cover_img"`
	Images        string    `json:"images"`
	DetailContent string    `json:"detail_content"`
	OriginalPrice int       `json:"original_price"`
	SellingPrice  int       `json:"selling_price"`
	StockNum      int       `json:"stock_num"`
	Tag           string    `json:"tag"`
	SellStatus    int       `json:"sell_status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// StockItem 库存数据结构
type StockItem struct {
	ItemID    int64     `json:"item_id"`   // 当前库存
	Stock     int       `json:"stock"`     // 当前库存
	Version   int64     `json:"version"`   // 版本号（乐观锁）
	Modified  time.Time `json:"modified"`  // 最后修改时间
	InitStock int       `json:"initStock"` // 初始库存
}

// DeductionLog 扣减日志
type DeductionLog struct {
	OrderID    string    `json:"order_id"`
	UserID     int64     `json:"user_id"`
	ItemID     int64     `json:"item_id"`
	Quantity   int64     `json:"quantity"`
	OldStock   int64     `json:"old_stock"`
	NewStock   int64     `json:"new_stock"`
	Timestamp  time.Time `json:"timestamp"`
	IsRollback bool      `json:"is_rollback"` // 是否回滚操作
}
