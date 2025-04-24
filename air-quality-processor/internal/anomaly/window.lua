-- KEYS[1]=zset  KEYS[2]=hash
-- ARGV[1]=ts  ARGV[2]=value  ARGV[3]=cutoff
local z,h = KEYS[1], KEYS[2]
local ts,val,cut = tonumber(ARGV[1]), tonumber(ARGV[2]), tonumber(ARGV[3])

redis.call('ZADD', z, ts, val)

local old = redis.call('ZRANGEBYSCORE', z, 0, cut)
local sum = tonumber(redis.call('HGET', h,'sum') or '0')
local cnt = tonumber(redis.call('HGET', h,'count') or '0')

sum = sum + val; cnt = cnt + 1
for _,v in ipairs(old) do
  sum = sum - tonumber(v); cnt = cnt - 1
end
if #old > 0 then
  redis.call('ZREMRANGEBYSCORE', z, 0, cut)
end
redis.call('HMSET', h, 'sum', sum, 'count', cnt)
return {sum, cnt}