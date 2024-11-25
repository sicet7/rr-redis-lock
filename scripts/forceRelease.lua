local key = KEYS[1]

local cursor = "0"
local lockPrefix = "l:" .. key .. ":*"

repeat
    local result = redis.call("SCAN", cursor, "MATCH", lockPrefix, "COUNT", 100)
    cursor = result[1]
    local keys = result[2]

    if #keys > 0 then
        redis.call("DEL", unpack(keys))
    end
until cursor == "0"
return 1