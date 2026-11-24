# Monit Exporter for Prometheus

> **Forked from** [commercetools/monit_exporter](https://github.com/commercetools/monit_exporter)

## English

### Introduction

Monit Exporter is a Prometheus Exporter 
that scrapes Monit status in XML format and exposes the metrics via an HTTP endpoint.

### Features

- **Enhanced Logging:**
    - Detailed logs provide insights into HTTP requests, metric collection processes, and potential issues.

- **Exposes Prometheus-Compatible Metrics:**
    - Seamlessly integrates with Prometheus for monitoring Monit-managed services.

- **Fully Configurable via Command-Line Flags:**
    - Customize exporter behavior and Monit scraping parameters as needed.

### Installation

1. **Install [Go](https://golang.org/dl/) 1.27.4 or later.**

2. **Clone the repository and build:**

    ```bash
    git clone https://github.com/ririnto/monit-exporter.git
    cd monit-exporter
    go build -o monit-exporter
    ```

### Usage

#### Commands

- **serve**: Starts the Monit Exporter server.

#### Flags

Below is an overview of the flags defined in `cmd/root.go`:

| Flag               | Default                                               | Description                                                             |
|--------------------|-------------------------------------------------------|-------------------------------------------------------------------------|
| `listen-address`   | `localhost:9388`                                      | The address on which the exporter will listen (e.g., '0.0.0.0:9388').   |
| `metrics-path`     | `/metrics`                                            | The HTTP path at which metrics are served (e.g., '/metrics').           |
| `ignore-ssl`       | `false`                                               | Whether to skip SSL certificate verification for Monit endpoints.       |
| `monit-scrape-uri` | `http://localhost:2812/_status?format=xml&level=full` | The Monit status URL to scrape (XML format).                            |
| `monit-user`       | *(empty)*                                             | Basic auth username for accessing Monit.                                |
| `monit-password`   | *(empty)*                                             | Basic auth password for accessing Monit.                                |
| `log-level`        | `info`                                                | Log level for the application (debug, info, warn, error, fatal, panic). |

**Launch the exporter with desired flags:**

```bash
./monit-exporter serve \
  --listen-address="0.0.0.0:9388" \
  --monit-scrape-uri="http://localhost:2812/_status?format=xml&level=full" \
  --monit-user="admin" \
  --monit-password="monitpassword" \
  --log-level="info"
```

**Visit the metrics endpoint:**

```bash
curl http://localhost:9388/metrics
```

#### Process resource metrics

These gauges use `check_name` for the Monit process service name. CPU and memory gauges also use `type` to identify the processes included.

| Metric | `type` | Value |
| --- | --- | --- |
| `monit_service_cpu_perc` | `percentage` | CPU usage by the process, as a percentage |
| `monit_service_cpu_perc` | `percentage_total` | CPU usage by the process and its children, as a percentage |
| `monit_service_mem_bytes` | `process` | Memory used by the process, in bytes |
| `monit_service_mem_bytes` | `process_and_children` | Memory used by the process and its children, in bytes |
| `monit_service_write_bytes_per_second` | | Disk write rate in bytes per second |
| `monit_service_write_cumulative_bytes` | | Cumulative disk bytes written as reported by Monit |

The exporter converts Monit's memory values to bytes and emits each series only when Monit reports its value.

#### Process restarts

`monit_service_process_pid` is a gauge labeled by `service_name`.
The exporter emits it only when Monit reports a valid PID for a process service.
Use `changes(monit_service_process_pid[1h])` to count observed PID changes per series over the past hour as an estimate of restarts.
Gaps between Monit polls or Prometheus scrapes can hide restarts, and the operating system can reuse a PID.

#### Troubleshooting Monit scrapes

If Monit returns HTTP 401, check that `--monit-user` and `--monit-password` match the credentials configured for Monit's HTTP server.
Older exporter releases could report Monit's HTML error page as an XML syntax error.

Use `/_status?format=xml&level=full` in `--monit-scrape-uri`.
Without `format=xml`, Monit serves the status as plain text.

### Metrics inventory

The exporter reads Monit's `/_status?format=xml&level=full` response.
The wire reference is Monit's release source, `src/http/xml.c` (`document_head`, `status_service`, and `status_xml`) and `src/http/cervlet.c` (`print_status`).
The five source archives in the fixture matrix had no standalone XML schema. Their serializers define the element names, nesting, and optional fields.
The [Monit 6.1.0 manual][inventory-monit-manual] documents what you can configure. The [source archives][inventory-monit-dist] and [release notes][inventory-monit-changes] establish what applies to each version, as checked on 2026-11-24.

`/_status` uses Monit's status format 1, with `monit/server/version` and `monit/service[@type]/name`.
`/_status2` uses format 2, with the Monit version on the root and services under `monit/services/service`, with `@name` and a `type` child.
The [parser](internal/monit/monit.go) reads the format 1 layout. A valid format 2 document can produce `monit_exporter_up 1` with no service series.
Both formats use XML declaration version `1.0`, which is independent of the Monit release and status format numbers.

#### Exported metrics

All `monit_*` families below are gauges. In the table, **service labels** means `service_name`, `service_type`, and `service_monitor_status`.
Both `service_name` and `check_name` carry the XML service name. `service_type` uses the names in the coverage table or `unknown` for an unrecognized type code.
`service_monitor_status` is the decimal string from `<monitor>`. Monit defines `0` as unmonitored, `1` as monitored, `2` as initializing, and `4` as the waiting flag in `src/monit.h`.
The collector clears service series before each scrape. A failed fetch or parse leaves only `monit_exporter_up` among the Monit families. [Metric definitions and collection](internal/exporter/exporter.go)

| Metric | Labels | Value |
| --- | --- | --- |
| `monit_exporter_up` | None | `1` after a successful fetch and XML decode, `0` on fetch or decode failure. |
| `monit_exporter_service_check` | Service labels | Raw `<status>` event-class bitmap. `0` means no bits are set. A missing element decodes as `0`. |
| `monit_service_process_pid` | `service_name` | Positive signed 32-bit decimal `<pid>` for `Process`. Missing and invalid PIDs are omitted. |
| `monit_service_cpu_perc` | `check_name`, `type` | `cpu/percent` for `percentage`, `cpu/percenttotal` for `percentage_total`. Both use Monit's percentage scale. |
| `monit_service_mem_bytes` | `check_name`, `type` | `memory/kilobyte` for `process`, `memory/kilobytetotal` for `process_and_children`. Both are multiplied by 1024. |
| `monit_service_write_bytes_per_second` | `check_name` | `write/bytes/count`, in bytes per second. |
| `monit_service_write_cumulative_bytes` | `check_name` | `write/bytes/total`, in bytes, exposed as a gauge. |
| `monit_service_block_usage_bytes` | Service labels | Raw `block/usage`. Monit emits MiB, and the exporter forwards the number unchanged. |
| `monit_service_block_total_bytes` | Service labels | Raw `block/total`, also in MiB. |
| `monit_service_block_usage_percent` | Service labels | `block/percent`, on Monit's percentage scale. |
| `monit_service_inode_usage` | Service labels | `inode/usage`, the used inode count. |
| `monit_service_inode_total` | Service labels | `inode/total`, the total inode count. |
| `monit_service_inode_usage_percent` | Service labels | `inode/percent`, on Monit's percentage scale. |
| `monit_service_port_response_seconds` | Service labels | `port/responsetime` in seconds, including Monit's `-1` if a check is unavailable. The last `<port>` supplies the response time when several are present. |
| `monit_service_system_loadavg_01` | Service labels | `system/load/avg01`, the one-minute load average. |
| `monit_service_system_loadavg_05` | Service labels | `system/load/avg05`, the five-minute load average. |
| `monit_service_system_loadavg_15` | Service labels | `system/load/avg15`, the fifteen-minute load average. |
| `monit_service_system_cpu_user_percent` | Service labels | `system/cpu/user` percentage. |
| `monit_service_system_cpu_system_percent` | Service labels | `system/cpu/system` percentage. |
| `monit_service_system_cpu_wait_percent` | Service labels | `system/cpu/wait` percentage. |
| `monit_service_system_memory_usage_percent` | Service labels | `system/memory/percent`. |
| `monit_service_system_memory_usage_kilobytes` | Service labels | `system/memory/kilobyte`, in 1024-byte units. |
| `monit_service_system_swap_usage_percent` | Service labels | `system/swap/percent`. |
| `monit_service_system_swap_usage_kilobytes` | Service labels | `system/swap/kilobyte`, in 1024-byte units. |

The four process resource families require `Process` and emit only finite, nonnegative values, including zero. The total CPU and memory series include descendants.
For Linux and macOS, the reviewed 5.26.0 sources use physical I/O for process `write/bytes`. The later serializers put physical I/O in `write/bytes` and generic I/O in `write/bytesgeneric`.
Block, inode, port, and system gauges depend on their enclosing XML element. With an enclosing element present, missing numeric children emit zero.
Monit 6.1.0 computes the legacy `<status>` and `<status_hint>` bitmaps from detailed events. See `src/event.h` and `src/event.c` in the [6.1.0 source][inventory-monit-610] for the bit and event definitions.

#### Service coverage

Every decoded service has the common status gauge. Monit's `Service_Type` enum in `src/monit.h` ends at `Service_Net` across the reviewed releases.
The fields below describe the 6.1.0 status XML. Earlier versions and individual platforms can omit fields.

| Code and `service_type` | Extra metrics today | Other status XML fields |
| --- | --- | --- |
| `0` `Filesystem` | Block and inode gauges | `fstype`, `fsflags`, `mode`, `uid`, `gid`, `read` and `write` bytes and operations, and `servicetime` read, write, wait, and run. |
| `1` `Directory` | None | Mode, UID, GID, hard-link count, and access, change, and modification timestamps. |
| `2` `File` | None | Directory-style metadata, size in bytes, configured checksum type and digest. Content-match outcomes are events. |
| `3` `Process` | PID, CPU, memory, disk write, and TCP or UDP port response | Parent PID, UID, EUID, GID, uptime, threads, children, memory percentages, and file descriptor counts and limits. `read` and `write` can report physical bytes, generic bytes, and operation rates and totals. Unix socket results also apply. |
| `4` `Remote host` | TCP or UDP port response | Repeated ports with endpoint, request, protocol, transport, response time, and optional certificate validity in days. Repeated ICMP and Unix socket checks report response times. |
| `5` `System` | Load, user/system/wait CPU, memory, and swap | Uptime, boot time, file descriptors, extra CPU modes (`nice`, `hardirq`, `softirq`, `steal`, `guest`, `guestnice`), and swap `paging/in` and `paging/out` rates. |
| `6` `Fifo` | None | Mode, UID, GID, hard-link count, and access, change, and modification timestamps. |
| `7` `Program` | None | `program/started` and `program/status` hold the start timestamp and exit code once available. `program/output` is text. |
| `8` `Network` | None | `link/state`, `link/speed`, `link/duplex`, plus upload and download packet, byte, and error rates (`now`) and totals (`total`). `src/net/Link.c` defines which fields use `-1` for unavailable values. |

Custom scripts use `CHECK PROGRAM`. Configured tests such as content matching, protocol checks, and send/expect attach to an existing service type.
Monit reports their state through the service status and any applicable program or connection fields. Program output has no structured numeric schema.
All service types can also carry collection time (`collected_sec` and `collected_usec`), status hints, monitor mode, reboot policy, pending action, an `every` schedule when configured, and detailed `errors/event` entries in 6.1.0.
The status document also carries daemon version, uptime, poll interval and start delay, HTTP interface settings, and platform identity, CPU count, and memory and swap capacity. Format 2 adds service groups.
Top-level event and M/Monit host group elements belong to M/Monit reports. The HTTP status handler supplies neither.

#### Fixture versions

The [parser and fetch tests](internal/monit/monit_test.go), [collector tests](internal/exporter/exporter_test.go), and [CLI tests](cmd/root_test.go) use inline XML labeled 5.26.0.
The process resource test puts the release on the root while keeping format 1 service elements. Its mixed layout is synthetic.
The 5.25.3 case is a plain-text rejection test. The historical 5.25.3 and 5.26.0 authentication reproductions above are separate from the checked-in fixtures.
The current direct metric assertions cover reachability, common status, PID, and process resources. Filesystem detail, port response, and system detail lack direct assertions.

| Monit release | Proposed fixture role | Coverage to capture |
| --- | --- | --- |
| [5.25.3][inventory-monit-525] | Keep the existing negative case | Plain-text rejection. The XML serializer is identical to 5.26.0, so use 5.26.0 for the legacy positive fixture. |
| [5.26.0][inventory-monit-526] | Legacy format 1 baseline | Capture an ISO-8859-1 response with a non-ASCII service name, legacy status bits, PID, CPU, memory, and process I/O. |
| [5.35.2][inventory-monit-535] | Last published 5.x baseline | Capture additional system CPU modes, system and process file descriptors, hard-link counts, and the process `bytesgeneric` and physical `bytes` distinction. |
| [6.0.0][inventory-monit-600] | 6.x transition case | Capture system paging and process memory on Linux. The release prefers proportional set size on Linux and uses physical footprint on macOS, so assert conversion of the captured input. |
| [6.1.0][inventory-monit-610] | Current 6.x baseline | Capture UTF-8 format 1 for all nine types, detailed events beside legacy status bits, repeated probes, and absent optional fields. Capture format 2 separately for explicit compatibility coverage. |

The matrix proposes regression targets. The source comparison does not establish runtime support for every release or platform.
Use versioned, sanitized responses from a controlled Monit instance. Record the release, platform, endpoint, encoding, and check configuration so each response can be regenerated.
Use focused synthetic cases for failure and boundary values. Exclude credentials, command output, request contents, and private host and path values from committed fixtures.

#### Conventions and compatibility

- `monit-exporter serve` uses CLI flags, listens on `localhost:9388` by default, and serves `/metrics` through `promhttp`.
  It fetches one Monit target on demand with Basic Auth, a five-second timeout, and TLS verification enabled by default.
  HTTP or XML failures set `monit_exporter_up` to `0` while the exporter can still answer the scrape. [CLI](cmd/root.go), [HTTP server](cmd/serve.go), [Monit client](internal/monit/monit.go), [Prometheus exporter guidance][inventory-prom-exporters]
- The registry also includes the client library's Go runtime metrics and, on non-Darwin platforms, process metrics for the exporter itself.
  `promhttp_metric_handler_requests_total` counts completed scrapes by `code` (HTTP status), and `promhttp_metric_handler_requests_in_flight` gauges current scrapes without labels. [HTTP server](cmd/serve.go), [Prometheus handler instrumentation][inventory-prom-handler]
- Synchronous collection and the upstream reachability gauge follow Prometheus exporter guidance. The collector serializes requests around mutable gauges and resets them each time. Prometheus recommends const metrics for custom collectors to handle concurrent scrapes and disappearing series. For new families, use base units, ratios for fractions, bounded labels, and `_total` for suitable counters. Review overlaps with node exporter for system and filesystem stats. [Prometheus exporter guidance][inventory-prom-exporters], [metric and label naming][inventory-prom-naming]
- The [Taskfile](Taskfile.yaml) provides build, test, and lint commands. [Pull request CI](.github/workflows/test.yaml) runs the Go suite and pinned `golangci-lint`. Tests use `httptest`, Prometheus exposition comparisons, and a built CLI subprocess. [Go test command][inventory-go-test], [Prometheus exposition formats][inventory-prom-formats]
- The [release workflow](.github/workflows/deploy.yaml) runs on `v*` tag pushes. It builds with `CGO_ENABLED=0` for Linux, macOS, and Windows on `amd64` and `arm64`, then uploads `monit-exporter-<goos>-<goarch>` with `.exe` on Windows. The tag filter and OS/architecture matrix define the trigger and artifact set. [GitHub Actions workflow syntax][inventory-actions]

#### Follow-up pull requests

Each item is a separate proposed pull request. For every new metric, include a versioned XML fixture and an exposition test that asserts the name, type, labels, value, and applicable missing-data behavior.
Keep established series compatible. Changes to names, labels, units, or target behavior need their own compatibility decision.

1. **Establish the status-format baseline.** Add representative format 1 fixtures and assert all nine service labels, a combined status bitmap, monitoring flag changes, and removal after a failed scrape. Use `/_status2` to reproduce the empty-service success case and make support or a diagnostic an explicit decision. Evidence is `Monit.Services`, `ParseMonitStatus`, and Monit's `print_status` and `status_xml`.
2. **Cover filesystem capacity and units.** Assert current block and inode values with `block` and optional `inode`, including absent children. Propose a correctly scaled capacity family with a migration plan for the byte-named legacy gauges. Evidence is `status_service` and `collectServiceMetrics`.
3. **Cover repeated port checks and TLS.** Keep the legacy response gauge and add per-probe response, availability, and certificate lifetime where present. Test repeated endpoints, `-1`, missing certificates, and checks with the same endpoint and protocol. Establish a stable identity without exporting raw requests. Evidence is `Port` in the parser and the Monit port loop in `status_service`.
4. **Add ICMP and Unix socket results.** Use the `icmp` and `unix` loops in `status_service`. Test multiple probes, no probe, and unavailable response times for host and process checks as applicable.
5. **Add network link metrics.** Use `link` in `status_service` and `src/net/Link.c` for state, speed, duplex, and directional bytes, packets, and errors. Test down or unknown links, unavailable speed, zero traffic, and totals above 32 bits.
6. **Complete process I/O.** Add read and operation metrics, and evaluate generic versus physical byte series using `_ioStatisticsProcess`. Test absent and zero values, reset behavior, and wide totals while retaining the existing physical write gauges.
7. **Add process lifecycle and limits.** Use the process branch in `status_service` for uptime, parent PID, thread and child counts, memory percentages, and descriptor counts and limits. Test process-engine-disabled and unavailable fields. Evaluate identity metadata (UID, EUID, GID) separately from workload gauges within that review.
8. **Add program results.** Use `program/started` and `program/status` to expose last-start time and exit code. Test never-run, successful, and nonzero exits. The raw output can contain arbitrary text and belongs in Monit diagnostics.
9. **Add filesystem I/O.** Use `_ioStatisticsFilesystem` and `servicetime` for read and write rates, totals, operations, and supported timings. Test an unavailable timing and a real zero. Treat filesystem type, flags, and ownership as metadata and limit any new label values.
10. **Add system detail.** Use the system branch and `system` block for uptime, boot time, file descriptors, extra CPU modes, and paging. Test missing platform fields and zero values across 5.x and 6.x. Preserve existing system gauges and check which new metrics add Monit-specific value.
11. **Add file, directory, and FIFO metadata.** Use the three respective branches for mode, ownership, hard links, timestamps, and file size. Test absent fields and newer hard-link fields. Use checksum and content event state for monitoring, with raw digests kept in Monit.
12. **Expose common service state.** Use `status_service` and `src/monit.h` for collection time, pending actions, monitor mode, reboot policy, and legacy hints. Test legacy bitmaps and changes across scrapes. Evaluate `every` as schedule metadata.
13. **Expose detailed event state.** Use the `errors/event` loop in `status_service` and ID and state definitions in `src/event.h` and `src/event.c`. Test older releases without the element, simultaneous events, event recovery, and unknown IDs.
14. **Evaluate daemon and platform metrics.** Use `document_head` for daemon uptime, poll interval, start delay, version, CPU count, and memory and swap capacity. Test capacity conversion and omit private metadata from labels. Service groups are configuration metadata. Event notifications, credentials, and raw program or request text are outside the scrape metric contract.

[inventory-monit-manual]: https://mmonit.com/monit/documentation/monit.html
[inventory-monit-changes]: https://mmonit.com/monit/changes/
[inventory-monit-dist]: https://mmonit.com/monit/dist/
[inventory-monit-525]: https://mmonit.com/monit/dist/monit-5.25.3.tar.gz
[inventory-monit-526]: https://mmonit.com/monit/dist/monit-5.26.0.tar.gz
[inventory-monit-535]: https://mmonit.com/monit/dist/monit-5.35.2.tar.gz
[inventory-monit-600]: https://mmonit.com/monit/dist/monit-6.0.0.tar.gz
[inventory-monit-610]: https://mmonit.com/monit/dist/monit-6.1.0.tar.gz
[inventory-prom-exporters]: https://prometheus.io/docs/instrumenting/writing_exporters/
[inventory-prom-naming]: https://prometheus.io/docs/practices/naming/
[inventory-prom-formats]: https://prometheus.io/docs/instrumenting/exposition_formats/
[inventory-prom-handler]: https://pkg.go.dev/github.com/prometheus/client_golang@v1.25.0/prometheus/promhttp#InstrumentMetricHandler
[inventory-go-test]: https://pkg.go.dev/cmd/go#hdr-Test_packages
[inventory-actions]: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax

### Upstream decisions

- On Monit 5.25.3 and 5.26.0, omitting Basic Auth credentials reproduced the reported XML syntax error.
  Monit returned an HTTP 401 HTML page, and the historical exporter tried to parse it as XML.
  The fork documented the configuration error and added tests for scrape diagnostics.
  The original report omitted the response and Monit version, so its cause remains unconfirmed.
  [Upstream issue #8][upstream-issue-8], [Fork PR #13][fork-pr-13]
- The fork added the `monit_service_process_pid` gauge for process services with a valid PID.
  Monit polling gaps, Prometheus scrape gaps, and PID reuse can hide restarts.
  [Upstream issue #9][upstream-issue-9], [Fork PR #14][fork-pr-14]
- The fork implemented per-process CPU, memory, and disk-write metrics.
  A series disappears when Monit omits its value or service, or when the scrape fails.
  The upstream listen default change, metric renames, and release tooling fell outside that scope.
  [Upstream PR #7][upstream-pr-7], [Fork PR #15][fork-pr-15]
- The fork declined the PID patch because it can emit zero for a missing PID and emit PID series for non-process services.
  The fork already uses `promhttp` and accepts only valid PIDs from process services.
  [Upstream PR #10][upstream-pr-10], [Fork PR #14][fork-pr-14]
- The fork declined the `config.toml` setup routine because the exporter reads configuration from command-line flags.
  Adding setup for an unused file would introduce a second configuration interface without a demonstrated need.
  [Upstream issue #12][upstream-issue-12]

### Running tests and lint

Run the pull-request checks locally.

```bash
go test ./...
go tool -modfile=lint.mod golangci-lint run --config .golangci.yaml ./...
```

### Project / Package Structure

```
.
├── cmd
│   ├── root.go       (Defines root command and flags)
│   └── serve.go      (Implements 'serve' command, server startup)
├── internal
│   ├── config
│   │   └── config.go (Holds the Config struct for the exporter)
│   ├── exporter
│   │   └── exporter.go (Implements the Prometheus Exporter logic)
│   └── monit
│       └── monit.go    (Fetches and parses Monit status data)
├── main.go             (Entrypoint: calls cmd.Execute())
├── README.md           (This file)
└── LICENSE             (MIT License)
```

### License

This project is licensed under the [MIT License](LICENSE).

## 한국어

### 소개

Monit Exporter는 Monit 상태 정보를 XML 형식으로 수집하고 이를 Prometheus 메트릭으로 변환하여 HTTP 엔드포인트로 노출하는 익스포터입니다.

### 기능

- **향상된 로깅:**
    - 자세한 로그를 통해 HTTP 요청, 메트릭 수집 과정 및 잠재적인 문제를 파악할 수 있습니다.

- **Prometheus 호환 메트릭 제공:**
    - Monit에서 관리하는 서비스를 Prometheus와 원활히 통합하여 모니터링할 수 있습니다.

- **커맨드라인 플래그로 완벽히 구성 가능:**
    - 익스포터의 동작과 Monit 스크래핑 매개변수를 필요에 따라 사용자 정의할 수 있습니다.

### 설치

1. **[Go](https://golang.org/dl/) 1.27.4 이상을 설치합니다.**

2. **레포지토리를 클론하고 빌드합니다:**

    ```bash
    git clone https://github.com/ririnto/monit-exporter.git
    cd monit-exporter
    go build -o monit-exporter
    ```

### 사용법

#### 명령어

- **serve**: Monit Exporter 서버를 시작합니다.

#### 플래그

`cmd/root.go`에 정의된 플래그는 다음과 같습니다:

| 플래그                | 기본값                                                   | 설명                                                      |
|--------------------|-------------------------------------------------------|---------------------------------------------------------|
| `listen-address`   | `localhost:9388`                                      | 익스포터가 수신할 주소 및 포트 (예: '0.0.0.0:9388').                  |
| `metrics-path`     | `/metrics`                                            | 메트릭을 제공할 HTTP 경로 (예: '/metrics').                       |
| `ignore-ssl`       | `false`                                               | Monit 엔드포인트에 대해 SSL 인증서 검증을 무시할지 여부.                    |
| `monit-scrape-uri` | `http://localhost:2812/_status?format=xml&level=full` | Monit 상태 정보를 수집할 XML URL.                               |
| `monit-user`       | *(없음)*                                                | Monit에 접근하기 위한 Basic auth 사용자 이름.                       |
| `monit-password`   | *(없음)*                                                | Monit에 접근하기 위한 Basic auth 비밀번호.                         |
| `log-level`        | `info`                                                | 애플리케이션의 로그 레벨 (debug, info, warn, error, fatal, panic). |

**익스포터를 실행하려면 다음 명령어를 사용합니다:**

```bash
./monit-exporter serve \
  --listen-address="0.0.0.0:9388" \
  --monit-scrape-uri="http://localhost:2812/_status?format=xml&level=full" \
  --monit-user="admin" \
  --monit-password="monitpassword" \
  --log-level="info"
```

**메트릭 엔드포인트를 확인하려면 다음 명령어를 사용합니다:**

```bash
curl http://localhost:9388/metrics
```

#### 프로세스 리소스 메트릭

이 게이지는 `check_name` 라벨에 Monit 프로세스 서비스 이름을 표시합니다. CPU와 메모리 게이지는 `type` 라벨로 집계에 포함된 프로세스의 범위도 표시합니다.

| 메트릭 | `type` | 값 |
| --- | --- | --- |
| `monit_service_cpu_perc` | `percentage` | 프로세스 CPU 사용률(%) |
| `monit_service_cpu_perc` | `percentage_total` | 프로세스와 자식 프로세스의 CPU 사용률(%) |
| `monit_service_mem_bytes` | `process` | 프로세스 메모리 사용량(바이트) |
| `monit_service_mem_bytes` | `process_and_children` | 프로세스와 자식 프로세스의 메모리 사용량(바이트) |
| `monit_service_write_bytes_per_second` | | 디스크 쓰기 속도(초당 바이트) |
| `monit_service_write_cumulative_bytes` | | Monit이 보고한 누적 디스크 쓰기 바이트 |

익스포터는 Monit의 메모리 값을 바이트로 변환하며, Monit이 값을 보고한 시계열만 노출합니다.

#### 프로세스 재시작

`monit_service_process_pid`는 `service_name` 라벨이 있는 게이지입니다.
Monit이 프로세스 서비스의 유효한 PID를 보고할 때만 이 메트릭을 노출합니다.
`changes(monit_service_process_pid[1h])`로 지난 한 시간 동안 시계열별로 관측된 PID 변경 횟수를 세어 재시작 횟수를 추정할 수 있습니다.
Monit 폴링이나 Prometheus 스크레이프 사이에 발생한 재시작을 놓칠 수 있으며, 운영체제가 같은 PID를 재사용할 수도 있습니다.

#### Monit 수집 오류 해결

Monit이 HTTP 401을 반환하면 `--monit-user`와 `--monit-password`가 Monit HTTP 서버에 설정된 인증 정보와 일치하는지 확인하세요.
이전 버전의 익스포터는 Monit의 HTML 오류 페이지를 XML 구문 오류로 보고할 수 있습니다.

`--monit-scrape-uri`에는 `/_status?format=xml&level=full`을 사용하세요.
`format=xml`을 생략하면 Monit은 상태를 일반 텍스트로 반환합니다.

### 업스트림 항목에 대한 결정

- Monit 5.25.3과 5.26.0에서 Basic Auth 인증 정보를 생략했을 때 보고된 XML 구문 오류가 재현되었습니다.
  Monit은 HTTP 401 HTML 페이지를 반환했고, 이전 익스포터는 이 페이지를 XML로 파싱하려 했습니다.
  포크는 설정 오류를 문서화하고 수집 오류 진단을 검증하는 테스트를 추가했습니다.
  원래 보고에는 응답과 Monit 버전이 없어 정확한 원인은 확인되지 않았습니다.
  [업스트림 이슈 #8][upstream-issue-8], [포크 PR #13][fork-pr-13]
- 포크는 유효한 PID가 있는 프로세스 서비스를 위해 `monit_service_process_pid` 게이지를 추가했습니다.
  Monit 폴링이나 Prometheus 스크레이프 사이에 발생한 재시작을 놓칠 수 있으며, PID 재사용도 관측에 영향을 줍니다.
  [업스트림 이슈 #9][upstream-issue-9], [포크 PR #14][fork-pr-14]
- 포크는 프로세스별 CPU, 메모리, 디스크 쓰기 메트릭을 구현했습니다.
  Monit 응답에 값이나 서비스가 없거나 스크레이프에 실패하면 해당 시계열이 사라집니다.
  업스트림 PR의 수신 주소 기본값 변경, 기존 메트릭 이름 변경, 릴리스 도구 추가는 이 범위에서 제외했습니다.
  [업스트림 PR #7][upstream-pr-7], [포크 PR #15][fork-pr-15]
- 업스트림 PID 패치는 PID가 없을 때 0을 노출하거나 프로세스가 아닌 서비스의 PID 시계열을 생성할 수 있어 채택하지 않았습니다.
  포크는 이미 `promhttp`를 사용하며 프로세스 서비스의 유효한 PID만 노출합니다.
  [업스트림 PR #10][upstream-pr-10], [포크 PR #14][fork-pr-14]
- 익스포터는 커맨드라인 플래그로 설정을 읽으므로 포크는 `config.toml` 설정 절차를 도입하지 않기로 했습니다.
  사용하지 않는 파일의 설정 절차를 추가하면 필요성이 확인되지 않은 두 번째 설정 인터페이스가 생깁니다.
  [업스트림 이슈 #12][upstream-issue-12]

### 테스트 및 린트 실행

풀 리퀘스트와 동일한 검사를 로컬에서 실행합니다.

```bash
go test ./...
go tool -modfile=lint.mod golangci-lint run --config .golangci.yaml ./...
```

### 프로젝트 / 패키지 구조

```
.
├── cmd
│   ├── root.go       (루트 명령어와 플래그 정의)
│   └── serve.go      (서버 실행 명령어 구현)
├── internal
│   ├── config
│   │   └── config.go (익스포터 설정 구조체 정의)
│   ├── exporter
│   │   └── exporter.go (Prometheus 익스포터 로직 구현)
│   └── monit
│       └── monit.go    (Monit 상태 수집 및 파싱)
├── main.go             (진입점: cmd.Execute() 호출)
├── README.md           (이 파일)
└── LICENSE             (MIT 라이선스)
```

### 라이선스

이 프로젝트는 [MIT License](LICENSE)에 따라 라이선스가 부여됩니다.

[upstream-issue-8]: https://github.com/commercetools/monit_exporter/issues/8
[upstream-issue-9]: https://github.com/commercetools/monit_exporter/issues/9
[upstream-pr-7]: https://github.com/commercetools/monit_exporter/pull/7
[upstream-pr-10]: https://github.com/commercetools/monit_exporter/pull/10
[upstream-issue-12]: https://github.com/commercetools/monit_exporter/issues/12
[fork-pr-13]: https://github.com/ririnto/monit-exporter/pull/13
[fork-pr-14]: https://github.com/ririnto/monit-exporter/pull/14
[fork-pr-15]: https://github.com/ririnto/monit-exporter/pull/15
