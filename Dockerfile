FROM 056154071827.dkr.ecr.us-east-1.amazonaws.com/base-image-openresty:latest
MAINTAINER cru.org <wmd@cru.org>

# Copy container configuration
COPY config /

# Install lua-resty-auto-ssl
RUN /usr/local/openresty/luajit/bin/luarocks install lua-resty-auto-ssl \
  && chown nobody /etc/resty-auto-ssl

# Install lua-resty-url \
RUN /usr/local/openresty/luajit/bin/luarocks install lua-resty-url

WORKDIR /usr/local/openresty/nginx
