
#FROM alpine:3.8
FROM golang:alpine

MAINTAINER Jojo le Barjos (jojolebarjos@gmail.com)

RUN apk --no-cache add \
    cmake \
    fontconfig-dev \
    freetype-dev \
    git \
    gcc \
    g++ \
    make \
    zlib-dev

WORKDIR /opt

# TODO ENABLE_DCTDECODER
# TODO ENABLE_LIBOPENJPEG (use libjpeg-turbo?)
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

# TODO minimize size
RUN go build server.go

# TODO use docker builder and copy result to scratch container

EXPOSE 8080/tcp

ENTRYPOINT ["./server"]
