# build phase
FROM golang:1.25.0 AS builder

WORKDIR /app

COPY go.mod go.sum .
COPY main.go .


ENV CGO_ENABLED=0 GOOS=linux 

RUN go mod tidy && go build -o main . 

# final phase
FROM alpine:latest

ARG PORT=8082
ENV PORT=${PORT}

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE ${PORT}

CMD ["./main"]

