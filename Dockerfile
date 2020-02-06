FROM golang:1.13.7-alpine AS build
RUN mkdir /app
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download
COPY . /app
RUN go build -o wildcard-resolver .

FROM alpine:latest
WORKDIR /
EXPOSE 53/udp
COPY --from=build /app/wildcard-resolver .
CMD ["./wildcard-resolver"]
