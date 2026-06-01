FROM golang:1.24.4 AS builder
COPY . /src
WORKDIR /src
RUN go mod download && \
    go mod verify && \
    go build -ldflags '-linkmode external -extldflags "-fno-PIC -static"' -v -o /usr/bin/amneziawg-go .

FROM alpine:3.19
ARG AWGTOOLS_RELEASE="1.0.20250901"

RUN apk --no-cache add iproute2 iptables bash unzip wget && \
    wget -q -O /tmp/awgtools.zip \
        "https://github.com/amnezia-vpn/amneziawg-tools/releases/download/v${AWGTOOLS_RELEASE}/alpine-3.19-amneziawg-tools.zip" && \
    unzip -j /tmp/awgtools.zip -d /usr/bin/ && \
    rm /tmp/awgtools.zip && \
    chmod +x /usr/bin/awg /usr/bin/awg-quick && \
    ln -sf /usr/bin/awg /usr/bin/wg && \
    ln -sf /usr/bin/awg-quick /usr/bin/wg-quick && \
    mkdir -p /etc/amneziawg

COPY --from=builder /usr/bin/amneziawg-go /usr/bin/amneziawg-go
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# AWG_CONFIG — path to config file inside container (default: /etc/amneziawg/awg0.conf)
ENV AWG_CONFIG=/etc/amneziawg/awg0.conf

ENTRYPOINT ["/entrypoint.sh"]
