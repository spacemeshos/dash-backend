FROM golang:1.22.5-alpine AS build
WORKDIR /src
COPY . .
RUN apk add --no-cache gcc musl-dev
RUN go build

FROM alpine:3.17
COPY --from=build /src/dash-backend /bin/
EXPOSE 5000
ENTRYPOINT ["/bin/dash-backend"]
