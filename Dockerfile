FROM golang:1.23.1 AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /cerberus

FROM golang:1.23.1 


# AA
FROM gcr.io/distroless/base-debian12

COPY --from=build /cerberus /cerberus

EXPOSE 8999

USER nonroot:nonroot

ENTRYPOINT [ "/cerberus /props.yaml" ]