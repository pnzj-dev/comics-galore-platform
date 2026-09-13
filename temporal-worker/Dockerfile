# Build the Temporal worker as a static binary and run it on a minimal image.
FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./worker

FROM alpine:3.20
RUN adduser -D app
USER app
COPY --from=build /worker /worker
CMD ["/worker"]
