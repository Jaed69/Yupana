# One image, three processes: the Compose service picks api, worker, bot or migrate.
FROM golang:1.26.5-trixie AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
# CGO stays off until the QR decoder (zxing-cpp) lands; then this stage needs a C++ toolchain.
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/yupana ./cmd/yupana

# No shell or package manager; runs as uid 65532.
FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/yupana /yupana
ENTRYPOINT ["/yupana"]
CMD ["api"]
