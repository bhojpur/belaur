FROM debian:buster

RUN apt-get update && apt-get install -y \
    build-essential autoconf git pkg-config \
    automake libtool curl make g++ unzip \
    && apt-get clean

# install protobuf first, then grpc
ENV GRPC_RELEASE_TAG v1.29.x
RUN git clone -b ${GRPC_RELEASE_TAG} https://github.com/grpc/grpc /var/local/git/grpc && \
	            cd /var/local/git/grpc && \
    git submodule update --init && \
    echo "--- installing Protocol Buffers ---" && \
    cd third_party/protobuf && \
    ./autogen.sh && ./configure --enable-shared && \
    make -j$(nproc) && make install && make clean && ldconfig && \
    echo "--- installing gRPC ---" && \
    cd /var/local/git/grpc && \
    make -j$(nproc) && make install && make clean && ldconfig

# Bhojpur Belaur internal port and data path.
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