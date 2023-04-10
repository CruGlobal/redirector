ARG OPENRESTY_VERSION=1.21.4.1
FROM openresty/openresty:$OPENRESTY_VERSION-alpine-fat

LABEL com.datadoghq.ad.check_names='["nginx"]'
LABEL com.datadoghq.ad.init_configs='[{}]'
LABEL com.datadoghq.ad.instances='[{"nginx_status_url": "http://%%host%%:81/nginx_status/"}]'
LABEL com.datadoghq.ad.logs='[{"source": "nginx"}]'

HEALTHCHECK --interval=10s --timeout=5s CMD curl -f http://127.0.0.1:81/health-check || exit 1

# Copy container configuration
COPY docker /

# Install lua-resty-auto-ssl and update to latest dehydrated
RUN apk add --no-cache openssl \
    && LUA_LIBDIR=/usr/local/openresty/lualib /usr/local/openresty/luajit/bin/luarocks install lua-resty-auto-ssl \
    && curl -LJOf https://raw.githubusercontent.com/dehydrated-io/dehydrated/v0.7.1/dehydrated \
    && chmod +x dehydrated \
    && mv dehydrated /usr/local/openresty/luajit/lib/luarocks/rocks-5.1/lua-resty-auto-ssl/0.13.1-1/bin/resty-auto-ssl/dehydrated \
    && opm get 3scale/lua-resty-url \
    && openssl req -new -newkey rsa:4096 -days 3650 -nodes -x509 \
      -subj "/C=US/ST=FL/O=Cru/CN=localhost" \
      -addext "subjectAltName = DNS:localhost.cru.org" \
      -keyout /usr/local/openresty/nginx/conf/server.key \
      -out /usr/local/openresty/nginx/conf/server.crt \
    && chown -R nobody /etc/resty-auto-ssl
