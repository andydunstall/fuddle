FROM golang:1.20-alpine3.16 AS build
WORKDIR /app
COPY . .
RUN go build -o fuddle

FROM alpine:3.16
COPY --from=build /app/fuddle /bin/fuddle
