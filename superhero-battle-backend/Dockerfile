# Start from a base image containing the Go runtime
FROM golang:1.18 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

## Multistage
FROM scratch as production

WORKDIR /app
COPY --from=build /app/app .

ENV GIN_MODE=release
EXPOSE 8080

CMD ["./app"]
