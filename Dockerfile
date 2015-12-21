FROM cruglobal/base-image-openresty:latest
MAINTAINER cru.org <wmd@cru.org>

# Add self signed certs
RUN mkdir -p /usr/local/openresty/nginx/cert
COPY cru.self.crt /usr/local/openresty/nginx/cert/cru.self.crt
COPY cru.self.key /usr/local/openresty/nginx/cert/cru.self.key
# Putting these at the bottom because they're most likely to change.
COPY nginx-httpd.conf /usr/local/openresty/nginx/conf.d/webapp.conf
COPY ssl_from_redis.lua /home/app/ssl_from_redis.lua
COPY redirect.lua /home/app/redirect.lua
COPY monitor.html /home/app/webapp/public/monitor.html
