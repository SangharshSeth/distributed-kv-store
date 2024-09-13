FROM golang:alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/main cmd/main.go

FROM alpine:edge

WORKDIR /app

COPY --from=build /app/bin/main ./bin/main

RUN apk --no-cache add ca-certificates tzdata

EXPOSE 9090

CMD ["./bin/main"]