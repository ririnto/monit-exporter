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

#### Troubleshooting Monit scrapes

If Monit returns HTTP 401, check that `--monit-user` and `--monit-password` match the credentials configured for Monit's HTTP server.
Older exporter releases could report Monit's HTML error page as an XML syntax error.

Use `/_status?format=xml&level=full` in `--monit-scrape-uri`.
Without `format=xml`, Monit serves the status as plain text.

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

#### Monit 수집 오류 해결

Monit이 HTTP 401을 반환하면 `--monit-user`와 `--monit-password`가 Monit HTTP 서버에 설정된 인증 정보와 일치하는지 확인하세요.
이전 버전의 익스포터는 Monit의 HTML 오류 페이지를 XML 구문 오류로 보고할 수 있습니다.

`--monit-scrape-uri`에는 `/_status?format=xml&level=full`을 사용하세요.
`format=xml`을 생략하면 Monit은 상태를 일반 텍스트로 반환합니다.

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
