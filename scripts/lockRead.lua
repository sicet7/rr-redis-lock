local key = KEYS[1]
local value = ARGV[1]
local ttl = tonumber(ARGV[2])
if ttl == nil then
    return "Invalid ttl"
end

local storeType = redis.call("TYPE", key)["ok"]

if storeType ~= "none" and storeType ~= "hash" then
    return 0
end

redis.call("HSET", key, value, "f")
if ttl > 0 then
    redis.call("HPEXPIRE", key, value)
end

if storeType == "hash" then
    local result = redis.call("HGET", key, value)
    if result ~= nil and result ~= false then
        return 1
    end
end