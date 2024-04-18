####################
# Development image
####################
FROM node:13.2-alpine AS development

VOLUME /src
WORKDIR /src

# Download dependencies into the image prior to copying over any source files.
# This allows the downloading dependencies step to be cached into a layer so
# that we don't have to repeat the download each time a source file changes.
COPY package.json package-lock.json /src/
RUN npm config set progress=false && \
    npm install --loglevel warn

# Copy all of the source code into the image so that it can be seen.  When
# running via docker-compose in development mode we'll override this and instead
# mount the source directory from the host machine so that we can see a live
# version of the source.
COPY . .

# Create a production build of the current code
RUN npm run-script build && \
    mv build /build

# In development mode listen on port 5000.
EXPOSE 5000

CMD npm run-script start

####################
# Production image
####################
FROM node:13.2-alpine

# Install the serve tool to use to run our app.
RUN npm install -g serve@13

# Copy the static build over.
COPY --from=development /build/ /build

WORKDIR /build
EXPOSE 5000
CMD serve -l 5000 -s
