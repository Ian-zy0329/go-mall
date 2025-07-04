package enum

const (
	REDIS_KEY_DEMO_ORDER_DETAIL = "GOMALL:DEMO:ORDER_DETAIL_%S"
)
const (
	REDIS_KEY_ACCESS_TOKEN       = "GOMALL:USER:ACCESS_TOKEN_%s"
	REDIS_KEY_REFRESH_TOKEN      = "GOMALL:USER:REFRESH_TOKEN_%s"
	REDIS_KEY_USER_SESSION       = "GOMALL:USER:SESSION_%d"
	REDISKEY_TOKEN_REFRESH_LOCK  = "GOMALL:USER:TOKEN_REFRESH_LOCK_%s"
	REDISKEY_PASSWORDRESET_TOKEN = "GOMALL:USER:PASSWORD_RESET_TOKEN_%s"
)

// Redis 库存数据结构
const (
	STOCK_KEY_PREFIX      = "mall:stock:item:" // 商品库存主数据 key
	STOCK_LOCK_KEY_PREFIX = "mall:stock:lock:" // 库存锁 key
	STOCK_LOG_KEY_PREFIX  = "mall:stock:log:"  // 库存流水 key
	STOCK_INIT_SETKEY     = "mall:stock:init"  // 已初始化商品集合
)
