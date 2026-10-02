FROM node:16.18
RUN apt-get update && \
      apt-get -y install sudo
WORKDIR /app/dawinchi-api
COPY dawinchi-api/package.json .
ARG NODE_ENV
RUN if [ "$NODE_ENV" = "development" ]; \
        then npm install; \
        else npm install --only=production; \
        fi
COPY . ../
ENV PORT 3500
EXPOSE $PORT
RUN npm install -g nodemon

RUN adduser --disabled-password --gecos '' admin
RUN adduser admin sudo
RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

USER admin
CMD ["node", "app.mjs"]