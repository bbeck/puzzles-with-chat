FROM python:3-alpine

ARG REQUIREMENTS_FILENAME=requirements.txt
ARG TIMEZONE="US/Central"
ENV TZ=${TIMEZONE}
WORKDIR /usr/src/app

# Copy just the requirements.txt files over first -- this is to make the
# container able to cache the layers with packages installed.
COPY requirements*.txt ./

# Some of our packages require native binaries to be compiled when installed
# so we'll need to temporarily add gcc and some headers to the container.
RUN apk --no-cache add --virtual .deps gcc musl-dev python3-dev libffi-dev openssl-dev && \
  pip install --no-cache-dir -r ${REQUIREMENTS_FILENAME} && \
  apk del .deps && \
  rm -rf /var/cache/apk/*

# Copy over all of our source.
COPY . ./

CMD ["python", "server.py"]