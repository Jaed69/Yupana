FROM golang:1.26-trixie AS build
RUN apt-get update -qq && apt-get install -y -qq --no-install-recommends cmake g++ pkg-config git >/dev/null
RUN git clone -q --depth 1 --recurse-submodules --shallow-submodules --branch v3.1.1 https://github.com/zxing-cpp/zxing-cpp.git /src/zxing-cpp
RUN cmake -S /src/zxing-cpp -B /src/build -DCMAKE_BUILD_TYPE=Release -DZXING_C_API=ON \
      -DZXING_EXAMPLES=OFF -DZXING_BLACKBOX_TESTS=OFF -DZXING_UNIT_TESTS=OFF -DZXING_PYTHON_MODULE=OFF -DBUILD_SHARED_LIBS=ON \
 && cmake --build /src/build -j"$(nproc)" && cmake --install /src/build >/dev/null && ldconfig
WORKDIR /app
COPY go.mod main.go ./
RUN go mod tidy && CGO_ENABLED=1 go build -ldflags="-s -w" -o /zxgo . && ls -la /zxgo /usr/local/lib/libZXing*

FROM debian:trixie-slim
COPY --from=build /usr/local/lib/libZXing.so* /usr/local/lib/
RUN ldconfig
COPY --from=build /zxgo /zxgo
ENTRYPOINT ["/zxgo"]
