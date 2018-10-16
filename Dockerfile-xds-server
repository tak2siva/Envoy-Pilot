FROM tak2siva/xds-build:v0.1.6 as builder

RUN go get -u github.com/derekparker/delve/cmd/dlv
ADD cmd /go/src/Envoy-Pilot/cmd

RUN rm -rf  /go/src/github.com/envoyproxy/go-control-plane/vendor
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/Envoy-Pilot /go/src/Envoy-Pilot/cmd/server/main.go

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
RUN mkdir -p /go/bin/
WORKDIR /go/bin/
COPY --from=builder /go/bin/Envoy-Pilot .
CMD ["./Envoy-Pilot"]