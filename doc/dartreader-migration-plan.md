# dartreader → krsec Go 마이그레이션 플랜

> 작성일: 2026-03-04

---

## 목차

1. [설계 방향](#1-설계-방향)
2. [아키텍처](#2-아키텍처)
3. [Phase 1: 핵심 인프라 + 고유번호](#3-phase-1-핵심-인프라--고유번호)
4. [Phase 2: 공시검색 + 기업개황](#4-phase-2-공시검색--기업개황)
5. [Phase 3: 재무정보](#5-phase-3-재무정보)
6. [Phase 4: 사업보고서](#6-phase-4-사업보고서)
7. [Phase 5: 지분공시 + 주요사항 + 증권신고서](#7-phase-5-지분공시--주요사항--증권신고서)
8. [Phase 6: 공시서류 원본 (XML/ZIP)](#8-phase-6-공시서류-원본-xmlzip)
9. [Phase 7: 서버 라우트 통합](#9-phase-7-서버-라우트-통합)
10. [Phase 8: 웹 스크래핑 유틸리티 (선택)](#10-phase-8-웹-스크래핑-유틸리티-선택)
11. [Config 변경](#11-config-변경)
12. [의존성](#12-의존성)
13. [dartreader 기능 매핑표](#13-dartreader-기능-매핑표)

---

## 1. 설계 방향

### DART ≠ Broker

DART는 증권사 API(KIS, Kiwoom)와 달리 주문/잔고/시세 등의 거래 기능이 없는 **공시 데이터 제공 서비스**이므로, `broker.Broker` 인터페이스를 구현하지 않습니다.

대신 독립적인 `DartProvider` 인터페이스를 정의하고, 서버에서 `/dart/...` 하위 라우트로 노출합니다.

### dartreader 문제점 → Go에서 해결

| dartreader 문제 | Go 해결 방식 |
|----------------|-------------|
| `None` 조인 (finstate 다중회사) | `[]string` 각 요소 검증 후 `strings.Join` |
| non-200 `None` 반환 | 모든 HTTP 호출 `error` 반환 |
| `dart_request_exception` 후 코드 계속 실행 | `if err != nil { return }` 패턴 |
| Singleton 설정 무시 | 명시적 `Config` struct, DI |
| Redis 필수 의존 (fallback 없음) | `RateLimiter` 인터페이스: local(in-process) / redis(분산) 선택형 |
| `print()` 디버깅 | `log` 패키지 또는 structured logging |
| 입력 검증 없음 | 각 함수 진입부에서 검증 + sentinel error |
| 테스트 부재 | 모듈별 `_test.go`, httptest 기반 |

---

## 2. 아키텍처

### 파일 구조

```
krsec/
├── internal/
│   ├── ratelimit/
│   │   ├── ratelimit.go           # Waiter 인터페이스 + 기존 Limiter (변경 최소화)
│   │   ├── redis.go               # RedisLimiter (분산 rate limiting)
│   │   └── ratelimit_test.go      # 기존 테스트 + RedisLimiter 테스트
│   │
│   ├── dart/
│   │   ├── client.go              # HTTP 클라이언트 (per-key rate limit, API key 로테이션)
│   │   ├── client_test.go
│   │   ├── corpcode.go            # 고유번호 다운로드/파싱/캐싱
│   │   ├── corpcode_test.go
│   │   ├── disclosure.go          # 공시검색 (list), 기업개황 (company)
│   │   ├── disclosure_test.go
│   │   ├── finstate.go            # 재무정보 (단일/다중/전체/XBRL taxonomy)
│   │   ├── finstate_test.go
│   │   ├── report.go              # 사업보고서 (22개 키워드)
│   │   ├── report_test.go
│   │   ├── share.go               # 지분공시 (대량보유, 임원소유)
│   │   ├── share_test.go
│   │   ├── event.go               # 주요사항보고서 (37개 filing type)
│   │   ├── event_test.go
│   │   ├── regstate.go            # 증권신고서 (6개 키워드)
│   │   ├── regstate_test.go
│   │   ├── document.go            # 공시서류 원본 (XML/ZIP 다운로드)
│   │   ├── document_test.go
│   │   ├── errors.go              # DART 전용 에러 타입
│   │   └── apikey.go              # API 키 로테이션 (일일 한도 관리)
│   │
│   └── server/
│       ├── handler_dart.go        # /dart/* HTTP 핸들러
│       └── handler_dart_test.go
│
└── pkg/
    ├── dart/
    │   ├── types.go               # 공개 타입 (CorpCode, Disclosure, FinState 등)
    │   ├── provider.go            # DartProvider 인터페이스
    │   └── adapter.go             # internal/dart를 감싸는 공개 adapter
    │
    └── config/
        └── config.go             # DartConfig 추가 (api_keys, rate_limit, redis_url)
```

### 핵심 인터페이스

```go
// pkg/dart/provider.go

// Provider defines the DART data access interface.
type Provider interface {
    // Corp code management
    RefreshCorpCodes(ctx context.Context) error
    FindCorpCode(ctx context.Context, nameOrCode string) (string, error)

    // Disclosure
    ListDisclosures(ctx context.Context, opts ListOpts) ([]Disclosure, error)
    GetCompany(ctx context.Context, corpCode string) (*Company, error)

    // Financial statements
    GetFinState(ctx context.Context, corpCode string, year int, reportCode string) ([]FinStateItem, error)
    GetFinStateMulti(ctx context.Context, corpCodes []string, year int, reportCode string) ([]FinStateItem, error)
    GetFinStateAll(ctx context.Context, corpCode string, year int, reportCode string, fsDiv string) ([]FinStateItem, error)
    GetXBRLTaxonomy(ctx context.Context, sjDiv string) ([]TaxonomyItem, error)

    // Business reports
    GetReport(ctx context.Context, corpCode string, keyword string, year int, reportCode string) ([]ReportItem, error)

    // Shareholding
    GetMajorShareholders(ctx context.Context, corpCode string) ([]ShareholderItem, error)
    GetExecShareholders(ctx context.Context, corpCode string) ([]ShareholderItem, error)

    // Events
    GetEvent(ctx context.Context, corpCode string, filingType string, opts DateRange) ([]EventItem, error)

    // Registration statements
    GetRegState(ctx context.Context, corpCode string, keyword string, opts DateRange) (*RegStateResult, error)

    // Documents
    GetDocument(ctx context.Context, rcpNo string) ([]byte, error)
    GetDocumentAll(ctx context.Context, rcpNo string) ([][]byte, error)
    DownloadXBRL(ctx context.Context, rcpNo string) ([]byte, error)
}
```

---

## 3. Phase 1: 핵심 인프라 + 고유번호

> 의존: 없음 | 예상 파일: 7개

### 3-1. `pkg/dart/types.go` — 공개 타입 정의

```go
type CorpCode struct {
    CorpCode   string `json:"corp_code" xml:"corp_code"`
    CorpName   string `json:"corp_name" xml:"corp_name"`
    StockCode  string `json:"stock_code" xml:"stock_code"`
    ModifyDate string `json:"modify_date" xml:"modify_date"`
}

type DateRange struct {
    Start time.Time
    End   time.Time
}

type ListOpts struct {
    CorpCode   string
    Start      time.Time
    End        time.Time
    Kind       string // A=정기공시, B=주요사항 ...
    KindDetail string
    Final      bool
}
```

### 3-2. `internal/dart/errors.go` — 에러 타입

```go
var (
    ErrCorpNotFound     = errors.New("dart: corp not found")
    ErrRateLimited      = errors.New("dart: rate limited (status 020)")
    ErrInvalidParameter = errors.New("dart: invalid parameter")
    ErrNoData           = errors.New("dart: no data")
)

// APIError는 DART API의 status/message를 래핑
type APIError struct {
    Status  string
    Message string
}
```

- dartreader의 `dart_request_exception` → `APIError` + sentinel 에러로 대체
- status `"020"` → `ErrRateLimited` (retry 로직 트리거)

### 3-3. `internal/dart/apikey.go` — API 키 로테이션

```go
type APIKeyManager struct {
    mu       sync.Mutex
    keys     []keyState
    current  int
    timezone *time.Location // Asia/Seoul
}

type keyState struct {
    key         string
    disabledAt  *time.Time  // 일일 한도 초과 시 설정
}

func (m *APIKeyManager) GetKey() (string, error)   // 사용 가능한 키 반환
func (m *APIKeyManager) DisableKey(key string)      // 일일 한도 초과 시 비활성화
func (m *APIKeyManager) ResetIfNewDay()             // 날짜 변경 시 전체 리셋
```

- dartreader의 `RatedSemaphore.get_api_key()` + `stop()` 대체
- 일일 한도 관리 (per-key disable/enable)는 API 키 단위로 in-process 관리
- 자정 기준 자동 리셋 (KST)
- per-second rate limiting은 아래 `RateLimiter` 인터페이스에서 담당 (별개 관심사)

### 3-3b. `internal/ratelimit/` — RateLimiter 인터페이스 추상화

기존 `ratelimit.Limiter`는 in-process 전용입니다. DART는 **멀티 pod에서 동일 API 키의 per-second 한도를 공유**해야 하므로 Redis 기반 분산 rate limiting이 필요합니다.

krsec 전체의 rate limiter를 인터페이스로 추상화하면 DART뿐 아니라 KIS/Kiwoom도 향후 멀티 pod 배포 시 동일하게 대응 가능합니다.

```go
// internal/ratelimit/ratelimit.go

// Waiter is the common rate-limiter interface.
// Both in-process and distributed (Redis) implementations satisfy this.
type Waiter interface {
    // Wait blocks until a request is allowed or ctx is cancelled.
    Wait(ctx context.Context) error
    // Allow reports whether a request can proceed right now without waiting.
    Allow() bool
}

// Limiter는 기존 in-process 구현 (변경 없이 Waiter 자동 만족)
type Limiter struct { ... }  // 기존 코드 유지
```

```go
// internal/ratelimit/redis.go

// RedisLimiter implements Waiter using Redis sliding window counter.
// 멀티 pod 환경에서 per-key rate limit을 공유합니다.
type RedisLimiter struct {
    client *redis.Client
    key    string          // Redis key prefix (e.g. "dart:ratelimit:{api_key}")
    rps    int             // requests per second
}

func NewRedisLimiter(redisURL, keyPrefix string, rps int) (*RedisLimiter, error)

// Wait implements Waiter using Redis INCR + EXPIRE (sliding window).
func (r *RedisLimiter) Wait(ctx context.Context) error
func (r *RedisLimiter) Allow() bool
```

**dartreader의 Redis 사용 패턴 보존:**

dartreader의 `RatedSemaphore.lock(api_key)`는 Redis의 `SlidingWindowCounterRateLimiter`를 사용해
per-API-key per-second rate limit을 멀티 pod에서 공유합니다.

```python
# dartreader 현재 동작
while not self.limiter.test(self.limit, api_key):   # Redis sliding window 체크
    time.sleep(self.period + jitter)                 # jitter로 thundering herd 방지
self.limiter.hit(self.limit, api_key)                # Redis에 요청 카운트 기록
```

Go에서도 동일한 의미를 `RedisLimiter.Wait()`으로 구현합니다:
- Redis key: `dart:ratelimit:{api_key}` (per-key 독립)
- 알고리즘: sliding window counter (INCR + EXPIRE, 또는 Lua script)
- Thundering herd 방지: jitter 포함 backoff

**사용처 선택 로직 (Config 기반):**

```go
func NewWaiter(cfg RateLimitConfig) (Waiter, error) {
    if cfg.RedisURL != "" {
        return NewRedisLimiter(cfg.RedisURL, cfg.KeyPrefix, cfg.RPS)
    }
    return New(cfg.Name, float64(cfg.RPS), cfg.Burst), nil
}
```

- `redis_url` 설정 있음 → `RedisLimiter` (분산)
- `redis_url` 없음 → 기존 `Limiter` (in-process)
- 기존 KIS/Kiwoom 코드는 변경 없음 (`Waiter` 인터페이스를 이미 만족)

### 3-4. `internal/dart/client.go` — HTTP 클라이언트

```go
type Client struct {
    httpClient  *http.Client
    keyManager  *APIKeyManager
    newLimiter  func(apiKey string) ratelimit.Waiter  // per-key limiter factory
    limiters    map[string]ratelimit.Waiter            // api_key → limiter 캐시
    mu          sync.RWMutex
}

func NewClient(keys []string, limiterFactory func(string) ratelimit.Waiter) *Client

// 핵심 메서드: 모든 DART API 호출의 단일 진입점
func (c *Client) GetJSON(ctx context.Context, url string, params url.Values, result interface{}) error
func (c *Client) GetBytes(ctx context.Context, url string, params url.Values) ([]byte, string, error)  // contentType 반환
```

- **Per-key rate limiting**: API 키별로 독립된 limiter 인스턴스 유지
  - limiter factory를 주입받아 `RedisLimiter` 또는 `Limiter` 자동 선택
  - Redis 모드: `dart:ratelimit:{api_key}` 키로 멀티 pod 공유
  - Local 모드: `ratelimit.New("dart:"+apiKey, rps, burst)` per-key
- API 키 자동 주입: `params.Set("crtfc_key", key)`
- 에러 처리:
  - non-200 → `fmt.Errorf("dart: HTTP %d: %s", statusCode, body)`
  - status `"020"` → `keyManager.DisableKey()` + 다음 키로 재시도
- 응답 검증: `result.(DARTResponse).Status` 체크 공통화

### 3-5. `internal/dart/corpcode.go` — 고유번호 관리

```go
type CorpCodeStore struct {
    mu       sync.RWMutex
    codes    []dart.CorpCode
    nameMap  map[string]string   // corp_name → corp_code
    codeSet  map[string]struct{} // corp_code set
}

func (s *CorpCodeStore) Refresh(ctx context.Context, client *Client) error
func (s *CorpCodeStore) Find(nameOrCode string) (string, error)
func (s *CorpCodeStore) FindMulti(names []string) ([]string, error)  // None 조인 버그 해결
```

- `FindMulti`: 각 요소 검증, 하나라도 실패 시 에러 반환 (dartreader None 조인 해결)
- ZIP 다운로드 → XML 파싱: `archive/zip` + `encoding/xml` (표준 라이브러리)
- `Refresh`에서 retry with backoff (status 020)

### 3-6. 테스트

- `internal/dart/client_test.go`: `httptest.Server`로 mock DART API
- `internal/dart/corpcode_test.go`: XML 파싱, Find/FindMulti 검증
- `internal/dart/apikey_test.go`: 키 로테이션, 날짜 리셋 검증
- `internal/ratelimit/ratelimit_test.go`: Waiter 인터페이스 + RedisLimiter 테스트

### Task 목록

- [ ] `internal/ratelimit/ratelimit.go`에 `Waiter` 인터페이스 추가 (기존 `Limiter` 호환)
- [ ] `internal/ratelimit/redis.go` 작성 (Redis sliding window) + 테스트
- [ ] `pkg/dart/types.go` 작성 (CorpCode, DateRange, ListOpts 등 모든 공개 타입)
- [ ] `internal/dart/errors.go` 작성
- [ ] `internal/dart/apikey.go` 작성 (일일 한도 관리) + 테스트
- [ ] `internal/dart/client.go` 작성 (GetJSON, GetBytes, per-key limiter) + 테스트
- [ ] `internal/dart/corpcode.go` 작성 (Refresh, Find, FindMulti) + 테스트

---

## 4. Phase 2: 공시검색 + 기업개황

> 의존: Phase 1 | 예상 파일: 2개

### 4-1. `internal/dart/disclosure.go`

```go
// List calls /api/list.json with automatic pagination
func (c *Client) List(ctx context.Context, store *CorpCodeStore, opts dart.ListOpts) ([]dart.Disclosure, error)

// Company calls /api/company.json
func (c *Client) Company(ctx context.Context, corpCode string) (*dart.Company, error)
```

- **페이지네이션**: dartreader의 `for page in range(2, jo['total_page'] + 1)` → Go `for` loop
- **검증**: `corpCode` 비어있지 않은지, `opts.Start` ≤ `opts.End` 체크
- `list` 키가 없으면 빈 슬라이스 반환 (dartreader는 빈 DataFrame 반환 — 동일 의미)

### Task 목록

- [ ] `pkg/dart/types.go`에 `Disclosure`, `Company` 타입 추가
- [ ] `internal/dart/disclosure.go` 작성 + 테스트

---

## 5. Phase 3: 재무정보

> 의존: Phase 1 | 예상 파일: 2개

### 5-1. `internal/dart/finstate.go`

```go
// FinState calls /api/fnlttSinglAcnt.json (단일) or /api/fnlttMultiAcnt.json (다중)
func (c *Client) FinState(ctx context.Context, corpCodes []string, year int, reportCode string) ([]dart.FinStateItem, error)

// FinStateAll calls /api/fnlttSinglAcntAll.json
func (c *Client) FinStateAll(ctx context.Context, corpCode string, year int, reportCode string, fsDiv string) ([]dart.FinStateItem, error)

// XBRLTaxonomy calls /api/xbrlTaxonomy.json
func (c *Client) XBRLTaxonomy(ctx context.Context, sjDiv string) ([]dart.TaxonomyItem, error)
```

- **다중회사 조회**: `corpCodes []string` → 각 요소 검증 후 `strings.Join(corpCodes, ",")`. dartreader의 None 조인 버그 완전 해결
- **reportCode 검증**: 허용 값 `{"11011", "11012", "11013", "11014"}` enum 체크
- **fsDiv 검증**: `{"CFS", "OFS"}` 체크
- dartreader의 `print(jo)` → `log.Printf` 또는 에러 반환

### Task 목록

- [ ] `pkg/dart/types.go`에 `FinStateItem`, `TaxonomyItem` 타입 추가
- [ ] `internal/dart/finstate.go` 작성 + 테스트

---

## 6. Phase 4: 사업보고서

> 의존: Phase 1 | 예상 파일: 2개

### 6-1. `internal/dart/report.go`

```go
// 22개 키워드 → API 엔드포인트 매핑
var reportKeywordMap = map[string]string{
    "조건부자본증권미상환": "cndlCaplScritsNrdmpBlce",
    "미등기임원보수":      "unrstExctvMendngSttus",
    // ... (dartreader의 key_word_map 전체)
}

func (c *Client) Report(ctx context.Context, corpCode, keyword string, year int, reportCode string) ([]dart.ReportItem, error)
```

- **키워드 검증**: `reportKeywordMap`에 없으면 `ErrInvalidParameter` 반환
- dartreader의 `raise ValueError('msg', keys)` (tuple 에러) → 명확한 에러 메시지
- 에러 응답에서 빈 `message` 전달 문제 해결: DART API message 필드 그대로 전달

### Task 목록

- [ ] `pkg/dart/types.go`에 `ReportItem` 타입 추가
- [ ] `internal/dart/report.go` 작성 (키워드 맵 포함) + 테스트

---

## 7. Phase 5: 지분공시 + 주요사항 + 증권신고서

> 의존: Phase 1 | 예상 파일: 6개

### 7-1. `internal/dart/share.go`

```go
func (c *Client) MajorShareholders(ctx context.Context, corpCode string) ([]dart.ShareholderItem, error)
func (c *Client) ExecShareholders(ctx context.Context, corpCode string) ([]dart.ShareholderItem, error)
```

### 7-2. `internal/dart/event.go`

```go
// 37개 filing type → API 엔드포인트 매핑
var eventFilingTypeMap = map[string]string{
    "B0011": "dfOcr",
    "B0012": "bsnSp",
    // ... (dartreader의 filing_type_map 전체)
}

func (c *Client) Event(ctx context.Context, corpCode, filingType string, opts dart.DateRange) ([]dart.EventItem, error)
```

### 7-3. `internal/dart/regstate.go`

```go
var regstateKeywordMap = map[string]string{
    "주식의포괄적교환이전": "extrRs",
    "합병":               "mgRs",
    // ...
}

func (c *Client) RegState(ctx context.Context, corpCode, keyword string, opts dart.DateRange) (*dart.RegStateResult, error)
```

- dartreader의 `group` 응답 핸들링 (regstate는 `list` 또는 `group` 반환)
- `'invalid respose'` 오타 해결 → 정상적인 에러 반환

### Task 목록

- [ ] `pkg/dart/types.go`에 `ShareholderItem`, `EventItem`, `RegStateResult` 추가
- [ ] `internal/dart/share.go` 작성 + 테스트
- [ ] `internal/dart/event.go` 작성 (filing type 맵 포함) + 테스트
- [ ] `internal/dart/regstate.go` 작성 (keyword 맵 + group 핸들링) + 테스트

---

## 8. Phase 6: 공시서류 원본 (XML/ZIP)

> 의존: Phase 1 | 예상 파일: 2개

### 8-1. `internal/dart/document.go`

```go
// GetDocument downloads /api/document.xml and returns the first XML document
func (c *Client) GetDocument(ctx context.Context, rcpNo string) ([]byte, error)

// GetDocumentAll downloads /api/document.xml and returns all XML documents
func (c *Client) GetDocumentAll(ctx context.Context, rcpNo string) ([][]byte, error)

// DownloadXBRL downloads /api/fnlttXbrl.xml and returns the raw ZIP content
func (c *Client) DownloadXBRL(ctx context.Context, rcpNo string) ([]byte, error)
```

- ZIP 처리: `archive/zip` 표준 라이브러리
- XML 파싱: `encoding/xml` 표준 라이브러리
- 인코딩 감지: dartreader의 `decode_bytes` (euc-kr/utf-8/cp949) → `golang.org/x/text/encoding/korean` 활용

### Task 목록

- [ ] `internal/dart/document.go` 작성 + 테스트

---

## 9. Phase 7: 서버 라우트 통합

> 의존: Phase 1-6 | 예상 파일: 4개

### 9-1. Config 확장

```go
// pkg/config/config.go에 추가
type DartConfig struct {
    APIKeys   []string `yaml:"api_keys"`
    RateLimit float64  `yaml:"rate_limit"`  // requests per second per key (기본: 10)
    RedisURL  string   `yaml:"redis_url"`   // 분산 rate limiting (선택)
}
```

### 9-2. `internal/server/handler_dart.go`

```
REST 라우트:

GET  /dart/disclosures                               공시검색
GET  /dart/companies/{corp}                           기업개황
GET  /dart/companies/{corp}/finstate                  재무정보 (단일)
GET  /dart/companies/finstate                         재무정보 (다중, ?corps=code1,code2)
GET  /dart/companies/{corp}/finstate/all              전체 재무제표
GET  /dart/companies/{corp}/reports/{keyword}         사업보고서
GET  /dart/companies/{corp}/shareholders              대량보유 상황보고
GET  /dart/companies/{corp}/shareholders/exec         임원 소유보고
GET  /dart/companies/{corp}/events/{filing_type}      주요사항보고서
GET  /dart/companies/{corp}/regstate/{keyword}        증권신고서
GET  /dart/documents/{rcp_no}                         공시서류 원본
GET  /dart/documents/{rcp_no}/all                     공시서류 전체
GET  /dart/taxonomy/{sj_div}                          XBRL 표준계정과목
GET  /dart/xbrl/{rcp_no}                              XBRL 원본 다운로드
```

- krsec 기존 패턴 준수: `fuego.Get/Post` + `OptionTags("DART")`
- `corp` 파라미터: 회사명 또는 고유번호 모두 허용 (내부에서 `CorpCodeStore.Find` 호출)
- 응답: 기존 `server.Response` 구조 (`{ok, data, error}`) 사용

### 9-3. `internal/server/server.go` 수정

```go
// New() 함수에 DART 초기화 추가
case "dart":
    // DartConfig에서 초기화
```

### Task 목록

- [ ] `pkg/config/config.go`에 `DartConfig` 추가 + Validate 로직
- [ ] `internal/server/handler_dart.go` 작성 + 테스트
- [ ] `internal/server/server.go`에 DART 라우트 등록
- [ ] `config.example.yaml` 업데이트

---

## 10. Phase 8: 웹 스크래핑 유틸리티 (선택)

> 의존: Phase 1 | 우선순위: 낮음

dartreader의 `DartUtil` 클래스 기능:

| 메서드 | 설명 | 대상 URL | 마이그레이션 권장 |
|--------|------|----------|-----------------|
| `list_date_ex` | 특정 날짜 공시 목록 (시간 포함) | `dart.fss.or.kr/dsac001/search.ax` | 검토 필요 |
| `sub_docs` | 하위 문서 목록 | `dart.fss.or.kr/dsaf001/main.do` | 검토 필요 |
| `attach_docs` | 첨부문서 목록 | `dart.fss.or.kr/dsaf001/main.do` | 검토 필요 |
| `attach_files` | 첨부파일 목록 | `dart.fss.or.kr/pdf/download/main.do` | 검토 필요 |
| `download` | URL 파일 다운로드 | 임의 URL | 불필요 (범용 기능) |

**유의사항:**
- 이 기능들은 **공식 Open DART API가 아닌 웹 스크래핑**
- DART 웹사이트 구조 변경 시 즉시 깨짐
- 정규식으로 JS 파싱하는 `sub_docs`는 특히 brittle
- 마이그레이션 시 `goquery` 라이브러리 필요
- 별도 논의 후 필요한 것만 선택적으로 구현 권장

### Task 목록 (실행 시)

- [ ] `internal/dart/scraper.go` — list_date_ex, sub_docs, attach_docs, attach_files
- [ ] 테스트 (HTML 응답 fixture 기반)

---

## 11. Config 변경

### config.example.yaml 추가 항목

```yaml
# DART 전자공시 설정 (선택)
dart:
  api_keys:
    - "your-dart-api-key-1"
    - "your-dart-api-key-2"     # 여러 키 로테이션 지원
  rate_limit: 10                # requests/second per key (기본: 10)
  redis_url: "redis://localhost:6379"  # 분산 rate limiting (선택)
```

### 검증 규칙

- `dart` 섹션 자체는 optional (DART 미사용 시 생략 가능)
- `dart` 섹션이 있으면 `api_keys`에 최소 1개 키 필수
- `rate_limit` ≤ 0이면 기본값 10 적용
- `redis_url`:
  - 설정 시 → Redis 기반 분산 rate limiting (멀티 pod 환경)
  - 미설정 시 → in-process rate limiting (단일 pod / 로컬 개발)

---

## 12. 의존성

### 새로 추가되는 의존성

| 패키지 | 용도 | Phase | 필수 |
|--------|------|-------|------|
| `github.com/redis/go-redis/v9` | 분산 rate limiting (Redis sliding window) | Phase 1 | 선택 (redis_url 설정 시) |

### 이미 krsec에 있는 의존성 (재활용)

| 패키지 | 용도 |
|--------|------|
| `golang.org/x/time/rate` | In-process rate limiting (fallback) |
| `golang.org/x/text` | euc-kr/cp949 디코딩 (이미 go.mod에 존재) |
| `encoding/xml` | XML 파싱 (표준 라이브러리) |
| `archive/zip` | ZIP 처리 (표준 라이브러리) |
| `net/http` | HTTP 클라이언트 (표준 라이브러리) |
| `github.com/stretchr/testify` | 테스트 assertion |

### Phase 8에서만 필요 (선택)

| 패키지 | 용도 |
|--------|------|
| `github.com/PuerkitoBio/goquery` | HTML 스크래핑 |

### build tag를 통한 Redis 선택적 컴파일

Redis 의존성을 원하지 않는 사용자를 위해 build tag로 분리 가능:

```go
//go:build redis

package ratelimit

// redis.go — go-redis 의존
```

`go build -tags redis` 시에만 Redis 구현 포함. tag 없으면 in-process만 사용.
단, 이는 선택사항이며 go-redis가 부담되지 않으면 tag 없이 포함해도 무방합니다.

---

## 13. dartreader 기능 매핑표

| dartreader 메서드 | Open DART API | Go 함수 | Phase |
|------------------|---------------|---------|-------|
| `refresh_company()` | `/api/corpCode.xml` | `CorpCodeStore.Refresh()` | 1 |
| `_find_corp_code()` | — | `CorpCodeStore.Find()` | 1 |
| `list()` | `/api/list.json` | `Client.List()` | 2 |
| `company()` | `/api/company.json` | `Client.Company()` | 2 |
| `finstate()` | `/api/fnlttSinglAcnt.json`, `/api/fnlttMultiAcnt.json` | `Client.FinState()` | 3 |
| `finstate_all()` | `/api/fnlttSinglAcntAll.json` | `Client.FinStateAll()` | 3 |
| `xbrl_taxonomy()` | `/api/xbrlTaxonomy.json` | `Client.XBRLTaxonomy()` | 3 |
| `report()` | `/api/{keyword}.json` (22개) | `Client.Report()` | 4 |
| `major_shareholders()` | `/api/majorstock.json` | `Client.MajorShareholders()` | 5 |
| `major_shareholders_exec()` | `/api/elestock.json` | `Client.ExecShareholders()` | 5 |
| `event()` | `/api/{filingType}.json` (37개) | `Client.Event()` | 5 |
| `regstate()` | `/api/{keyword}.json` (6개) | `Client.RegState()` | 5 |
| `document()` | `/api/document.xml` | `Client.GetDocument()` | 6 |
| `document_all()` | `/api/document.xml` | `Client.GetDocumentAll()` | 6 |
| `finstate_xml()` | `/api/fnlttXbrl.xml` | `Client.DownloadXBRL()` | 6 |
| `list_date_ex()` | 웹 스크래핑 | `Scraper.ListDateEx()` | 8 |
| `sub_docs()` | 웹 스크래핑 | `Scraper.SubDocs()` | 8 |
| `attach_docs()` | 웹 스크래핑 | `Scraper.AttachDocs()` | 8 |
| `attach_files()` | 웹 스크래핑 | `Scraper.AttachFiles()` | 8 |
| `download()` | — | (범용 HTTP, 별도 구현 불필요) | — |

---

## 실행 순서 요약

```
Phase 1 (핵심 인프라)
  ├── types.go, errors.go
  ├── apikey.go + test
  ├── client.go + test
  └── corpcode.go + test
      │
      ├── Phase 2 (공시검색) ──── disclosure.go + test
      ├── Phase 3 (재무정보) ──── finstate.go + test
      ├── Phase 4 (사업보고서) ── report.go + test
      ├── Phase 5 (지분/이벤트) ─ share.go, event.go, regstate.go + tests
      └── Phase 6 (문서 원본) ── document.go + test
          │
          Phase 7 (서버 통합) ── handler_dart.go, config 확장
          │
          Phase 8 (스크래핑) ── [선택] scraper.go
```

Phase 2-6은 Phase 1 완료 후 **병렬 진행 가능**합니다.
