FROM golang:1.22 as builder

WORKDIR /workspace
COPY go.mod .
COPY go.sum .
# download dependencies
RUN go mod download
# copy source code
COPY . .
# Build
# Our old way was linux centric, we'll throw that out and use no specific arch
# This means it uses whatever your local is at the time, which works for today
#RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o main main.go
RUN CGO_ENABLED=0 GO111MODULE=on go build -a -o main main.go

# Use distroless as minimal base image
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/main .
USER 65532:65532

ENTRYPOINT ["/main", "serve"]

