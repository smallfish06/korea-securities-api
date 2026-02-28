# Contributing

PR과 이슈 환영합니다.

## 개발 환경

```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
make deps
```

## PR 보내기 전에

```bash
make test
make lint
```

- 새 기능에는 테스트를 같이 작성해주세요.
- 커밋 메시지는 [Conventional Commits](https://www.conventionalcommits.org/) 형식을 따릅니다.
  - `feat:`, `fix:`, `docs:`, `test:`, `ci:`, `chore:`

## 새 증권사 추가

1. `internal/{broker}/` 에 raw API 클라이언트 작성
2. `internal/{broker}/adapter/` 에 `broker.Broker` 인터페이스 구현
3. `internal/server/server.go`에 브로커 wiring 추가
4. 테스트 작성
5. README 업데이트

## 이슈

버그 리포트나 기능 제안은 GitHub Issues를 이용해주세요.
