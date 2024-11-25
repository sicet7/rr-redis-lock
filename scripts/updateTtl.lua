local key = KEYS[1]
local value = ARGV[1]
local ttl = tonumber(ARGV[2])
if ttl == nil then
    return "Invalid ttl"
end

local writeLockKey = "l:" .. key .. ":w"
local readLockKey = "l:" .. key .. ":r:" .. value

if redis.call("EXISTS", readLockKey) == 1 then
    if ttl > 0 then
        redis.call("PEXPIRE", readLockKey, ttl)
    else
        redis.call("PERSIST", readLockKey)
    end
    return 1
end

if redis.call("GET", writeLockKey) == value then
    if ttl > 0 then
        redis.call("PEXPIRE", writeLockKey, ttl)
    else
        redis.call("PERSIST", writeLockKey)
    end
    return 1
end
return 0