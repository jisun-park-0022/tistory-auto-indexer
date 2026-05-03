# Tistory Auto Indexer - Architecture Design

## 개요

Tistory 블로그의 sitemap.xml을 파싱하여 신규 포스트를 감지하고,
Google Search Console에 Sitemap을 자동 제출하는 Go 프로그램.

> Tistory Open API는 2024년 2월 종료됨. 인증 없이 공개된 sitemap.xml 파싱으로 대체.

---

## 목표

- Tistory `sitemap.xml` HTTP GET으로 포스트 URL 목록 파싱
- 이전 실행 대비 신규 포스트 URL 감지
- 신규 포스트 존재 시 Google Search Console API에 Sitemap 제출
- 로컬 cron으로 주기 실행

---

## 시스템 구성도

```
┌──────────────────────────────────────────────────────────┐
│                    tistory-indexer                        │
│                                                           │
│  ┌─────────────┐    ┌──────────┐    ┌─────────────────┐  │
│  │   Sitemap   │    │  State   │    │  Google Search  │  │
│  │   Parser    │───▶│  Store   │───▶│  Console Client │  │
│  └─────────────┘    └──────────┘    └─────────────────┘  │
│         │                │                  │            │
│  HTTP GET (인증없음)  last_state.json    Sitemap 제출    │
│  sitemap.xml 파싱    (로컬 파일 저장)   (PUT /sitemaps) │
└──────────────────────────────────────────────────────────┘
          │                                   │
 [tistory.com/sitemap.xml]      [Google Search Console API]
```

---

## 디렉토리 구조

```
tistory-indexer/
├── cmd/
│   └── indexer/
│       └── main.go              # 진입점, config 로드, 실행 오케스트레이션
├── internal/
│   ├── sitemap/
│   │   ├── parser.go            # sitemap.xml HTTP fetch + XML 파싱
│   │   └── model.go             # URL, Sitemap 구조체
│   ├── gsc/
│   │   ├── client.go            # Google Search Console API 클라이언트
│   │   └── model.go             # Sitemap 요청/응답 구조체
│   ├── state/
│   │   └── store.go             # 마지막 조회 상태 저장/로드 (JSON 파일)
│   └── indexer/
│       └── service.go           # 핵심 비즈니스 로직 (신규 URL 감지 → 제출)
├── pkg/
│   └── config/
│       └── config.go            # 환경변수/설정 파일 로드
├── data/
│   └── last_state.json          # 실행 상태 저장 (gitignore 대상)
├── credentials/
│   └── .gitkeep                 # service account JSON 키 위치 (gitignore 대상)
├── config.yaml                  # 설정 파일 (비밀값 제외)
├── .env.example                 # 환경변수 예시
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

---

## 핵심 플로우

```
main()
  │
  ├─ 1. Config 로드 (.env + config.yaml)
  │
  ├─ 2. sitemap.xml HTTP GET 요청
  │       └─ https://[블로그명].tistory.com/sitemap.xml
  │           XML 파싱 → URL 목록 추출
  │
  ├─ 3. State Store에서 이전 실행 URL 목록 로드
  │       └─ 파일 없음 → 빈 state로 초기화
  │
  ├─ 4. 신규 URL 감지 (diff)
  │       └─ 신규 없음 → State 갱신 없이 종료
  │           신규 있음 → 다음 단계
  │
  ├─ 5. Google Search Console API로 Sitemap 제출
  │       └─ PUT /webmasters/v3/sites/{siteUrl}/sitemaps/{sitemapUrl}
  │
  └─ 6. State Store 업데이트 (현재 URL 목록 + 제출 시각 저장)
```

---

## API 명세

### Tistory sitemap.xml

| 항목 | 내용 |
|------|------|
| 인증 | 불필요 (공개 URL) |
| URL | `https://[블로그명].tistory.com/sitemap.xml` |
| 포맷 | XML (표준 Sitemap 0.9 스펙) |
| 주요 태그 | `<url>`, `<loc>`, `<lastmod>` |
| 특이사항 | Tistory가 자동 생성 및 갱신 |

**sitemap.xml 예시:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://example.tistory.com/123</loc>
    <lastmod>2026-05-01</lastmod>
  </url>
  <url>
    <loc>https://example.tistory.com/124</loc>
    <lastmod>2026-05-03</lastmod>
  </url>
</urlset>
```

### Google Search Console API

| 항목 | 내용 |
|------|------|
| 인증 방식 | Google Service Account (JWT → OAuth2 Bearer) |
| Sitemap 제출 | `PUT https://www.googleapis.com/webmasters/v3/sites/{siteUrl}/sitemaps/{feedpath}` |
| siteUrl 형식 | `https://[블로그명].tistory.com/` (URL 인코딩 필요) |
| feedpath | `https://[블로그명].tistory.com/sitemap.xml` (URL 인코딩 필요) |
| 성공 응답 | HTTP 204 No Content |

---

## 데이터 모델

### State Store (data/last_state.json)

```json
{
  "last_run_at": "2026-05-03T10:00:00Z",
  "last_submitted_at": "2026-05-03T10:00:01Z",
  "known_urls": [
    "https://example.tistory.com/123",
    "https://example.tistory.com/124"
  ]
}
```

### Config

