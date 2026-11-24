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
