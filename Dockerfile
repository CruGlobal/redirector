ARG OPENRESTY_VERSION=1.21.4.1
FROM openresty/openresty:$OPENRESTY_VERSION-alpine-fat

LABEL com.datadoghq.ad.check_names='["nginx"]'
LABEL com.datadoghq.ad.init_configs='[{}]'
LABEL com.datadoghq.ad.instances='[{"nginx_status_url": "http://%%host%%:81/nginx_status/"}]'
LABEL com.datadoghq.ad.logs='[{"source": "nginx"}]'

HEALTHCHECK --interval=10s --timeout=5s CMD curl -f http://127.0.0.1:81/health-check || exit 1

# Copy container configuration
COPY docker /

# Install lua-resty-auto-ssl
ENV PATH=$PATH:/usr/local/openresty/openssl/bin
RUN LUA_LIBDIR=/usr/local/openresty/lualib /usr/local/openresty/luajit/bin/luarocks install lua-resty-auto-ssl \
    && openssl req -new -newkey rsa:2048 -days 3650 -nodes -x509 \
      -keyout /usr/local/openresty/nginx/conf/server.key \
      -out /usr/local/openresty/nginx/conf/server.crt \
    && chown -R nobody /etc/resty-auto-ssl \
    && ln -sf /usr/local/openresty/openssl/bin/openssl /usr/local/bin/openssl \
    && opm get 3scale/lua-resty-url
