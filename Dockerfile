# Build the manager binary
FROM golang:1.15 as builder
WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
WORKDIR /
RUN   microdnf update && microdnf install -y curl bash tar gzip
#shadow-utils \
#&& adduser -r -u 1000 -g 0 /
RUN   curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.18.6/bin/linux/amd64/kubectl && \
      chmod +x ./kubectl &&  mv ./kubectl /usr/local/bin/kubectl
RUN   curl -LO https://get.helm.sh/helm-v3.5.0-linux-amd64.tar.gz && tar -zxvf helm-v3.5.0-linux-amd64.tar.gz && \
      chmod +x linux-amd64/helm && mv linux-amd64/helm /usr/local/bin/helm

COPY --from=builder /workspace/manager .

RUN chmod 755 /manager

USER root

RUN cd / && mkdir -p .config/helm
RUN chmod 777 .config/helm

ENTRYPOINT ["/manager"]
