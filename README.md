# korea-securities-api

[![CI](https://github.com/smallfish06/korea-securities-api/actions/workflows/ci.yml/badge.svg)](https://github.com/smallfish06/korea-securities-api/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/smallfish06/korea-securities-api.svg)](https://pkg.go.dev/github.com/smallfish06/korea-securities-api)
[![Go Report Card](https://goreportcard.com/badge/github.com/smallfish06/korea-securities-api)](https://goreportcard.com/report/github.com/smallfish06/korea-securities-api)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

한국 증권사 REST API를 하나의 통일된 인터페이스로 제공하는 Go 게이트웨이.

---

## Features

- **통합 인터페이스** — 증권사별 API 차이를 `broker.Broker` 인터페이스로 추상화
- **시세** — 현재가, 일봉(OHLCV), 호가
- **주문** — 신규/정정/취소, 체결내역 조회, 주문 상태 추적
- **계좌** — 잔고, 포지션, 멀티계좌 통합 조회
- **채권** — 채권 시세, 잔고, 일봉
- **해외 주식** — 미국(NASDAQ/NYSE/AMEX), 홍콩, 중국, 일본, 베트남
- **종목 마스터** — 서버 시작 시 KIS `.mst`/`.cod` 파일 자동 로딩 (24시간 리로드)
- **주문 컨텍스트 영속화** — 서버 재시작 후에도 주문 조회/정정/취소 복구
- **OpenAPI / Swagger UI** — 자동 생성
- **라이브러리 사용** — `pkg/broker`와 `pkg/server`를 import해서 커스텀 브로커 연결 가능

## Supported Brokers

| Broker | Status |
|---|---|
| 한국투자증권 (KIS) | ✅ Production ready |
| 키움증권 | 🚧 In progress |
| LS증권 | Planned |

## Quick Start

### Install

```bash
go install github.com/smallfish06/korea-securities-api/cmd/kr-broker@latest
```

Or download a pre-built binary from the [Releases](https://github.com/smallfish06/korea-securities-api/releases) page.

### Configure

```bash
cp config.example.yaml config.yaml
# Edit config.yaml with your broker credentials
```

```yaml
server:
  port: 8080

storage:
  token_dir: ".kr-broker/tokens"
  order_context_dir: ".kr-broker/orders"

accounts:
  - name: "main"
    broker: kis
    sandbox: true
    app_key: "YOUR_APP_KEY"
    app_secret: "YOUR_APP_SECRET"
    account_id: "12345678-01"
```

### Run

```bash
kr-broker -config config.yaml
```

## API

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/health` | Health check |
| `GET` | `/quotes/{market}/{symbol}` | 현재가 조회 |
| `GET` | `/quotes/{market}/{symbol}/ohlcv` | OHLCV 데이터 |
| `GET` | `/instruments/{market}/{symbol}` | 종목 정보 |
| `GET` | `/accounts` | 전체 계좌 목록 |
| `GET` | `/accounts/summary` | 멀티계좌 통합 잔고 |
| `GET` | `/accounts/{id}/balance` | 잔고 조회 |
| `GET` | `/accounts/{id}/positions` | 포지션 조회 |
| `POST` | `/orders` | 주문 실행 |
| `GET` | `/orders/{id}` | 주문 상태 |
| `GET` | `/orders/{id}/fills` | 체결내역 |
| `PUT` | `/orders/{id}` | 주문 정정 |
| `DELETE` | `/orders/{id}` | 주문 취소 |
| `POST` | `/auth/token` | 인증 토큰 발급 |
| `GET` | `/swagger/` | Swagger UI |

### Examples

```bash
# 삼성전자 현재가
curl http://localhost:8080/quotes/KRX/005930

# 일봉 데이터
curl "http://localhost:8080/quotes/KRX/005930/ohlcv?interval=1d&limit=30"

# 잔고 조회
curl http://localhost:8080/accounts/12345678-01/balance

# 매수 주문
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"account_id":"12345678-01","symbol":"005930","market":"KRX","side":"buy","type":"limit","quantity":10,"price":70000}'
```

All responses follow a consistent format:

```json
{
  "ok": true,
  "data": { ... },
  "broker": "KIS"
}
```

## Use as a Library

```go
import (
    "github.com/smallfish06/korea-securities-api/pkg/broker"
    apiserver "github.com/smallfish06/korea-securities-api/pkg/server"
)

func main() {
    var myBroker broker.Broker = NewMyBroker()

    srv := apiserver.New(apiserver.Options{
        Host: "0.0.0.0",
        Port: 8080,
        Accounts: []apiserver.Account{
            {ID: "acc-1", Name: "Main", Broker: "custom"},
        },
        Brokers: map[string]broker.Broker{
            "acc-1": myBroker,
        },
    })
    srv.Run()
}
```

See [`examples/`](./examples) for more.

## Project Structure

```
├── cmd/kr-broker/       # CLI entrypoint
├── pkg/
│   ├── broker/          # Public interface & types (import this)
│   └── server/          # Embeddable HTTP server
├── internal/
│   ├── kis/             # KIS raw API client
│   │   └── adapter/     # KIS → broker.Broker adapter
│   ├── kiwoom/          # Kiwoom (WIP)
│   ├── config/          # YAML config loader
│   └── server/          # Internal HTTP handlers
└── examples/
```

## Development

```bash
make deps      # Download dependencies
make test      # Run tests
make lint      # Lint (requires golangci-lint)
make build     # Build binary → bin/kr-broker
make mock      # Regenerate mocks (mockery v3)
```

## KIS API Setup

1. [KIS Developers](https://apiportal.koreainvestment.com/)에서 앱 등록
2. APP Key / APP Secret 발급
3. 모의투자 계좌 개설 (sandbox 테스트용)
4. `config.yaml`에 입력

## License

[MIT](LICENSE)

## Disclaimer

이 소프트웨어는 교육 및 개인 사용 목적으로 제공됩니다. 실제 투자에 사용 시 발생하는 손실에 대해 개발자는 책임지지 않습니다.
