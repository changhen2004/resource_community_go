FROM golang:1.25.6-alpine

WORKDIR /app

RUN apk add --no-cache git build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 3000

CMD ["go", "run", "."]
