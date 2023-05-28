ARG APP_EXECUTABLE=graviflow

FROM docker.io/library/golang:1.20-bullseye as builder

ARG APP_EXECUTABLE

WORKDIR /app

RUN go install github.com/kyleconroy/sqlc/cmd/sqlc@latest

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

RUN sqlc generate

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ${APP_EXECUTABLE} ./cmd

FROM debian:bullseye as runtime

ARG APP_EXECUTABLE

WORKDIR /var

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates

COPY --from=builder /app/${APP_EXECUTABLE} /usr/bin/

COPY ./provider/postgres/schema /var/graviflow/postgres/schema

CMD [ "${APP_EXECUTABLE}", "-h" ]