FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go build -o task-service ./cmd/server

FROM alpine:3.19
RUN addgroup -S app && adduser -S app -G app
USER app
WORKDIR /app
COPY --from=build /app/task-service ./task-service
EXPOSE 8080
CMD ["./task-service"]
