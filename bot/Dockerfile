FROM python:3-alpine

ARG REQUIREMENTS_FILENAME=requirements.txt
WORKDIR /usr/src/app

# Copy just the requirements.txt files over first -- this is to make the
# container able to cache the layers with packages installed.
COPY requirements*.txt ./

# Install needed python packages.
RUN pip install --no-cache-dir -r ${REQUIREMENTS_FILENAME}

# Copy over all of our source.
COPY . ./

CMD ["python", "run.py"]
