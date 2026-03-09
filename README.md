# Alt DNSRadar

[Русская версия](./README.ru.md)

Alt DNSRadar is a lightweight cross-platform Go CLI that discovers alternative IP addresses for a domain using EDNS Client Subnet (ECS) probing and ranks reachable results by real TCP connect latency. Thanks to Ori from ntc.party for the ECS scanning idea.

The tool helps reveal IP addresses that may not be visible through the local DNS resolver.

Many modern platforms return different DNS answers depending on the geographic location of the client. A typical local DNS query only reveals a small subset of available endpoints. By simulating DNS requests from many different client networks using ECS, DNSRadar can discover additional IP addresses that are normally hidden from the local resolver.

This makes the tool useful for understanding how a platform distributes its infrastructure, how different resolvers see the same domain, and which reachable endpoints respond fastest from the current network.

## Features

- DNS diagnostics through:
  - Local DNS
  - Google DNS over UDP
  - Google DNS over HTTPS
  - Cloudflare DNS over HTTPS
- ECS scanning across 540 public subnets
- Discovery of additional IP addresses returned by DNS when queried from different client network locations
- TCP latency measurement with:
  - 3 probes per IP
  - median latency result
- TLS handshake diagnostics separate from ranking
- Geo / ASN metadata enrichment for top results via `ipinfo.io`
- Cross-platform support:
  - Linux
  - macOS
  - Windows
- Minimal dependency footprint

## How It Works

1. The program runs DNS diagnostics and compares answers from Local DNS, Google UDP, Google DoH, and Cloudflare DoH.
2. It performs an ECS scan across 540 public subnets and aggregates all unique IPs returned by DNS.
3. It measures TCP latency to each discovered IP with 3 probes and uses the median as the final ranking value.
4. It enriches the fastest endpoints with metadata from `ipinfo.io`.
5. It prints the fastest reachable endpoints in a compact table.

## Design Constraints

- TCP connect latency is the only ranking metric.
- TLS is diagnostic only and is never used for ranking.
- Geo metadata is fetched only for the top results.
- The tool is designed to stay lightweight and portable.

## Download Prebuilt Packages

Prebuilt packages for the main platforms will appear on the GitHub Releases page.

Release asset links:

- Linux amd64: `https://github.com/SoloIl/alt-dnsradar/releases/latest/download/alt-dnsradar-linux-amd64.tar.gz`
- Linux arm64: `https://github.com/SoloIl/alt-dnsradar/releases/latest/download/alt-dnsradar-linux-arm64.tar.gz`
- macOS amd64: `https://github.com/SoloIl/alt-dnsradar/releases/latest/download/alt-dnsradar-darwin-amd64.tar.gz`
- macOS arm64: `https://github.com/SoloIl/alt-dnsradar/releases/latest/download/alt-dnsradar-darwin-arm64.tar.gz`
- Windows amd64: `https://github.com/SoloIl/alt-dnsradar/releases/latest/download/alt-dnsradar-windows-amd64.zip`
- Windows arm64: `https://github.com/SoloIl/alt-dnsradar/releases/latest/download/alt-dnsradar-windows-arm64.zip`

## Install From Source

### Prerequisites

You need Go installed on your system.

Check that Go is available:

```bash
go version
```

### Linux

1. Install Go from your distribution package manager or from the official Go website.
2. Clone the repository:

```bash
git clone https://github.com/SoloIl/alt-dnsradar.git
cd alt-dnsradar
```

3. Build the binary:

```bash
go build -o alt-dnsradar .
```

4. Run it:

```bash
./alt-dnsradar example.com
```

### macOS

1. Install Go from the official package or with Homebrew:

```bash
brew install go
```

2. Clone the repository:

```bash
git clone https://github.com/SoloIl/alt-dnsradar.git
cd alt-dnsradar
```

3. Build the binary:

```bash
go build -o alt-dnsradar .
```

4. Run it:

```bash
./alt-dnsradar example.com
```

### Windows

1. Install Go from the official Windows installer.
2. Open PowerShell or Command Prompt.
3. Clone the repository:

```powershell
git clone https://github.com/SoloIl/alt-dnsradar.git
cd alt-dnsradar
```

4. Build the binary:

```powershell
go build -o alt-dnsradar.exe .
```

5. Run it:

```powershell
.\alt-dnsradar.exe example.com
```

## Quick Start

If you already downloaded a prebuilt package:

### Linux / macOS

