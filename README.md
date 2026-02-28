# krsec

[![CI](https://github.com/smallfish06/krsec/actions/workflows/ci.yml/badge.svg)](https://github.com/smallfish06/krsec/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/smallfish06/krsec.svg)](https://pkg.go.dev/github.com/smallfish06/krsec)
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
go install github.com/smallfish06/krsec/cmd/krsec@latest
```

바이너리: [Releases](https://github.com/smallfish06/krsec/releases)

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
krsec -config config.yaml
```

## API

| Method | Path | 설명 |
|---|---|---|
| `POST` | `/kis/{path...}` | KIS 엔드포인트 호출 (`/uapi` 경로를 구현된 함수로 매핑) |
| `POST` | `/kiwoom/{path...}` | Kiwoom 엔드포인트 호출 (`/api` 경로 + `api_id`를 구현/문서 등록(비웹소켓 REST) 함수로 명시 매핑) |
| `GET` | `/quotes/{market}/{symbol}` | 현재가 |
| `GET` | `/quotes/{market}/{symbol}/ohlcv` | 일봉 |
| `GET` | `/instruments/{market}/{symbol}` | 종목 정보 |
| `GET` | `/accounts/{id}/balance` | 잔고 |
| `GET` | `/accounts/{id}/positions` | 포지션 |
| `GET` | `/accounts/summary` | 멀티계좌 통합 잔고 |
| `POST` | `/accounts/{account_id}/orders` | 주문 |
| `GET` | `/accounts/{account_id}/orders/{id}` | 주문 상태 |
| `GET` | `/accounts/{account_id}/orders/{id}/fills` | 체결내역 |
| `PUT` | `/accounts/{account_id}/orders/{id}` | 주문 정정 |
| `DELETE` | `/accounts/{account_id}/orders/{id}` | 주문 취소 |
| `GET` | `/swagger/` | Swagger UI |

```bash
curl http://localhost:8080/quotes/KRX/005930
```

```json
{"ok": true, "data": {"symbol": "005930", "price": 70000, ...}, "broker": "KIS"}
```

```bash
curl -X POST http://localhost:8080/kis/overseas-price/v1/quotations/price \
  -H "Content-Type: application/json" \
  -d '{
    "tr_id": "HHDFS00000300",
    "params": {
      "AUTH": "",
      "EXCD": "NAS",
      "SYMB": "AAPL"
    }
  }'
```

## 라이브러리로 사용

```go
import (
    "github.com/smallfish06/krsec/pkg/broker"
    apiserver "github.com/smallfish06/krsec/pkg/server"
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
cmd/krsec/        서버
pkg/broker/           공개 인터페이스
pkg/server/           임베드 가능한 HTTP 서버
internal/kis/         KIS 클라이언트 + 어댑터
internal/kiwoom/      키움 클라이언트 + 어댑터
internal/server/      HTTP 핸들러
examples/             사용 예시
```

## 범위

REST API만 지원합니다. WebSocket(실시간 시세, 실시간 체결 등)은 지원 범위에 포함되지 않습니다.

## 개발

```bash
make test      # 테스트
make build     # 빌드
make lint      # lint
make mock      # mock 재생성
```

## License

[Apache 2.0](LICENSE)
