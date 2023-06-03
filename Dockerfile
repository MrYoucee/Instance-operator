# Parent image to build main binary
FROM golang:1.20.4 as builder

WORKDIR /workspace

# Copy the Go Modules manifests. # Manage dependencies
COPY go.mod .
COPY go.sum .

# Cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY . .

# Build and compile
RUN CGO_ENABLED=0 GOOS=linux go build -a -o myapp main.go

# Use distroless as minimal base image to package the main binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static-debian11
COPY --from=builder /workspace/myapp .
USER nonroot:nonroot
CMD ["/myapp"]

