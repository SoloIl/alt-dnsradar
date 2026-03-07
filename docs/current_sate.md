Alt DNSRadar — AI Project Context

Project

Alt DNSRadar

Command-line tool for discovering CDN edge servers using ECS-based DNS probing and ranking them by real TCP latency.

Primary goals:
	•	discover hidden CDN edges
	•	detect DNS manipulation
	•	measure network reachability
	•	identify fastest CDN edges for a given client location

The tool is intentionally lightweight, cross-platform, and has minimal dependencies.

Current State

Version: Alt DNSRadar v0.12

Stable working pipeline.

Example run:

go run . instagram.com

Execution Pipeline

domain input
↓
DNS diagnostics
↓
ECS global scan
↓
unique IP aggregation
↓
parallel edge latency measurement
↓
median ranking
↓
geo enrichment
↓
table output

Step 1 — DNS Diagnostics

Purpose: detect DNS manipulation before performing ECS scan.

Resolvers checked:

Local DNS
Google DNS UDP (8.8.8.8)
Google DNS DoH

Comparison allows detection of DNS poisoning and DNS interception.

Additionally DoH resolved IP is tested for TCP connectivity.

Example:

BLOCKED 157.240.205.174

This may indicate DPI filtering or network blocking.

Step 2 — ECS Global Scan

The core discovery mechanism.

DNSRadar sends DNS queries with EDNS Client Subnet (ECS) values representing different geographic client prefixes.

Subnet grid used:

0..255 step 10
0..255 step 10

Total ECS requests:

540 subnets

Each ECS query may return different CDN edges depending on simulated client location.

Example result:

DNS successful replies: 540
Unique IP discovered: 68

These IPs represent candidate CDN edge servers.

Step 3 — Unique IP Aggregation

All DNS replies are merged and duplicates removed.

Purpose: build a list of unique CDN edge IPs discovered via ECS probing.

Example:

Unique IP discovered: 68

Step 4 — Edge Latency Measurement

Latency is measured for every discovered IP.

Important design decision: latency is NOT measured per POP cluster.

Instead the system measures latency directly to the discovered IPs.

This ensures that the fastest edges returned are real reachable endpoints.

Latency measurement method:

3 TCP connection attempts
port 443
no payload

Only TCP connect time is measured.

TLS handshake is intentionally excluded from ranking.

Reason: DPI systems frequently interfere with TLS handshake which can freeze connections or produce misleading timing results.

Therefore TCP connect latency is considered a more reliable measurement of reachability and routing performance.

Median Latency

To stabilize measurements:

3 probes are executed and median latency is used.

Example:

probe1 = 48ms
probe2 = 46ms
probe3 = 120ms

Median result:

48ms

Median avoids distortion from packet jitter, routing spikes, and temporary congestion.

Parallel Latency Workers

Latency tests are executed in parallel using a worker pool.

Typical configuration:

workers ≈ 20
probes = 3

Progress bar example:

Edge latency 68/68 [===============================| 100 %]

Parallelization significantly reduces scan time.

Step 5 — Geo Enrichment

Fastest edges are enriched using the ipinfo.io API.

Fields extracted:

city
country
ASN
organization

CDN detection is derived from the organization field.

Example mapping:

Facebook → Meta
Google → Google
Cloudflare → Cloudflare
Amazon → Amazon
Microsoft → Microsoft

Only the fastest edges are queried to reduce API usage.

Typical scan requires about 5 API requests.

Example Output

Top fastest edges

IP               TCP    CITY               COUNTRY  ASN       CDN
157.240.200.174  42ms   Ballerup           DK       AS32934   Meta
57.144.244.34    47ms   Frankfurt am Main  DE       AS32934   Meta
157.240.253.174  48ms   Frankfurt am Main  DE       AS32934   Meta
157.240.17.174   50ms   Zürich             CH       AS32934   Meta
57.144.222.34    50ms   Amsterdam          NL       AS32934   Meta

Project Structure

dnsradar/

main.go

config.go
flags.go
logging.go

dns.go
ecs.go

latency.go

geo.go

clustering.go
results.go

utils.go

netfilter.go

Module Responsibilities

main.go
Program entry point controlling the full pipeline: diagnostics → ECS scan → latency measurement → ranking → output.

dns.go
Implements DNS diagnostics including local resolver checks, Google UDP DNS queries, Google DoH queries, and DoH IP connectivity testing.

ecs.go
Core ECS scanning engine. Generates ECS subnets, performs parallel DNS queries, displays progress bars, and aggregates DNS responses.

