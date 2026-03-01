# latest-import-smoke

외부 소비자 관점에서 `github.com/smallfish06/krsec`를 가져와서
공개 패키지(`pkg/broker`, `pkg/server`)만으로 서버를 띄우고 호출해보는 스모크 프로젝트입니다.

## 실행

```bash
cd examples/latest-import-smoke
go run .
```

## 확인 포인트

- `GET /health`
- `GET /quotes/KRX/005930`
- `GET /accounts/demo-acc-1/balance`

실행 시 로그에 각 API 응답(JSON)이 출력됩니다.