1. Unpack the archive.
2. Open a terminal.
3. Either change into that folder and run:

```bash
./alt-dnsradar example.com
```

4. Or drag the `alt-dnsradar` file into the terminal window to insert its full path, then add a domain, for example:

```bash
/full/path/to/alt-dnsradar example.com
```

### Windows

1. Unpack the archive.
2. Open PowerShell.
3. Either change into that folder and run:

```powershell
.\alt-dnsradar.exe example.com
```

4. Or drag `alt-dnsradar.exe` into the PowerShell window to insert its full path, then add a domain, for example:

```powershell
C:\path\to\alt-dnsradar.exe example.com
```

## Usage

### Show Help

```bash
go run . --help
```

### Russian UI

```bash
go run . example.com --lang ru
```

### Default Run

```bash
alt-dnsradar example.com
```

With a plain command like that, the tool will:

- run DNS diagnostics through Local DNS, Google UDP, Google DoH, and Cloudflare DoH
- run initial TCP/TLS diagnostics for the discovered initial endpoints
- use a 3-second TCP timeout and 3 TCP probes with median latency
- run ECS scanning across 540 public subnets
- measure latency with 20 worker threads
- run TLS diagnostics and fill `ipinfo.io` metadata for the top 5 fastest endpoints

### Examples

Show all discovered IPs:

```bash
alt-dnsradar example.com --all
```

Write a compact log file:

```bash
alt-dnsradar example.com -l dnsradar.log
```

Disable colors:

```bash
alt-dnsradar example.com --no-color
```

## Example Output

Illustrative output for `youtube.com`:

```text
Alt DNSRadar v0.12

Processing URL "youtube.com"

DNS diagnostics
----------------------------
Running initial TCP/TLS diagnostics for 5 unique endpoint(s)...

Initial endpoint diagnostics

SOURCE            IP               TCP       TLS      NOTE
--------------------------------------------------------------------------------
Local DNS         172.217.20.174   22ms      TIMEOUT  shared with DoH reference
Google UDP        172.217.20.174   22ms      TIMEOUT  shared with DoH reference
Google DoH        173.194.222.136  24ms      TIMEOUT  reference
Google DoH        173.194.222.190  20ms      TIMEOUT  reference
Google DoH        173.194.222.91   20ms      TIMEOUT  reference
Google DoH        173.194.222.93   22ms      TIMEOUT  reference
Cloudflare DoH    172.217.20.174   22ms      TIMEOUT  reference

DNS diagnostic summary
- Local DNS returned a different multi-endpoint set from Google DoH
- Google UDP and Google DoH returned different multi-endpoint sets; possible cache, CDN variance, or interception
- Cloudflare DoH and Google DoH returned different multi-endpoint sets; reference confidence is lower

-------------------------------------------

Starting ECS scan
Total ECS subnets: 540

ECS scan 540/540 [================================================| 100 %]

DNS successful replies: 540
Unique IP discovered: 398

TCP latency 398/398 [================================================| 100 %]

Preparing top endpoint table for youtube.com (geo lookup + TLS diagnostics)...

Top fastest endpoints for youtube.com

IP               TCP     TLS      CDN            ASN      LOCATION
--------------------------------------------------------------------------------
142.251.142.238  22ms    TIMEOUT  Google         AS15169  SE   Stockholm
142.251.38.110   22ms    TIMEOUT  Google         AS15169  US   Mountain View
142.251.143.142  23ms    TIMEOUT  Google         AS15169  US   Mountain View
172.217.20.174   23ms    TIMEOUT  Google         AS15169  US   Mountain View
```

Actual output depends on the domain, network conditions, resolver behavior, and endpoint reachability.

## Notes

- `ipinfo.io` metadata requests are limited by the external service. The current workflow queries only the top 5 fastest endpoints to reduce usage (1000 requests per day).
- DNS answers for multi-endpoint domains may differ between resolvers without implying manipulation.
- Some networks or resolvers may block ECS behavior or return limited results.
- For reliable results, disable software that changes end-user traffic: proxies, VPNs, and anti-DPI tools.

## Limitations

- The ECS scan uses a coarse grid and may not discover all possible endpoints.
- Metadata accuracy depends on the `ipinfo.io` database.
- TLS diagnostics may be affected by middleboxes or DPI systems and should be interpreted as diagnostics, not ranking input.

## Testing

Run the unit tests:

```bash
go test ./...
```

## License

MIT
