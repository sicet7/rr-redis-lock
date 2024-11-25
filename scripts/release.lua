local key = KEYS[1]
local value = ARGV[1]

local readLockKey = "l:" .. key .. ":r:" .. value

if redis.call("DEL" , readLockKey) > 0 then
    return 1
end

local writeLockKey = "l:" .. key .. ":w"
if redis.call("GET", writeLockKey) == value then
    if redis.call("DEL" , writeLockKey) > 0 then
        return 1
    end
end
return 0