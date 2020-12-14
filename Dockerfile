FROM golang:1.14.6-alpine3.12 AS build
WORKDIR /src
COPY . .
RUN go build

FROM alpine:3.12
COPY --from=build /src/dash-backend /bin/
EXPOSE 8080
ENTRYPOINT ["/bin/dash-backend"]
