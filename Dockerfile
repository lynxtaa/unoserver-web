FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ ./cmd/...

FROM ubuntu:26.04 AS base

WORKDIR /app

ENV DEBIAN_FRONTEND=noninteractive

# Libreoffice + unoserver
RUN apt-get update && \
    apt-get install -y --no-install-recommends software-properties-common && \
    add-apt-repository -y ppa:libreoffice/ppa && \
    apt-get update && \
    apt-get install -y --no-install-recommends libreoffice python3-pip && \
    pip install --no-cache-dir unoserver --break-system-packages && \
    apt-get remove -y --auto-remove software-properties-common python3-pip && \
    rm -rf /var/lib/apt/lists/* /root/.cache && \
    libreoffice --version && unoserver --version

# Some additional MS fonts for better WMF conversion
COPY fonts/*.ttf /usr/share/fonts/

RUN fc-cache -f -v

# helper for reaping zombie processes
ARG TINI_VERSION=0.19.0
ADD --chmod=755 https://github.com/krallin/tini/releases/download/v${TINI_VERSION}/tini-static /tini

FROM base AS test

COPY --from=golang:1.26 /usr/local/go /usr/local/go
ENV PATH="/usr/local/go/bin:${PATH}"

# gcc is required by the race detector
RUN apt-get update && \
    apt-get install -y --no-install-recommends gcc libc6-dev && \
    rm -rf /var/lib/apt/lists/*

FROM base AS final

COPY --from=build /out/server /usr/local/bin/server

ENTRYPOINT [ "/tini", "--" ]
CMD [ "server" ]
EXPOSE 3000
