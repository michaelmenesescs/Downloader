# Use the official Go image as the base image for building
FROM golang:1.21 AS builder

# Set the working directory
WORKDIR /app

# Copy Go module files
COPY go.mod ./

# Download Go dependencies
RUN go mod download

# Copy the source code
COPY main.go ./

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o downloader .

# Use a Python base image for runtime (needed for the external tools)
FROM python:3-slim

# Set the working directory to /app
WORKDIR /app

# Install the required dependencies
RUN apt-get update -y && \
    apt-get install -y git ffmpeg && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Clone the Tidal Media Downloader repository
RUN git clone https://github.com/yaronzz/Tidal-Media-Downloader.git

# Install Tidal Media Downloader
RUN pip install -r Tidal-Media-Downloader/TIDALDL-PY/requirements.txt && \
    python Tidal-Media-Downloader/TIDALDL-PY/setup.py install

# Install SCDL
RUN pip3 install scdl

# Install yt-dlp for YouTube DJ set scraping with enhanced features
RUN pip install yt-dlp --upgrade

# Copy the Go binary from the builder stage
COPY --from=builder /app/downloader /app/downloader
RUN chmod +x /app/downloader

# Define the volume
VOLUME /app/downloads

# Set the entrypoint to the Go binary
ENTRYPOINT ["/app/downloader"]
