# korea-securities-api

[![CI](https://github.com/smallfish06/korea-securities-api/actions/workflows/ci.yml/badge.svg)](https://github.com/smallfish06/korea-securities-api/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/smallfish06/korea-securities-api.svg)](https://pkg.go.dev/github.com/smallfish06/korea-securities-api)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

한국 증권사 REST API 통합 게이트웨이. 증권사마다 다른 인증, 파라미터, 응답 구조를 `broker.Broker` 인터페이스 하나로 통일합니다.

## 지원 증권사

| 증권사 | 상태 |
|---|---|
| 한국투자증권 (KIS) | ✅ |
| 키움증권 | ✅ |
| LS증권 | 예정 |

## 설치

```bash
go install github.com/smallfish06/korea-securities-api/cmd/kr-broker@latest
```

바이너리: [Releases](https://github.com/smallfish06/korea-securities-api/releases)

## 설정

```bash
cp config.example.yaml config.yaml
```

```yaml
server:
  port: 8080

accounts:
  - name: "main"
    broker: kis
    sandbox: true
    app_key: "YOUR_APP_KEY"
    app_secret: "YOUR_APP_SECRET"
    account_id: "12345678-01"
```

## 실행

```bash
kr-broker -config config.yaml
```

## API

| Method | Path | 설명 |
|---|---|---|
| `GET` | `/quotes/{market}/{symbol}` | 현재가 |
| `GET` | `/quotes/{market}/{symbol}/ohlcv` | 일봉 |
| `GET` | `/instruments/{market}/{symbol}` | 종목 정보 |
| `GET` | `/accounts/{id}/balance` | 잔고 |
| `GET` | `/accounts/{id}/positions` | 포지션 |
| `GET` | `/accounts/summary` | 멀티계좌 통합 잔고 |
| `POST` | `/orders` | 주문 |
| `GET` | `/orders/{id}` | 주문 상태 |
| `GET` | `/orders/{id}/fills` | 체결내역 |
| `PUT` | `/orders/{id}` | 주문 정정 |
| `DELETE` | `/orders/{id}` | 주문 취소 |
| `GET` | `/swagger/` | Swagger UI |

```bash
curl http://localhost:8080/quotes/KRX/005930
```

```json
{"ok": true, "data": {"symbol": "005930", "price": 70000, ...}, "broker": "KIS"}
```

## 라이브러리로 사용

```go
import (
    "github.com/smallfish06/korea-securities-api/pkg/broker"
    apiserver "github.com/smallfish06/korea-securities-api/pkg/server"
)

srv := apiserver.New(apiserver.Options{
    Port: 8080,
    Accounts: []apiserver.Account{{ID: "acc-1", Name: "Main", Broker: "custom"}},
    Brokers:  map[string]broker.Broker{"acc-1": myBroker},
})
srv.Run()
```

## 구조

```
cmd/kr-broker/        서버
pkg/broker/           공개 인터페이스
pkg/server/           임베드 가능한 HTTP 서버
internal/kis/         KIS 클라이언트 + 어댑터
internal/kiwoom/      키움 클라이언트 + 어댑터
internal/server/      HTTP 핸들러
examples/             사용 예시
```

## 개발

```bash
make test      # 테스트
make build     # 빌드
make lint      # lint
make mock      # mock 재생성
```

## License

[Apache 2.0](LICENSE)
