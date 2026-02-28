# kr-broker-api

한국 증권사 통합 API 게이트웨이. 여러 증권사의 REST API를 하나의 통일된 인터페이스로 제공합니다.

## 특징

- 🏦 **통합 인터페이스**: 증권사별 API 차이를 추상화
- 🚀 **경량 바이너리**: Go로 작성된 빠르고 가벼운 서비스
- 🔐 **자동 인증**: OAuth2 토큰 자동 갱신
- 🧪 **모의투자 지원**: 실전 투자 전 안전하게 테스트
- 📦 **종목 마스터 부트스트랩**: 서버 시작 시 KIS 종목정보파일(.mst/.cod) 로딩
- 💾 **주문 컨텍스트 영속화**: 재기동 후에도 주문 조회/정정/취소 복구
- 📘 **OpenAPI 자동 노출**: `/swagger/openapi.json`, `/swagger/` (Swagger UI)

## 지원 증권사

- [x] 한국투자증권 (KIS)
- [ ] 키움증권 (예정)
- [ ] LS증권 (예정)

## 빠른 시작

### 1. 설치

```bash
git clone https://github.com/smallfish06/kr-broker-api.git
cd kr-broker-api
go mod download
```

### 2. 설정

`config.example.yaml`을 복사하여 `config.yaml`을 생성하고 자격증명을 입력하세요:

```bash
cp config.example.yaml config.yaml
```

`config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

storage:
  # Optional (recommended for deterministic local/dev paths)
  token_dir: ".kr-broker/tokens"
  order_context_dir: ".kr-broker/orders"

# enum-like guide
# - broker: ["kis"]  # currently supported
# - sandbox: [true, false]
# - account_id format: "12345678-01" (recommended), "12345678" (also accepted)
accounts:
  - name: "메인계좌"
    broker: kis
    sandbox: true
    app_key: "YOUR_APP_KEY"
    app_secret: "YOUR_APP_SECRET"
    account_id: "12345678-01"
```

### 3. 실행

```bash
# 직접 실행
make run

# 또는 빌드 후 실행
make build
./bin/kr-broker -config config.yaml
```

### 서버 없이 계좌 잔고 조회 (예제)

```bash
# 첫 번째 계좌 잔고 조회
go run ./examples/kis-balance -config config.yaml

# 특정 계좌(계좌번호 또는 이름) 조회
go run ./examples/kis-balance -config config.yaml -account 73027400-01

# 포지션까지 같이 조회
go run ./examples/kis-balance -config config.yaml -account "메인계좌" -positions
```

### 커스텀 브로커 서버 실행 (예제)

`pkg/server` + `pkg/broker`만 사용해서 외부 구현 브로커를 붙이는 샘플입니다.

```bash
go run ./examples/custom-broker-server
```

포트/계좌 ID를 바꿔 실행:

```bash
go run ./examples/custom-broker-server -host 127.0.0.1 -port 18090 -account-id demo-acc-1
```

실행 후 확인:

```bash
curl http://127.0.0.1:18090/health
curl http://127.0.0.1:18090/quotes/KRX/005930
curl http://127.0.0.1:18090/accounts/demo-acc-1/balance
curl http://127.0.0.1:18090/swagger/openapi.json
```

토큰 캐시:
- 기본 저장 경로: `<project-root>/.kr-broker/tokens` (`go.mod` 기준)
- 파일명은 `app_key` 해시 기반(`*.json`)
- 프로젝트 루트를 찾지 못하면 OS 사용자 캐시 디렉터리로 폴백
- `storage.token_dir`로 경로를 명시적으로 설정 가능

주문 컨텍스트 저장:
- 기본 저장 경로: `~/.kr-broker/orders`
- `storage.order_context_dir`로 경로를 명시적으로 설정 가능

커스텀 토큰 매니저 주입:
- `kis.TokenManager` 인터페이스를 구현해서 `kis.NewClientWithTokenManager` 또는 `kisadapter.NewAdapterWithTokenManager`에 주입 가능
- `fx`에서는 `TokenManager` 구현체를 `Provide`하고 생성자 인자로 받아 wiring 하면 됨

## API 사용법

라이브러리로 사용할 때 공용 타입/인터페이스는 [`pkg/broker`](./pkg/broker), 공개 HTTP 서버는 [`pkg/server`](./pkg/server) 패키지를 import 해서 사용합니다.

### 라이브러리로 커스텀 증권사 붙이기

```go
package main

import (
  "log"

  "github.com/smallfish06/kr-broker-api/pkg/broker"
  apiserver "github.com/smallfish06/kr-broker-api/pkg/server"
)

func main() {
  // myBroker는 broker.Broker 인터페이스 구현체
  var myBroker broker.Broker = NewMyBroker()

  srv := apiserver.New(apiserver.Options{
    Host: "0.0.0.0",
    Port: 8080,
    Accounts: []apiserver.Account{
      {ID: "my-acc-1", Name: "Main", Broker: "mybroker"},
    },
    Brokers: map[string]broker.Broker{
      "my-acc-1": myBroker,
    },
  })

  if err := srv.Run(); err != nil {
    log.Fatal(err)
  }
}
```

