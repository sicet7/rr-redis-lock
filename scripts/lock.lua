local key = KEYS[1]
local value = ARGV[1]
local ttl = tonumber(ARGV[2])
if ttl == nil then
    return "Invalid ttl"
end

local writeLockKey = "l:" .. key .. ":w"

if redis.call("EXISTS", writeLockKey) == 1 then
    if redis.call("GET", writeLockKey) == value then
        if ttl > 0 then
            redis.call("PEXPIRE", writeLockKey, ttl)
        else
            redis.call("PERSIST", writeLockKey)
        end
        return 1
    else
        return 0
    end
end

local readLockPrefix = "l:" .. key .. ":r:*"
local cursor = "0"
repeat
    local result = redis.call("SCAN", cursor, "MATCH", readLockPrefix, "COUNT", 100)
    cursor = result[1]
    local keys = result[2]

    if #keys > 0 then
        return 0
    end
until cursor == "0"

if ttl > 0 then
    redis.call("SET", writeLockKey, value, "PX", ttl)
else
    redis.call("SET", writeLockKey, value)
end
return 1