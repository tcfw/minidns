FROM golang:1.13 as builder

WORKDIR /app
ENV GO11MODULES=on
ENV CGO_ENABLED=0

COPY . .
RUN make build

FROM alpine

COPY --from=builder /app/bin/minidns /

CMD [ "/minidns" ]