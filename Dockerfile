# CADDY_VERSION set by build.sh based on .tool-versions file
ARG CADDY_VERSION=0
FROM public.ecr.aws/docker/library/caddy:${CADDY_VERSION}-builder-alpine AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o /usr/bin/redirector ./...

ARG CADDY_VERSION=0
FROM public.ecr.aws/docker/library/caddy:${CADDY_VERSION}-alpine

COPY --from=builder /usr/bin/redirector /usr/bin/redirector

#LABEL com.datadoghq.ad.check_names='["openmetrics"]'
#LABEL com.datadoghq.ad.init_configs='[{}]'
#LABEL com.datadoghq.ad.instances='[{"openmetrics_endpoint": "http://%%host%%:81/metrics"}]'
LABEL com.datadoghq.ad.logs='[{"source": "caddy"}]'

# Upgrade alpine packages (useful for security fixes)
RUN apk upgrade --no-cache

# Copy redirector config
COPY Caddyfile /etc/caddy/Caddyfile

CMD ["redirector", "run", "--config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile"]
