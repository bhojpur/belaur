FROM ruby:2.7-buster

# Version and other variables which can be changed.
ENV BELAUR_PORT=8080 \
    BELAUR_HOME_PATH=/data

# Directory for the binary
WORKDIR /app

# Copy Bhojpur Belaur binary into docker image
COPY belaur-linux-amd64 /app

# Fix permissions & setup known hosts file for ssh agent.
RUN chmod +x ./belaur-linux-amd64 \
    && mkdir -p /root/.ssh \
    && touch /root/.ssh/known_hosts \
    && chmod 600 /root/.ssh

# Set homepath as volume
VOLUME [ "${BELAUR_HOME_PATH}" ]

# Expose port
EXPOSE ${BELAUR_PORT}

# Copy entry point script
COPY docker/docker-entrypoint.sh /usr/local/bin/

# Start Bhojpur Belaur
ENTRYPOINT [ "docker-entrypoint.sh" ]