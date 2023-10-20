# Use Golang Docker Image
FROM golang:1.21
# Set working dir
WORKDIR /home/rly
# Install git
RUN apt-get install git
# Clone relayer repository and install relayer
RUN git clone --single-branch --branch v2.4.2 --depth 1 https://github.com/cosmos/relayer.git && cd relayer && git checkout v2.4.2 && make build && cp ./build/rly /usr/local/bin/
RUN chmod +x /usr/local/bin/rly
