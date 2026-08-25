# Development stage - full environment for active development
FROM ghcr.io/osgeo/gdal:ubuntu-small-3.8.5 AS dev

# Install Go 1.24.1
RUN apt-get update && \
    apt-get install -y wget && \
    wget https://dl.google.com/go/go1.24.1.linux-$(dpkg --print-architecture).tar.gz -O /tmp/go.tar.gz && \
    tar -C /usr/local -xzf /tmp/go.tar.gz && \
    rm /tmp/go.tar.gz && \
    apt-get remove -y wget && \
    apt-get autoremove -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Set Go environment variables
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH=/go

# Set working directory and copy Go mod files
WORKDIR /app

# zip needed for creating release-assets
# git needed for version and tag information
RUN apt-get update && \
    apt-get install -y zip git && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# We need gdal_ls in path for validate function in cloud
RUN cp /usr/lib/python3/dist-packages/osgeo_utils/samples/gdal_ls.py /bin && \
    chmod +x /bin/gdal_ls.py

# Set git safe directory
RUN git config --global --add safe.directory /app

# Builder stage - optimized for binary compilation
FROM dev AS builder

# Copy source code and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$(dpkg --print-architecture) go build -buildvcs=false \
    -ldflags="-X main.GitTag=$(git describe --tags --always --dirty) -X main.GitCommit=$(git rev-parse --short HEAD) -X main.BuildDate=$(date +%Y-%m-%d)" \
    -o /flows2fim main.go

# Production stage - minimal runtime
FROM ghcr.io/osgeo/gdal:ubuntu-small-3.8.5 AS prod

# Copy GDAL utility script
RUN cp /usr/lib/python3/dist-packages/osgeo_utils/samples/gdal_ls.py /bin && \
    chmod +x /bin/gdal_ls.py

# Copy compiled binary from builder
COPY --from=builder /flows2fim /bin/

ENTRYPOINT ["flows2fim"]