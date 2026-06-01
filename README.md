# Go Implementation of AmneziaWG

AmneziaWG is a contemporary version of the WireGuard protocol. It's a fork of WireGuard-Go and offers protection against detection by Deep Packet Inspection (DPI) systems. At the same time, it retains the simplified architecture and high performance of the original.

The precursor, WireGuard, is known for its efficiency but had issues with detection due to its distinctive packet signatures.
AmneziaWG addresses this problem by employing advanced obfuscation methods, allowing its traffic to blend seamlessly with regular internet traffic.
As a result, AmneziaWG maintains high performance while adding an extra layer of stealth, making it a superb choice for those seeking a fast and discreet VPN connection.
## This Fork

Extends the original AmneziaWG with additional DPI evasion techniques:

- **TCP decoy server** — listens on the same port as UDP, responds as nginx to active probing (TSPU/GFW send TCP connections to suspected VPN ports to fingerprint the server type)
- **QUIC/DTLS mimicry** — junk packets mimic QUIC Initial and DTLS ClientHello headers for larger sizes
- **Keepalive jitter** — ±2 s random jitter on the 10 s keepalive interval to break fixed-interval ML signatures; ±500 ms on persistent keepalive
- **Adaptive obfuscation** — junk packet rate escalates from 10 % to 30 % after repeated handshake timeouts (DPI interference suspected)
- **Anti-probing** — drops raw WireGuard message types 1–4 when custom headers are configured, so the server is invisible to passive scanners
- **S1–S3 collision warnings** — logs an error at startup if any padded packet size recreates a vanilla WireGuard signature or two message types produce the same size
- **Bimodal junk size distribution** — junk packets use a bimodal distribution (small / medium / large) instead of uniform random, better matching real HTTPS traffic patterns

## Docker Deployment

Build the image from this repository:

```bash
docker build -t my-amneziawg .
```

Run the container:

```bash
docker run -d \
  --name amneziawg \
  --restart unless-stopped \
  --cap-add NET_ADMIN \
  --device /dev/net/tun \
  --sysctl net.ipv4.ip_forward=1 \
  --sysctl net.ipv6.conf.all.forwarding=1 \
  -v /etc/amneziawg/awg0.conf:/etc/amneziawg/awg0.conf:ro \
  -p 41442:41442/udp \
  my-amneziawg
```

The container brings the interface up via `awg-quick` on start and tears it down cleanly on stop.

**Environment variables:**

| Variable | Default | Description |
|---|---|---|
| `AWG_CONFIG` | `/etc/amneziawg/awg0.conf` | Path to the config file inside the container |

> [!NOTE]
> The container requires `NET_ADMIN` capability and access to `/dev/net/tun`. The kernel module (`amneziawg-linux-kernel-module`) is not needed — the Go userspace implementation is used instead.

## Usage

Simply run:

```
$ amneziawg-go wg0
```

This will create an interface and fork into the background. To remove the interface, use the usual `ip link del wg0`, or if your system does not support removing interfaces directly, you may instead remove the control socket via `rm -f /var/run/amneziawg/wg0.sock`, which will result in amneziawg-go shutting down.

To run amneziawg-go without forking to the background, pass `-f` or `--foreground`:

```
$ amneziawg-go -f wg0
```
When an interface is running, you may use [`amneziawg-tools `](https://github.com/amnezia-vpn/amneziawg-tools) to configure it, as well as the usual `ip(8)` and `ifconfig(8)` commands.

To run with more logging you may set the environment variable `LOG_LEVEL=debug`.

## Platforms

### Linux

This will run on Linux; you should run amnezia-wg instead of using default linux kernel module.

### macOS

This runs on macOS using the utun driver. It does not yet support sticky sockets, and won't support fwmarks because of Darwin limitations. Since the utun driver cannot have arbitrary interface names, you must either use `utun[0-9]+` for an explicit interface name or `utun` to have the kernel select one for you. If you choose `utun` as the interface name, and the environment variable `WG_TUN_NAME_FILE` is defined, then the actual name of the interface chosen by the kernel is written to the file specified by that variable.
This runs on MacOS, you should use it from [amneziawg-apple](https://github.com/amnezia-vpn/amneziawg-apple)

### Windows

This runs on Windows, you should use it from [amneziawg-windows](https://github.com/amnezia-vpn/amneziawg-windows), which uses this as a module.


## Building

This requires an installation of the latest version of [Go](https://go.dev/).

```
$ git clone https://github.com/amnezia-vpn/amneziawg-go
$ cd amneziawg-go
$ make
```

## Configuration

> [!NOTE]
> If there is no value specified (for any param), AWG treats it as 0

### Junk packets

The amount of junk packets specified in `Jc` with a random size between `Jmin` and `Jmax` would be generated and sent prior every handshake

- `Jc: int`, recommended range is 4-12
- `Jmin: int` <= `Jmax:int`

> [!TIP]
> Junk packets do not carry any actual data, so there is no need to specify it on both sides. General recommendation is to use it on the client side only

> [!IMPORTANT]
> If Jmax >= system MTU (not the one specified in AWG), then the system can fracture this packet into fragments, which looks suspicious from the censor side

### Message paddings

- `S1: int` - padding of handshake initial message
- `S2: int` - padding of handshake response message
- `S3: int` - padding of handshake cookie message
- `S4: int` - padding of transport messages

### Message headers

Every message in wireguard has `uint32` type at the beginning of the packet. This field could be controlled by specifying the params below:

- `H1: string` - header range of handshake initial message
- `H2: string` - header range of handshake initial message
- `H3: string` - header range of handshake cookie message
- `H4: string` - header range of transport message

Values could be specified as:
- range: `x-y`, x <= y; e.g. `123-456`
- single value `1234`

### Custom signature packets

These packets are being send prior to every handshake, in the same way as Junk packets do. The sending order is `I1`, `I2`, `I3`, `I4`, `I5`. If there is no value specified, the packet is skipped.

- `I1: string`
- `I2: string`
- `I3: string`
- `I4: string`
- `I5: string`

Value is a sequence of tags specified below:
- `<b 0x[seq]>` - static bytes tag. Dumps `[seq]` as-is to the packet. `[seq]` is hex-encoded sequence which represents bytes sequence (2 hex numbers per byte) and is always even-sized
- `<r [size]>` - random bytes tag. Dumps `[size]` amount of randomly-generated bytes to the packet
- `<rd [size]>` - random digits tag. Dumps `[size]` amount of randomly-generated bytes from `[0-9]` set to the packet
- `<rc [size]>` - random chars tag. Dumps `[size]` amount of randomly-generated bytes from `[a-zA-Z] set to the packet
- `<t>` - timestamp tag. Dumps 4-bytes long current system time in UNIX format

> [!TIP]
> Custom signature packets does not carry any actual data, so there is no need to specify it on both sides. General recommendation is to use it on the client side only

> [!IMPORTANT]
> If the final size of any packet exceeds system MTU, it would be fractured into fragments, which looks suspicious