FROM golang:1.26.4-alpine3.23 AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates git openssl

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/h0neytr4p .

RUN openssl req \
          -nodes \
          -x509 \
          -sha512 \
          -newkey rsa:4096 \
          -keyout "/out/app.key" \
          -out "/out/app.crt" \
          -days 3650 \
          -subj '/C=AU/ST=Some-State/O=Internet Widgits Pty Ltd'

FROM scratch

COPY --from=builder /out/h0neytr4p /opt/h0neytr4p/h0neytr4p
COPY --from=builder /app/traps /opt/h0neytr4p/traps
COPY --from=builder /out/app.key /opt/h0neytr4p/
COPY --from=builder /out/app.crt /opt/h0neytr4p/

WORKDIR /opt/h0neytr4p
CMD ["-cert=app.crt", "-key=app.key", "-log=log/log.json", "-catchall=false", "-payload=/opt/h0neytr4p/payloads/", "-wildcard=true", "-traps=traps/"]
ENTRYPOINT ["./h0neytr4p"]
