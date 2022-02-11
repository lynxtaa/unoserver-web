FROM ubuntu:20.04

WORKDIR /app

ENV DEBIAN_FRONTEND noninteractive

# Common libraries
RUN apt-get update && \
    apt-get install -y curl && \
    rm -rf /var/lib/apt/lists/*

# Libreoffice
RUN apt-get update && \
    apt-get install -y --no-install-recommends libreoffice && \
    rm -rf /var/lib/apt/lists/*

# Unoserver
RUN apt-get update && \
    apt-get install -y python3-pip && \
    pip install unoserver && \
    apt-get remove -y --auto-remove python3-pip && \
    rm -rf /var/lib/apt/lists/*

# NodeJS
RUN curl -fsSL https://deb.nodesource.com/setup_16.x | bash - && \
    apt-get install -y nodejs && \
    rm -rf /var/lib/apt/lists/*

COPY package*.json ./

RUN npm ci

COPY . .

ARG NODE_ENV
ENV NODE_ENV=$NODE_ENV

RUN if [ "$NODE_ENV" = "production" ] ; \
  then npm run build && npm prune --production && npm cache clean --force && rm -rf ./src ; \
  fi

# helper for reaping zombie processes
ARG TINI_VERSION=0.19.0
ADD https://github.com/krallin/tini/releases/download/v${TINI_VERSION}/tini-static /tini
RUN chmod +x /tini
ENTRYPOINT [ "/tini", "--" ]

CMD [ "npm", "start" ]

EXPOSE 3000
