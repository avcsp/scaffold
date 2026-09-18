# BUILDER
FROM golang:1.27-alpine AS build

ARG ARCH=amd64

RUN apk add --no-cache make

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make build ARCH=${ARCH}

# RUNNER
FROM scratch
COPY --from=build /app/app /bin/app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt
EXPOSE 8080/tcp
CMD [ "/bin/app" ]
