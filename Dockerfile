FROM golang:1.18-alpine3.17 AS build
WORKDIR /src
COPY . .
RUN apk add --no-cache gcc musl-dev
RUN go build

FROM alpine:3.17
COPY --from=build /src/dash-backend /bin/
EXPOSE 8080
ENTRYPOINT ["/bin/dash-backend"]
