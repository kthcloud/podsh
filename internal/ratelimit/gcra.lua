local tat = redis.call("GET", KEYS[1])

local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local ttl = tonumber(ARGV[4])

local interval = 1000000000 / rate
local burst_offset = interval * burst

if not tat then
	tat = now
else
	tat = tonumber(tat)
end

local new_tat = math.max(tat, now) + interval
local allow_at = new_tat - burst_offset

if now < allow_at then
	local retry_after = allow_at - now
	return { 0, retry_after }
end

redis.call("SET", KEYS[1], new_tat, "PX", ttl)

return { 1, new_tat }
