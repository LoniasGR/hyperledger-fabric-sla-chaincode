FROM node:16-alpine AS build

RUN apk add --no-cache g++ make python3 dumb-init

WORKDIR /app

COPY --chown=node:node . /app

RUN npm ci
RUN npm run build
RUN npm prune --production

FROM node:16-alpine
ENV NODE_ENV production
WORKDIR /app

COPY --from=build /usr/bin/dumb-init /usr/bin/dumb-init
COPY --chown=node:node --from=build /app .

EXPOSE 8000

ENTRYPOINT [ "dumb-init", "--", "npm", "run"]
CMD ["start"]