# Monit Exporter for Prometheus

> **Forked from** [commercetools/monit_exporter](https://github.com/commercetools/monit_exporter)

## English

### Introduction

Monit Exporter is a Prometheus Exporter that scrapes Monit status in XML format, then exposes the metrics via an HTTP
endpoint. It uses [Cobra](https://github.com/spf13/cobra) for the command-line interface and logs HTTP requests in
Common Log Format (CLF).

### Features

- Scrapes Monit status periodically
- Exposes Prometheus-compatible metrics
- CLI integration with Cobra
- Logs HTTP requests in CLF
- Fully configurable via command-line flags

### Installation

1. Install [Go](https://golang.org/dl/) (version 1.16 or higher recommended).
2. Clone the repository and build:

    ```bash
    git clone https://github.com/yourusername/monit_exporter.git
    cd monit_exporter
    go build -o monit_exporter
    ```

### Usage

#### Commands

- **serve**: Starts the Monit Exporter server.

#### Flags

Below is an overview of the flags defined in `cmd/root.go`:

| Flag               | Default                                               | Description                                                              |
|--------------------|-------------------------------------------------------|--------------------------------------------------------------------------|
| `listen-address`   | `localhost:9388`                                      | The address on which the exporter.go will listen (e.g., '0.0.0.0:9388'). |
| `metrics-path`     | `/metrics`                                            | The HTTP path at which metrics are served (e.g., '/metrics').            |
| `ignore-ssl`       | `false`                                               | Whether to skip SSL certificate verification for Monit endpoints.        |
| `monit-scrape-uri` | `http://localhost:2812/_status?format=xml&level=full` | The Monit status URL to scrape (XML format).                             |
| `monit-user`       | *(empty)*                                             | Basic auth username for accessing Monit.                                 |
| `monit-password`   | *(empty)*                                             | Basic auth password for accessing Monit.                                 |
| `log-level`        | `info`                                                | Log level for the application (debug, info, warn, error, fatal, panic).  |

Launch the exporter with desired flags:

```bash
./monit_exporter serve \
  --listen-address="0.0.0.0:9388" \
  --monit-scrape-uri="http://localhost:2812/_status?format=xml&level=full" \
  --monit-user="admin" \
  --monit-password="monitpassword" \
  --log-level="info"
```

Visit the metrics endpoint:

```bash
curl http://localhost:9388/metrics
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
└── main.go             (Entrypoint: calls cmd.Execute())
```

### License

This project is licensed under the [MIT License](LICENSE).

---

## 한국어

### 개요

Monit Exporter는 Monit 상태 정보를 XML 형식으로 수집하고, 이를 Prometheus 메트릭으로 변환하여 HTTP 엔드포인트로 노출하는
Exporter입니다. [Cobra](https://github.com/spf13/cobra)를 사용하여 CLI를 제공하며, HTTP 요청을 Common Log Format(CLF)으로 로깅합니다.

### 기능

- 주기적으로 Monit 상태를 스크랩
- Prometheus 호환 형식으로 메트릭 노출
- Cobra 기반 CLI
- Common Log Format 로깅
- 커맨드 라인 플래그로 모든 설정 가능

### 설치

1. [Go](https://golang.org/dl/) (버전 1.16 이상 권장)을 설치합니다.
2. 저장소를 클론하고 빌드합니다:

    ```bash
    git clone https://github.com/yourusername/monit_exporter.git
    cd monit_exporter
    go build -o monit_exporter
    ```

### 사용법

#### 명령어

- **serve**: Monit Exporter 서버를 시작합니다.

#### 플래그

`cmd/root.go`에서 정의된 플래그는 다음 표와 같습니다:

| Flag               | 기본값                                                   | 설명                                                     |
|--------------------|-------------------------------------------------------|--------------------------------------------------------|
| `listen-address`   | `localhost:9388`                                      | Exporter가 수신할 주소 및 포트 (예: `0.0.0.0:9388`)              |
| `metrics-path`     | `/metrics`                                            | 메트릭이 제공될 HTTP 경로 (예: `/metrics`)                       |
| `ignore-ssl`       | `false`                                               | Monit 엔드포인트에 대해 SSL 인증서 검증을 무시할지 여부                    |
| `monit-scrape-uri` | `http://localhost:2812/_status?format=xml&level=full` | Monit 상태를 스크랩할 XML URL                                 |
| `monit-user`       | *(없음)*                                                | Monit에 접근하기 위한 Basic auth 사용자 이름                       |
| `monit-password`   | *(없음)*                                                | Monit에 접근하기 위한 Basic auth 비밀번호                         |
| `log-level`        | `info`                                                | 애플리케이션의 로그 레벨 (debug, info, warn, error, fatal, panic) |

Exporter를 다음과 같이 실행할 수 있습니다:

```bash
./monit_exporter serve \
  --listen-address="0.0.0.0:9388" \
  --monit-scrape-uri="http://localhost:2812/_status?format=xml&level=full" \
  --monit-user="admin" \
  --monit-password="monitpassword" \
  --log-level="info"
```

그리고 다음처럼 메트릭 엔드포인트를 확인합니다:

```bash
curl http://localhost:9388/metrics
```

### 패키지 구조

```
.
├── cmd
│   ├── root.go       (루트 명령 및 플래그 설정)
│   └── serve.go      (serve 명령 구현 및 서버 실행)
├── internal
│   ├── config
│   │   └── config.go (Exporter를 위한 설정 구조체)
│   ├── exporter
│   │   └── exporter.go (Prometheus Exporter 로직 구현)
│   └── monit
│       └── monit.go    (Monit 상태를 가져오고 파싱)
└── main.go             (진입점: cmd.Execute() 호출)
```

### 라이선스

이 프로젝트는 [MIT License](LICENSE)에 따라 라이선스가 부여됩니다.