### Health Check

```bash
curl http://localhost:8080/health
```

### OpenAPI / Swagger UI

```bash
# OpenAPI JSON
curl http://localhost:8080/swagger/openapi.json

# Swagger UI
open http://localhost:8080/swagger/
```

### 인증

```bash
curl -X POST http://localhost:8080/auth/token \
  -H "Content-Type: application/json" \
  -d '{
    "broker": "kis",
    "credentials": {
      "app_key": "YOUR_APP_KEY",
      "app_secret": "YOUR_APP_SECRET"
    },
    "sandbox": true
  }'
```

### 현재가 조회

```bash
curl http://localhost:8080/quotes/KRX/005930
```

**응답:**

```json
{
  "ok": true,
  "data": {
    "symbol": "005930",
    "market": "KRX",
    "price": 70000,
    "open": 69500,
    "high": 70500,
    "low": 69000,
    "close": 70000,
    "volume": 15234567,
    "timestamp": "2024-01-15T15:30:00Z"
  },
  "broker": "KIS"
}
```

### 일봉 데이터 조회

```bash
curl http://localhost:8080/quotes/KRX/005930/ohlcv
```

옵션 예시:

```bash
curl "http://localhost:8080/quotes/KRX/005930/ohlcv?interval=1w&from=2026-01-01&to=2026-02-28&limit=50"
```

- `interval`: `1d` | `1w` | `1mo`
- `from`, `to`: `YYYY-MM-DD` 또는 `YYYYMMDD`
- `limit`: 최대 2000

### 종목 정보 조회

```bash
curl http://localhost:8080/instruments/KRX/005930
```

### 잔고 조회

```bash
curl http://localhost:8080/accounts/12345678-01/balance
```

### 포지션 조회

```bash
curl http://localhost:8080/accounts/12345678-01/positions
```

### 주문 실행

```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "account_id": "12345678-01",
    "symbol": "005930",
    "market": "KRX",
    "side": "buy",
    "type": "limit",
    "quantity": 10,
    "price": 70000
  }'
```

### 주문 체결내역 조회

```bash
curl http://localhost:8080/orders/000123/fills
```

## 프로젝트 구조

```
kr-broker-api/
├── cmd/
│   └── kr-broker/         # 메인 애플리케이션
├── pkg/
│   ├── broker/            # 외부 import 가능한 공통 인터페이스/타입
│   └── server/            # 외부 주입형 HTTP API 서버
├── internal/
│   ├── kis/               # 한국투자증권 도메인
│   │   └── adapter/       # Broker 인터페이스 어댑터
│   ├── config/            # 설정 관리
│   └── server/            # HTTP 서버
├── config.example.yaml    # 설정 예시
├── Makefile
└── README.md
```

## 개발

### 의존성 설치

```bash
make deps
```

### 빌드

```bash
make build
```

### 테스트

```bash
make test
```

### 코드 포맷팅

```bash
make fmt
```

## API 문서

### 공통 응답 형식

모든 API는 다음 형식으로 응답합니다:

```json
{
  "ok": true,
  "data": { ... },
  "broker": "KIS",
  "error": "error message (on failure)"
}
```

### 엔드포인트

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/auth/token` | 인증 토큰 발급 |
| GET | `/quotes/{market}/{symbol}` | 현재가 조회 |
| GET | `/quotes/{market}/{symbol}/ohlcv` | OHLCV 데이터 조회 |
| GET | `/instruments/{market}/{symbol}` | 종목 정보 조회 |
| GET | `/accounts/{account_id}/balance` | 잔고 조회 |
| GET | `/accounts/{account_id}/positions` | 포지션 조회 |
| GET | `/orders/{order_id}` | 주문 상태 조회 |
| GET | `/orders/{order_id}/fills` | 주문 체결내역 조회 |
| POST | `/orders` | 주문 실행 |
| DELETE | `/orders/{order_id}` | 주문 취소 |
| PUT | `/orders/{order_id}` | 주문 정정 |

주문 상태 조회(`GET /orders/{order_id}`)는 KIS 체결/미체결 조회 API를 우선 사용하여 상태를 산출합니다.

서버 시작 시 KIS 종목정보파일을 백그라운드에서 읽어 메모리 인덱스를 구성합니다.
`/instruments/{market}/{symbol}` 조회는 이 인덱스를 우선 사용하고, 없으면 KIS REST API로 폴백합니다.
종목 마스터는 백그라운드에서 24시간 주기로 자동 리로드됩니다.

## 한국투자증권 API 설정

1. [KIS Developers](https://apiportal.koreainvestment.com/) 접속
2. 계정 생성 및 앱 등록
3. APP Key와 APP Secret 발급
4. 모의투자 계좌 개설
5. `config.yaml`에 자격증명 입력

## 라이선스

MIT License

## 기여

이슈와 PR을 환영합니다!

## 면책 조항

이 소프트웨어는 교육 및 개인 사용 목적으로 제공됩니다. 실제 투자에 사용 시 발생하는 손실에 대해 개발자는 책임지지 않습니다.
