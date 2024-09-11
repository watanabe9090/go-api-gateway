FROM golang:1.23.1 AS build

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /cerberus

FROM gcr.io/distroless/base-debian12

COPY --from=build /cerberus /cerberus

COPY  /props.docker-dev.yaml /props.docker-dev.yaml

EXPOSE 8999

USER nonroot:nonroot


ENTRYPOINT [ "/cerberus", "/props.docker-dev.yaml" ]