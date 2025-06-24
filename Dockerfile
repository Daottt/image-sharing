
FROM golang:1.24.3

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .
COPY internal/db/schema.sql /app/schema.sql

RUN go build -v -o /app/app ./cmd/app

ENV SCHEMA_PATH="/app/schema.sql"

EXPOSE 8080

CMD ["/app/app"]
