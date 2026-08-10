FROM golang:1.26.5-alpine AS build

WORKDIR /src
COPY . .

RUN CGO_ENABLED=0 go build -o /out/endpoint-probe .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /out/endpoint-probe ./endpoint-probe
COPY configs ./configs

ENTRYPOINT ["./endpoint-probe"]