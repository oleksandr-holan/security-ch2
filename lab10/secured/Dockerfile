FROM golang:1.25-alpine

WORKDIR /app

# Install git to fetch dependencies if needed
RUN apk add --no-cache git

# Copy and download deps
COPY main.go .
RUN go mod init sql-lab && go get github.com/lib/pq && go mod tidy

# Build
RUN go build -o server main.go

# Copy static files
COPY static ./static

CMD ["./server"]