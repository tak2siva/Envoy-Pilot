# FROM golang:latest

# RUN go get -u github.com/lyft/protoc-gen-validate
# RUN go get -u github.com/golang/protobuf/proto
# RUN go get -u github.com/golang/protobuf/protoc-gen-go
# RUN go get -u golang.org/x/net/context
# RUN go get -u google.golang.org/grpc

# RUN mkdir -p /go/src/github.com/envoyproxy
# WORKDIR /go/src/github.com/envoyproxy/
# RUN git clone --branch v0.4 https://github.com/envoyproxy/go-control-plane.git

# RUN mkdir -p $GOPATH/src/github.com/gogo
# WORKDIR $GOPATH/src/github.com/gogo
# RUN git clone https://github.com/gogo/googleapis.git
# RUN git clone https://github.com/gogo/protobuf.git
# RUN go get github.com/derekparker/delve/cmd/dlv
# RUN go get github.com/google/uuid
# RUN go get github.com/hashicorp/consul/api
# RUN go get github.com/joho/godotenv

# RUN mkdir -p $GOPATH/src/github.com/prometheus
# WORKDIR $GOPATH/src/github.com/prometheus
# RUN git clone https://github.com/prometheus/client_golang.git
# RUN git clone https://github.com/prometheus/common.git
# RUN git clone https://github.com/prometheus/client_model.git
# RUN git clone https://github.com/prometheus/procfs.git

# RUN mkdir -p $GOPATH/src/github.com/beorn7/
# WORKDIR $GOPATH/src/github.com/beorn7/
# RUN git clone https://github.com/beorn7/perks.git

# RUN mkdir -p $GOPATH/src/github.com/matttproud
# WORKDIR $GOPATH/src/github.com/matttproud
# RUN git clone https://github.com/matttproud/golang_protobuf_extensions.git

# RUN go get github.com/rs/xid

# # TODO Compile on build
# #RUN mkdir res
# #RUN find $GOPATH/src/envoy -name '*.proto' | xargs -I % sh -c 'protoc --go_out=/res/ % --proto_path=$GOPATH/src/ \
# #       --proto_path=$GOPATH/src/github.com/lyft/protoc-gen-validate \
# #       --proto_path=/gogo-genproto/prometheus \
# #       --proto_path=/gogo-genproto/googleapis/ \
# #       --proto_path=/gogo-genproto/opencensus/proto \
# #       --proto_path=/gogo-genproto/opencensus/proto/trace'
# #

FROM golang:latest

RUN go get -u github.com/golang/dep/cmd/dep
RUN go get -u github.com/derekparker/delve/cmd/dlv

RUN mkdir /go/src/Envoy-Pilot
ADD Gopkg.lock /go/src/Envoy-Pilot/
ADD Gopkg.toml /go/src/Envoy-Pilot/

WORKDIR /go/src/Envoy-Pilot/

RUN ls -l
RUN dep ensure -vendor-only