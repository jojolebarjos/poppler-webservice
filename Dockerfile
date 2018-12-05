
FROM golang:alpine as builder

RUN apk --no-cache add \
    cairo-dev \
    cmake \
    fontconfig-dev \
    freetype-dev \
    git \
    gcc \
    g++ \
    lcms2-dev \
    libjpeg-turbo-dev \
    libpng-dev \
    make \
    openjpeg-dev \
    openjpeg-tools \
    tiff-dev \
    zlib-dev

WORKDIR /opt

RUN \
    git clone --depth 1 https://anongit.freedesktop.org/git/poppler/poppler.git && \
    cd poppler && \
    ln -s /usr/include/openjpeg-*/* /usr/include/ && \
    cmake \
        -DBUILD_CPP_TESTS=OFF \
        -DBUILD_GTK_TESTS=OFF \
        -DBUILD_QT5_TESTS=OFF \
        -DCMAKE_BUILD_TYPE=Release \
        -DENABLE_CPP=OFF \
        -DENABLE_DCTDECODER=libjpeg \
        -DENABLE_GLIB=OFF \
        -DENABLE_GOBJECT_INTROSPECTION=OFF \
        -DENABLE_LIBCURL=OFF \
        -DENABLE_LIBOPENJPEG=openjpeg2 \
        -DENABLE_QT5=OFF \
        -DENABLE_SPLASH=OFF \
        -DCMAKE_INSTALL_LIBDIR=lib && \
    make && \
    make install

COPY server.go .

RUN go build -ldflags="-w -s" server.go

RUN \
    mkdir /tmp/root && \
    mkdir /tmp/root/bin && \
    mkdir /tmp/root/lib && \
    cp /usr/local/bin/pdftotext /usr/local/bin/pdftocairo server /tmp/root/bin && \
    { ldd /tmp/root/bin/pdftotext; ldd /tmp/root/bin/pdftocairo; ldd /tmp/root/bin/server; } | awk '{ if ($2 == "=>") print $3; else print $1; }' > deps.txt && \
    cat deps.txt && \
    xargs -I '{}' cp '{}' /tmp/root/lib < deps.txt

RUN adduser -D -g '' user

FROM alpine:3.8

MAINTAINER Jojo le Barjos (jojolebarjos@gmail.com)

RUN apk add --no-cache fontconfig

COPY --from=builder /tmp/root /

COPY --from=builder /etc/passwd /etc/passwd

EXPOSE 8080/tcp

USER user

ENTRYPOINT ["/bin/server"]
