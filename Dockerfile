FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN go build -o /out/hr-cloud-service ./cmd/api

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/hr-cloud-service /app/hr-cloud-service
EXPOSE 8080
ENTRYPOINT ["/app/hr-cloud-service"]
