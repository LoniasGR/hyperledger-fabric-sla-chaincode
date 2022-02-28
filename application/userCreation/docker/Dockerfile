FROM node:14.13-alpine

WORKDIR /usr/src/app

COPY --chown=node:node package.json /usr/src/app/
COPY --chown=node:node yarn.lock /usr/src/app/

RUN yarn --production=true

COPY --chown=node:node . /usr/src/app/

USER node
CMD [ "yarn", "start" ]