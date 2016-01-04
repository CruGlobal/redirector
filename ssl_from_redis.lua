local log = ngx.log
local exit = ngx.exit
local null = ngx.null
local ERR = ngx.ERR
local INFO = ngx.INFO
local DEBUG = ngx.DEBUG
local HTTP_INTERNAL_SERVER_ERROR = ngx.HTTP_INTERNAL_SERVER_ERROR

-- Setup Redis connection
local redis = require "resty.redis"
local red = redis:new()

local ok, err = red:connect(os.getenv("REDIS_PORT_6379_TCP_ADDR"), "6379")
if not ok then
    log(ERR, "REDIS: Failed to connect to redis: ", err)
    return exit(HTTP_INTERNAL_SERVER_ERROR)
end

-- use db number 3
red:select(3)

-- Setup TLS related.
local ssl = require "ngx.ssl"
local server_name = ssl.server_name()
local addr, addrtyp, err = ssl.raw_server_addr()
local byte = string.byte
local key, cert

-- Local cache related
local cert_cache = ngx.shared.cert_cache
local cert_cache_duration = 7200 -- 2 hours

ssl.clear_certs()

-- Check for SNI request - if we don't have the server name, attempt to use the IP address instead.
if server_name == nil then
    log(INFO, "SNI Not present - performing IP lookup")

    -- Set server name as IP address.
    server_name = string.format("%d.%d.%d.%d", byte(addr, 1), byte(addr, 2), byte(addr, 3), byte(addr, 4))
    log(INFO, "IP Address: ", server_name)
end

-- Check cache for certficate
key  = cert_cache:get(server_name .. "_k")
cert = cert_cache:get(server_name .. "_c")

if key ~= nil and cert ~= nil then
    log(DEBUG, "Cert cache HIT for: ", server_name)
else
    log(DEBUG, "Cert cache MISS for: ", server_name)

    -- If the cert isn't in the cache, attept to retrieve from Redis
    local domain, err = red:get("domain:" .. server_name)
    local cert_pair, err = red:hmget("certs:" .. domain, "key", "cert")

    if cert_pair[1] == null then
        log(ERR, "failed to retreive certificates for domain: ", server_name, " Err: ", err)
                local cert_path = '/usr/local/openresty/nginx/cert/cru.self.crt'
        local f = assert(io.open(cert_path))
        local cert_data = f:read("*a")
        f:close()

        local cert_data_der, err = ssl.cert_pem_to_der(cert_data)

        local ok, err = ssl.set_der_cert(cert_data_der)
        if not ok then
          ngx.log(ngx.ERR, "failed to set DER cert: ", err)
          return
        end

        local key_path = '/usr/local/openresty/nginx/cert/cru.self.key'
        local f = assert(io.open(key_path))
        local pkey_data = f:read("*a")
        f:close()

        local key_der, err = ssl.priv_key_pem_to_der(pkey_data)

        local ok, err = ssl.set_der_priv_key(key_der)

        if not ok then
          ngx.log(ngx.ERR, "failed to set DER key: ", err)
          return
        end
        return
    end

    key = cert_pair[1]
    cert = cert_pair[2]

    -- If we've retrieved the cert and key, attempt to cache
    if key ~= nil and cert ~= nil then

        -- Add key and cert to the cache
        local success, err, forcible = cert_cache:set(server_name .. "_k", key, cert_cache_duration)
        log(DEBUG, "Caching Result: ", success, " Err: ",  err)

        success, err, forcible = cert_cache:set(server_name .. "_c", cert, cert_cache_duration)
        log(DEBUG, "Caching Result: ", success, " Err: ",  err)

        log(DEBUG, "Cert and key retrieved and cached for: ", server_name)

    else
        log(ERR, "Failed to retrieve " .. (key and "" or "key ") .. (cert and "" or "cert "), "for ", server_name)
        return
    end
end

-- Set cert
local cert_der, err = ssl.cert_pem_to_der(cert)

local ok, err = ssl.set_der_cert(cert_der)
if not ok then
    log(ERR, "failed to set DER cert: ", err)
    return
end

-- Set key
local key_der, err = ssl.priv_key_pem_to_der(key)

local ok, err = ssl.set_der_priv_key(key_der)
if not ok then
    log(ERR, "failed to set DER key: ", err)
    return
end
