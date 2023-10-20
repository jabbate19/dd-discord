FROM docker.io/golang:1.19.2-alpine3.16 AS build

WORKDIR /src/
RUN apk add git
COPY . .
ENV GIN_MODE=release
RUN go build -v -o dd-discord

FROM docker.io/alpine:3.16
COPY --from=build /src/dd-discord /dd-discord

ENTRYPOINT [ "/dd-discord" ]