FROM node:16.13.2-alpine

WORKDIR /app

# Install ganache-cli globally
RUN npm install -g ganache-cli

# Set the default command for the image
CMD ["ganache-cli", "-h", "0.0.0.0", "--networkId", "15"]