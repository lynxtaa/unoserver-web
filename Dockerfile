FROM node:22.11.0-bullseye-slim as node

FROM ubuntu:24.04

COPY --from=node /usr/local/ /usr/local/

WORKDIR /app

ENV DEBIAN_FRONTEND noninteractive

# Common libraries
RUN apt-get update && \
    apt-get install -y curl && \
    rm -rf /var/lib/apt/lists/*

# Libreoffice
RUN apt-get update && \
    apt-get install -y software-properties-common && \
    add-apt-repository ppa:libreoffice/ppa && \
    apt-get update && \
    apt-get install -y --no-install-recommends libreoffice && \
    apt-get remove -y --auto-remove software-properties-common && \
    rm -rf /var/lib/apt/lists/*

# Unoserver
RUN apt-get update && \
   apt-get install -y python3-pip && \
   pip install unoserver --break-system-packages && \
   apt-get remove -y --auto-remove python3-pip && \
   rm -rf /var/lib/apt/lists/*

RUN corepack disable && corepack enable

# Some additional MS fonts for better WMF conversion
COPY fonts/*.ttf /usr/share/fonts/

RUN fc-cache -f -v

COPY pnpm-lock.yaml package.json ./

RUN pnpm fetch

COPY . .

RUN pnpm install --offline

ARG NODE_ENV
ENV NODE_ENV=$NODE_ENV

RUN if [ "$NODE_ENV" = "production" ] ; \
  then pnpm run build && rm -rf node_modules && pnpm install --prod --ignore-scripts && pnpm store prune && rm -rf ./src ; \
  fi

# helper for reaping zombie processes
ARG TINI_VERSION=0.19.0
ADD https://github.com/krallin/tini/releases/download/v${TINI_VERSION}/tini-static /tini
RUN chmod +x /tini
ENTRYPOINT [ "/tini", "--" ]

CMD [ "node", "-r", "dotenv-safe/config", "build/index.js" ]

EXPOSE 3000
