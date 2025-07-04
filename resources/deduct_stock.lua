-- KEYS[1]: 库存key
-- KEYS[2]: 库存日志key
-- KEYS[3]: 锁key
-- ARGV[1]: 扣减数量
-- ARGV[2]: 订单ID
-- ARGV[3]: 用户ID
-- ARGV[4]: 超时时间(ms)
-- ARGV[5]: 当前时间戳

--尝试获取锁
local lock = redis.call("SET", KEYS[3], "LOCK", "NX", "PX", tonumber(ARGV[4]))
if not lock then
    return {"err", "E_LOCK_FAILED", "Failed to acquire lock"}
end

--获取当前库存
local stockData = redis.call("HGETALL", KEYS[1])
if #stockData == 0 then
    redis.call("DEL", KEYS[3])
    return {"err", "E_ITEM_NOT_FOUND", "Item not found"}
end
-- 解析库存数据
local currentStock = nil
local version = nil
for i = 1, #stockData, 2 do
    if stockData[i] == "stock" then
        currentStock = tonumber(stockData[i+1])
    elseif stockData[i] == "version" then
        version = tonumber(stockData[i+1])
    end
end

if currentStock == nil or version == nil then
    redis.call("DEL", KEYS[3])
    return {"err", "E_INVALID_STOCK_DATA", "Invalid stock data"}
end

local deductQty = tonumber(ARGV[1])
-- 检查库存是否充足
if currentStock < deductQty then
    redis.call("DEL", KEYS[3])
    return {"err", "E_STOCK_INSUFFICIENT", "Insufficient stock"}
end

-- 更新库存
local newStock = currentStock - deductQty
local newVersion = version + 1

redis.call("HMSET", KEYS[1],
    "stock", newStock,
    "version", newVersion,
    "modified", ARGV[5]
)

-- 记录日志
local logEntry = {
    order_id = ARGV[2],
    user_id = ARGV[3],
    item_id = string.match(KEYS[1], "item:(%d+)$"),
    quantity = deductQty,
    old_stock = currentStock,
    new_stock = newStock,
    timestamp = ARGV[5],
    is_rollback = false
}

redis.call("RPUSH", KEYS[2], cjson.encode(logEntry))

-- 释放锁
redis.call("DEL", KEYS[3])

return {"SUCCESS", newStock, newVersion}