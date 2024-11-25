local key = KEYS[1]
local value = ARGV[1]

if value == "*" then
    local lockPrefix = "l:" .. key .. ":*"
    local cursor = "0"
    repeat
        local result = redis.call("SCAN", cursor, "MATCH", lockPrefix, "COUNT", 100)
        cursor = result[1]
        local keys = result[2]

        if #keys > 0 then
            return 1
        end
    until cursor == "0"
    return 0
end

local readLockKey = "l:" .. key .. ":r:" .. value

if redis.call("EXISTS", readLockKey) == 1 then
    return 1
end

local writeLockKey = "l:" .. key .. ":w"
if redis.call("GET", writeLockKey) == value then
    return 1
end
return 0