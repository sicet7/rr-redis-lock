local key = KEYS[1]
local value = ARGV[1]

local storeType = redis.call("TYPE", key)["ok"]

if storeType == "none" then
    return 0
elseif storeType ~= "none" and value == "*" then
    return 1
elseif storeType == "hash" then
    local result = redis.call("HGET", key, value)
    if result ~= nil and result ~= false then
        return 1
    end
elseif storeType == "string" then
    if redis.call("GET", key) == value then
        return 1
    end
end
return 0