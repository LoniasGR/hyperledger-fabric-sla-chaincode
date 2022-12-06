FROM node:16-alpine AS build
ENV NODE_ENV build
RUN apk add --no-cache g++ make python3 dumb-init

WORKDIR /app

COPY package.json package-lock.json tsconfig.json /app/

RUN npm ci

COPY --chown=node:node ./src /app/src

RUN npm run build
RUN npm prune --production

#------------------------------------------------------------------
#------------------------------------------------------------------

FROM alpine as fabric-build

ARG FABRIC_VERSION=2.4.7

RUN apk add --no-cache make go

RUN wget https://github.com/hyperledger/fabric/archive/refs/tags/v${FABRIC_VERSION}.tar.gz && \
    tar xzf v${FABRIC_VERSION}.tar.gz && \
    cd /fabric-${FABRIC_VERSION} && \
    make peer && \
    mv /fabric-${FABRIC_VERSION}/build/bin/peer ${HOME}

#------------------------------------------------------------------
#------------------------------------------------------------------

FROM node:16-alpine

ENV NODE_ENV production
WORKDIR /app

COPY --chown=node:node ["package.json", "package-lock.json", "/app/"]
RUN npm ci

COPY --from=build /usr/bin/dumb-init /usr/bin/dumb-init
COPY --chown=node:node --from=build /app/build /app/build
COPY --chown=node:node --from=fabric-build /root/peer /usr/bin/peer

COPY core.yaml /fabric/
COPY --chown=node:node msp /msp

EXPOSE 8000

USER node

ENTRYPOINT [ "dumb-init", "--", "npm", "run"]
CMD ["start"]