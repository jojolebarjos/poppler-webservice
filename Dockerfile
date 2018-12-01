
FROM golang:alpine as builder

RUN apk --no-cache add \
    cmake \
    fontconfig-dev \
    freetype-dev \
    git \
    gcc \
    g++ \
    make \
    upx \
    zlib-dev

WORKDIR /opt

RUN \
    git clone https://anongit.freedesktop.org/git/poppler/poppler.git && \
    mkdir build && \
    cd build && \
    cmake \
        -DBUILD_CPP_TESTS=OFF \
        -DBUILD_GTK_TESTS=OFF \
        -DBUILD_QT5_TESTS=OFF \
        -DCMAKE_BUILD_TYPE=Release \
        -DENABLE_CPP=OFF \
        -DENABLE_DCTDECODER=none \
        -DENABLE_GLIB=OFF \
        -DENABLE_GOBJECT_INTROSPECTION=OFF \
        -DENABLE_LIBCURL=OFF \
        -DENABLE_LIBOPENJPEG=none \
        -DENABLE_QT5=OFF \
        -DENABLE_SPLASH=OFF \
        -DCMAKE_INSTALL_LIBDIR=lib \
        ../poppler && \
    make && \
    make install

COPY server.go .

RUN \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" server.go && \
    upx -q --brute server

RUN \
    mkdir /tmp/root && \
    mkdir /tmp/root/bin && \
    mkdir /tmp/root/lib && \
    ldd `which pdftotext` | awk '{ if ($2 == "=>") print $3; else print $1; }' | xargs -I '{}' cp '{}' /tmp/root/lib && \
    ldd server | awk '{ if ($2 == "=>") print $3; else print $1; }' | xargs -I '{}' cp '{}' /tmp/root/lib && \
    cp /usr/local/bin/pdftotext server /tmp/root/bin

RUN adduser -D -g '' user

FROM scratch

MAINTAINER Jojo le Barjos (jojolebarjos@gmail.com)

COPY --from=builder /tmp/root /

COPY --from=builder /etc/passwd /etc/passwd

EXPOSE 8080/tcp

USER user

ENTRYPOINT ["/bin/server"]
