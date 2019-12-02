####################
# Development image
####################
FROM node:13.2-alpine AS development

VOLUME /src
WORKDIR /src

# Download dependencies into the image prior to copying over any source files.
# This allows the downloading dependencies step to be cached into a layer so
# that we don't have to repeat the download each time a source file changes.
COPY package.json /src
COPY package-lock.json /src
RUN npm install

# Copy all of the source code into the image so that it can be seen.  When
# running via docker-compose in development mode we'll override this and instead
# mount the source directory from the host machine so that we can see a live
# version of the source.
COPY . .

# npm run build?

####################
# Production image
####################

EXPOSE 3000
