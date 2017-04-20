local server_name = ngx.var.host

local redis = require "resty.redis"
local red = redis:new()

red:set_timeout(1000) -- 1 second

local ok, err = red:connect(os.getenv("REDIS_PORT_6379_TCP_ADDR"), 6379)
if not ok then
    ngx.log(ngx.ERR, "failed to connect to redis: ", err)
    return ngx.exit(500)
end

-- use db number 3
red:select(3)

local target, err = red:get("domain_redirect:" .. server_name)
if not target then
    ngx.log(ngx.ERR, "failed to get redis key: ", err)
    return ngx.exit(500)
end

if target == ngx.null then
    ngx.log(ngx.ERR, "no host found for key ", server_name)
    return ngx.exit(400)
end

return ngx.redirect(target, 301)
