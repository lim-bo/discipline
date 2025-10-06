FROM golang:1.24.4-alpine as build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /main cmd/api/main.go

FROM alpine:latest
COPY --from=build /main /bin/main
COPY --from=build /app/configs/.env /configs/

ADD https://github.com/ufoscout/docker-compose-wait/releases/download/2.2.1/wait /wait
RUN chmod +x /wait

CMD /wait && /bin/main