local key = KEYS[1]
local value = ARGV[1]
local ttl = tonumber(ARGV[2])
if ttl == nil then
    return "Invalid ttl"
end

local storeType = redis.call("TYPE", key)["ok"]

if storeType == "none" or (storeType == "string" and redis.call("GET", key) == value) then
    redis.call("SET", key, value)
    if ttl > 0 then
        redis.call("PEXPIRE", key, ttl)
    end
    return 1
end
return 0