```yaml
# config.yaml
tistory:
  sitemap_url: "https://your-blog.tistory.com/sitemap.xml"

google:
  site_url: "https://your-blog.tistory.com/"
  sitemap_url: "https://your-blog.tistory.com/sitemap.xml"

state:
  file_path: "./data/last_state.json"

http:
  timeout_seconds: 10
  user_agent: "tistory-indexer/1.0"
```

```env
# .env (gitignore 대상)
GOOGLE_SERVICE_ACCOUNT_JSON=./credentials/gsc-service-account.json
```

---

## 주요 인터페이스

```go
// internal/sitemap/parser.go
type Parser interface {
    Fetch(ctx context.Context, sitemapURL string) (*Sitemap, error)
}

// internal/gsc/client.go
type GSCClient interface {
    SubmitSitemap(ctx context.Context, siteURL, sitemapURL string) error
}

// internal/state/store.go
type StateStore interface {
    Load() (*State, error)
    Save(state *State) error
}

// internal/sitemap/model.go
type Sitemap struct {
    URLs []SitemapURL
}

type SitemapURL struct {
    Loc     string
    LastMod time.Time
}

// internal/state/store.go
type State struct {
    LastRunAt       time.Time `json:"last_run_at"`
    LastSubmittedAt time.Time `json:"last_submitted_at"`
    KnownURLs       []string  `json:"known_urls"`
}

// internal/indexer/service.go
type IndexerService struct {
    parser  sitemap.Parser
    gsc     gsc.GSCClient
    state   state.StateStore
    config  *config.Config
}

func (s *IndexerService) Run(ctx context.Context) error
```

---

## 신규 URL 감지 로직

```go
// set 기반 diff — O(n) 비교
func detectNewURLs(current []string, known []string) []string {
    knownSet := make(map[string]struct{}, len(known))
    for _, u := range known {
        knownSet[u] = struct{}{}
    }
    var newURLs []string
    for _, u := range current {
        if _, exists := knownSet[u]; !exists {
            newURLs = append(newURLs, u)
        }
    }
    return newURLs
}
```

---

## 외부 의존성

```
golang.org/x/oauth2                    # Google Service Account 인증
google.golang.org/api/webmasters/v3    # Search Console API
github.com/spf13/viper                 # config 로드
github.com/joho/godotenv               # .env 파일 로드
```

> `encoding/xml` 은 Go 표준 라이브러리로 외부 의존성 없이 sitemap 파싱 가능

---

## 에러 처리 전략

| 시나리오 | 처리 방식 |
|----------|----------|
| sitemap.xml fetch 실패 (네트워크) | 에러 반환, state 업데이트 안 함 |
| sitemap.xml 파싱 실패 (XML 손상) | 에러 반환, state 업데이트 안 함 |
| GSC API 실패 (4xx/5xx) | 에러 로그 후 state 업데이트 건너뜀 → 다음 실행에 재시도 |
| GSC API rate limit (429) | exponential backoff, 최대 3회 재시도 |
| State 파일 없음 | 빈 state로 초기화 (첫 실행 정상 케이스) |
| State 파일 손상 | 빈 state로 초기화 + 경고 로그 |

---

## 실행 방식

```bash
# 단발 실행
go build -o tistory-indexer.exe ./cmd/indexer/
./tistory-indexer

# Debug 모드로 실행
$env:LOG_LEVEL="debug"; go run cmd/indexer/main.go

# cron 등록 예시 (매일 오전 9시)
0 9 * * * /path/to/tistory-indexer >> /var/log/tistory-indexer.log 2>&1
```

---

## 사전 준비 작업 (개발 전)

1. **Google Cloud 프로젝트 생성** → Search Console API 활성화
2. **OAuth2 클라이언트 ID 생성**
   - API 및 서비스 → 사용자 인증 정보 → OAuth 클라이언트 ID 만들기
   - 애플리케이션 유형: **데스크톱 앱** 선택
   - Client ID / Client Secret 발급
3. **Refresh Token 발급** (1회)
   ```powershell
   go run cmd/authorize/main.go
   ```
   브라우저에서 본인 Google 계정으로 로그인 → 승인 → 터미널에 Refresh Token 출력
4. **`.env` 파일 생성**
   ```
   GOOGLE_CLIENT_ID=xxxx.apps.googleusercontent.com
   GOOGLE_CLIENT_SECRET=GOCSPX-xxxx
   GOOGLE_REFRESH_TOKEN=1//xxxx...
   ```
5. **Search Console 사이트 소유권 인증** → 로그인한 Google 계정이 속성 소유자인지 확인

> **Note**: Service Account 방식은 Google Search Console UI가 서비스 계정 이메일을 Google 계정으로 인식하지 못해 권한 부여 불가. OAuth2 Refresh Token 방식으로 대체.

---

## 개발 순서 (권장)

1. `pkg/config` — 설정 로드
2. `internal/sitemap` — sitemap.xml fetch + XML 파싱
3. `internal/state` — 상태 저장소
4. `internal/gsc` — GSC API 클라이언트 (Service Account 인증 포함)
5. `internal/indexer` — 비즈니스 로직 조합 (신규 감지 → 제출)
6. `cmd/indexer/main.go` — 진입점 연결
