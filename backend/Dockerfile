FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /out/api ./cmd/api

FROM alpine:3.20
RUN adduser -D -H appuser
USER appuser
COPY --from=build /out/api /api
EXPOSE 8080
CMD ["/api"]
