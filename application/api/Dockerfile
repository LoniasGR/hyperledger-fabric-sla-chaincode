FROM node:16-alpine as builder

WORKDIR /app

COPY ["package.json", "package-lock.json",  "/app/"]

RUN npm install

COPY . /app/

RUN npm run build-dist

# step 2
FROM node:16-alpine

ENV NODE_ENV=production
ARG EXPRESS_PORT=8001

WORKDIR /app

COPY ["package.json", "package-lock.json",  "/app/"]

RUN npm install --production

COPY --from=builder /app/build /app/build
COPY .env .

EXPOSE ${EXPRESS_PORT}

CMD [ "npm", "start" ]