latency.go
Handles latency measurement using worker pools. Executes three TCP probes per IP, calculates median latency, and reports progress.

geo.go
Fetches metadata from ipinfo API including city, country, ASN, and organization. Converts organization values into simplified CDN names.

clustering.go
Clusters IPs by /24 network. Currently not used in ranking but kept for future features such as POP mapping or topology analysis.

results.go
Sorting and output utilities including ranking and formatted table output.

utils.go
Shared utilities such as URL normalization, duplicate removal, and common helper functions.

netfilter.go
Filters non-public IPv4 ranges such as RFC1918, loopback, link-local, multicast, and CGNAT ranges. Used during ECS subnet generation.

Key Design Decisions

ECS instead of brute DNS

ECS reveals CDN edges associated with different client geographies.

TCP latency instead of TLS latency

TLS handshake is frequently interfered with by DPI systems, so TCP connect latency provides more reliable measurements.

Median latency

Median of three probes improves measurement stability and reduces the effect of network jitter.

Parallel scanning

Worker pools allow efficient latency measurement across large IP sets.

Limited external API usage

Geo lookup is performed only for fastest edges to minimize API requests.

Coding Guidelines for AI Assistants

When modifying this project:

Do not remove ECS scanning.
Do not replace median latency with average latency.
Do not move latency measurement to POP clusters.
Do not introduce TLS handshake into latency ranking.

Maintain concurrency model.
Keep CLI simple.
Maintain deterministic output format.
Keep external dependencies minimal.

Future Extension Points

Possible future features:

POP topology mapping
automatic CDN detection improvements
client_subnet tuning for DNS resolvers
edge distribution analytics

These may reuse clustering logic, ECS scan datasets, and latency metrics.

Architecture Anchor: DNS And TLS Diagnostics

This section records the currently agreed redesign direction for future implementation.

Status:

Discussed and agreed at architecture level.
Not yet implemented in code.

Core agreement:

TCP connect latency remains the only ranking metric.

TLS handshake must never be used for edge ranking.

TLS handshake may be used only as a diagnostic signal.

DNS Comparison Model

DNS comparison must no longer use a simple boolean overlap check.

The comparison model should use 4 states:

unavailable
exact_match
partial_overlap
no_overlap

Interpretation rules:

Google UDP vs Google DoH:
used to detect possible interception or spoofing of unencrypted DNS traffic to Google.

Local DNS vs Google DoH:
used to compare local resolver behavior against encrypted reference results.

partial_overlap:
must be treated as an ambiguous result and may reflect cache, TTL drift, CDN rotation, or geo variance.

no_overlap:
stronger signal of mismatch and possible manipulation, but still should be phrased carefully.

unavailable:
must not be reported as a security event. It means there was not enough data for comparison.

TLS Diagnostic Scope

TLS diagnostics should be used only in 2 places:

initial resolved IPs
top 5 fastest edges

The purpose is diagnostic only:

TCP reachability shows whether the IP is reachable.

TLS diagnostic shows whether the next practical protocol step is blocked, stalled, or interfered with.

TLS Diagnostic Timeout

TLS diagnostic should use a short timeout window of about 2-3 seconds.

Reason:

Full waiting time may be too long in DPI-interfered networks.
Short timeout is an intentional compromise for practical diagnostics.

TLS Diagnostic Status Model

Preferred TLS status values:

OK
FAIL
TIMEOUT
SKIP

Definitions:

OK:
TLS handshake completed.

FAIL:
TLS handshake returned an error before timeout.

TIMEOUT:
TLS handshake did not complete within the short diagnostic timeout.

SKIP:
TLS diagnostic was not executed for this IP.

UI Direction

The UI direction agreed for future implementation is:

1. Initial DNS/TCP/TLS diagnostics should be shown in a compact table.
2. A short DNS summary should be printed after that table.
3. The final ranking table should include TLS status as an additional column.
4. Sorting must remain based only on TCP latency.

Table intent:

Initial diagnostics table:
should combine DNS-derived initial IPs with TCP latency and TLS status for faster operator reading.

Final ranking table:
should remain ranking-focused, but include TLS status as an additional diagnostic column.

Constraints That Remain Unchanged

Do not remove ECS scanning.

Do not replace direct IP latency measurement with POP-based ranking.

Do not replace median latency with average latency.

Do not use TLS timing in ranking.

Keep the project lightweight, modular, and cross-platform.
