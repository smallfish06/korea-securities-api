# Security Policy

## 지원 버전

| 버전 | 지원 |
|---|---|
| latest release | ✅ |
| 이전 버전 | ❌ |

## 취약점 신고

보안 취약점을 발견하면 **공개 이슈를 만들지 마시고** 아래로 연락해주세요:

- GitHub Security Advisory: [Report a vulnerability](https://github.com/smallfish06/korea-securities-api/security/advisories/new)

48시간 내에 확인 후 회신하겠습니다.

## 주의사항

- `config.yaml`에는 증권사 API 키가 포함됩니다. 절대 커밋하지 마세요.
- `.gitignore`에 `config.yaml`이 포함되어 있지만, fork 시 확인해주세요.
- 모의투자(sandbox) 환경에서 먼저 테스트하세요.
