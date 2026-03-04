# dartreader 코드 리뷰 및 Go(krsec) 마이그레이션 검토

> 작성일: 2026-03-04
> 대상 버전: dartreader v0.0.44 / krsec (commit 1839111)

---

## 1. dartreader 문제점 분석

### CRITICAL — 로직 결함

#### 1-1. `finstate()` 다중 회사 조회 시 None 조인 (`dart.py:143-148`)

```python
if ',' in corp:
    code_list = [self._find_corp_code(c.strip()) for c in corp.split(',')]
    corp_code = ','.join(code_list)  # None이 포함되면 "00266961,None,00126380"
```

`_find_corp_code`가 `None`을 반환해도 검증 없이 join되어 API에 `"None"` 문자열이 전달됨.

#### 1-2. `get_json()` non-200 응답 시 `None` 반환 (`rated_request.py:68-94`)

```python
if response.status_code == 200:
    ...
# else 분기 없음 → None 반환
```

호출측에서 `jo['status']` 접근 시 `TypeError: 'NoneType' object is not subscriptable` 발생.

#### 1-3. 에러 발생 후 코드 계속 실행 (`dart_report.py:56-58`, `dart_share.py:19-22`)

```python
if jo['status'] != '000' or 'list' not in jo:
    dart_request_exception(jo['status'], "")
return pd.DataFrame(jo['list'])  # status='000'이고 list 없으면 KeyError
```

`status='000'`인데 `'list'` 키가 없는 케이스에서 `dart_request_exception`은 `DartRequestError`를 발생시키지만, 에러 메시지가 빈 문자열.

---

### HIGH — 신뢰성 문제

| # | 위치 | 문제 |
|---|------|------|
| 1 | `dart.py:21-72` | Singleton 패턴: 두 번째 생성 시 다른 파라미터 silent 무시 |
| 2 | `dart_utils.py:197-200` | `dcm_no=None`일 때 print만 하고 return 없이 진행 → URL에 `dcm_no=None` |
| 3 | `dart_utils.py:54-66` | HTML 파싱 시 IndexError/AttributeError 핸들링 없음 |
| 4 | `dart_utils.py:91-100` | 정규식으로 JS 파싱, `node[123]` 하드코딩으로 node4+ 매칭 불가 |
| 5 | `semaphore.py:25-28` | Redis 필수 의존, fallback 없음 |
| 6 | `rated_request.py:50-51` | `params.update()` — 호출자 dict mutation |

---

### MEDIUM — 코드 품질

| 위치 | 문제 |
|------|------|
| `dart.py:164-166` | `ValueError` 메시지 괄호 불일치 |
| `dart.py:171` | 라이브러리에서 `print()` 사용 (logging 대신) |
| `dart_finstate.py:28,63` | `print(jo)`, `print()` — stdout 오염 |
| `dart_regstate.py:52` | 오타: `'invalid respose'` |
| `dart_event.py:60`, `dart_report.py:46` | `raise ValueError('msg', keys)` — tuple이 에러 메시지 |
| `dart_utils.py:20-28` | `_validate_dates()` 정의 후 미사용 |
| `__init__.py:10-12` | `sys.modules` 직접 조작 |
| 전체 | 날짜 파싱·에러 체크 로직 중복 (3~4곳) |

---

### 테스트 커버리지

- **있음**: `DartList` 4개 테스트 (list, company, document, document_all)
- **없음**: DartReport, DartFinstate, DartEvent, DartShare, DartRegstate, DartUtil, RatedRequester, RatedSemaphore — **전부 0개**
- 에러 경로·엣지 케이스 테스트: 0개

---

## 2. krsec 아키텍처 대비 비교

| 항목 | dartreader (Python) | krsec (Go) |
|------|---------------------|------------|
| 에러 처리 | 예외 기반이나 누락 다수, None/silent failure | error 반환 패턴 일관, fmt.Errorf + wrapping |
| 타입 안전성 | 동적 타입, 힌트 불완전 | struct 기반 컴파일 타임 검증 |
| Rate Limiting | Redis 필수, BoundedSemaphore 상속 | x/time/rate 토큰 버킷, 외부 의존 없음 |
| HTTP 클라이언트 | httpx 동기, tenacity retry | net/http + context 타임아웃 |
| 설정 관리 | Singleton (문제적) | YAML config + Validate() |
| 테스트 | 4개 | 모듈별 _test.go, mock 인터페이스 |
| 코드 구조 | 플랫 모듈 | internal/ + pkg/ + cmd/ 표준 레이아웃 |
| 인터페이스 | 없음 | Broker 인터페이스 + adapter 패턴 |
| 입력 검증 | 거의 없음 | config.Validate(), regex, 정규화 |

---

## 3. 마이그레이션 검토 의견

### 찬성 요인

1. **타입 안전성**: None 조인, non-200 미처리 등이 Go의 명시적 에러 반환으로 근본 해결
2. **인프라 공유**: Rate limiter, HTTP client, config, token 관리 등 krsec 인프라 재활용
3. **일관 아키텍처**: Broker 인터페이스에 DART를 데이터 소스로 추가, 통합 관리
4. **테스트 용이**: 인터페이스 기반 mock, _test.go 컨벤션 확립됨
5. **Redis 의존 제거**: in-process 토큰 버킷으로 대체 가능

### 고려사항

1. **pandas 대체**: Go에서는 `[]struct{}` 슬라이스, 분석 용도로는 JSON API 후 Python 소비 패턴 고려
2. **HTML 스크래핑**: list_date_ex, sub_docs 등은 goquery로 대체 가능하나 유지보수 부담 존재
3. **XML/ZIP 처리**: encoding/xml, archive/zip 표준 라이브러리로 충분

### 추천 구조

```
krsec/
├── internal/
│   └── dart/
│       ├── client.go       # HTTP 클라이언트 (rate limit 내장)
│       ├── disclosure.go   # 공시검색, 기업개황
│       ├── finstate.go     # 재무정보
│       ├── report.go       # 사업보고서
│       ├── share.go        # 지분공시
│       ├── event.go        # 주요사항보고서
│       ├── regstate.go     # 증권신고서
│       ├── corpcode.go     # 고유번호 관리
│       └── *_test.go
├── pkg/
│   └── dart/
│       ├── types.go        # 공개 타입 (CorpCode, FinState, Disclosure 등)
│       └── adapter.go      # 공개 인터페이스
```

### 마이그레이션 우선순위

1. **Phase 1**: Open DART 공식 API — 공시검색, 기업개황, 재무정보, 사업보고서 (가장 핵심이면서 잘 정의된 API)
2. **Phase 2**: 지분공시, 주요사항보고서, 증권신고서
3. **Phase 3**: 웹 스크래핑 유틸리티 (list_date_ex, sub_docs, attach_docs 등) — 공식 API가 아니므로 후순위

---

## 4. 결론

**마이그레이션 추천.** dartreader는 기능적으로 동작하지만 critical 버그, 부족한 검증, 거의 없는 테스트로 인해 신뢰성이 낮음. Python에서 이를 수정하는 것보다 krsec의 검증된 패턴 위에 Go로 재구현하는 것이 장기적으로 효율적.
