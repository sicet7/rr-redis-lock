local key = KEYS[1]
local value = ARGV[1]
local ttl = tonumber(ARGV[2])
if ttl == nil then
    return "Invalid ttl"
end

local writeLockKey = "l:" .. key .. ":w"

-- cannot lock if exclusive lock is set
if redis.call("EXISTS", writeLockKey) == 1 then
    return 0
end

local readLockKey = "l:" .. key .. ":r:" .. value
if ttl > 0 then
    redis.call("SET", readLockKey, "val", "PX", ttl)
else
    redis.call("SET", readLockKey, "val")
end
return 1