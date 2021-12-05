FROM alpine:3.14.1 as xmrig-builder
RUN apk add git make cmake libstdc++ gcc g++ automake libtool autoconf linux-headers
ARG XMRIG_VERSION=6.16.2-mo2
RUN git clone https://github.com/moneroocean/xmrig.git --branch v${XMRIG_VERSION}
RUN mkdir xmrig/build \
	&& cd xmrig/scripts \
    && sed -i 's/DonateLevel=1/DonateLevel=0/g' /xmrig/src/donate.h \
	&& ./build_deps.sh \
	&& cd ../build \
	&& cmake .. -DXMRIG_DEPS=scripts/deps \
      -DBUILD_STATIC=ON \
      -DCMAKE_BUILD_TYPE=Release \
      -DWITH_OPENCL=OFF \
      -DWITH_CUDA=OFF \
	&& make -j$(nproc)

FROM golang:alpine3.14 as server-builder

WORKDIR /app
COPY server/*.go ./
RUN go build -ldflags="-s -w"  -o /server main.go

FROM alpine:3.14.1 as compressor
RUN apk add upx
WORKDIR /staging
COPY --from=server-builder /server ./
COPY --from=xmrig-builder /xmrig/build/xmrig ./
RUN upx ./server
RUN upx ./xmrig

FROM alpine:3.14.1 as xmrig-workers-builder

RUN apk add git npm
RUN git clone https://github.com/xmrig/xmrig-workers
WORKDIR /xmrig-workers
RUN sed -i 's/createBrowser/createHash/g' ./src/store/history.js
RUN npm install
RUN rm /xmrig-workers/public/assets/js/app* /xmrig-workers/public/assets/css/app*
RUN npm run build

FROM alpine:3.14.1
ENV PGID=1000 \
    PUID=1000 \
    XMRIG_API_ENABLED=true \
    XMRIG_WORKERS_AUTOCONFIGURE=true \
    XMRIG_WORKERS_ENABLED=true

RUN mkdir /xmrig-workers
COPY --from=compressor /staging/server /xmrig-workers/server
COPY --from=xmrig-workers-builder /xmrig-workers/public /xmrig-workers/www
COPY --from=compressor /staging/xmrig /usr/local/bin/xmrig
COPY rootfs /

ENTRYPOINT ["/entrypoint.sh"]
