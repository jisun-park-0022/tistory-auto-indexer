# Google Search Console API 인증 설정 
 - Google Cloud Console에서 Google Search Console에 대한 인증을 진행합니다. 
 - 인증 진행 후 
  ---
  ### 1. Google Cloud 프로젝트 생성 (또는 기존 것 사용)

  1. https://console.cloud.google.com 접속
  2. 상단 프로젝트 선택 → 새 프로젝트 생성

  ---
  ### 2. Search Console API 활성화

  1. 좌측 메뉴 → API 및 서비스 → 라이브러리
  2. Google Search Console API 검색 → 사용 설정

  ---
  ### 3. OAuth 동의 화면 설정

  1. https://console.cloud.google.com → 해당 프로젝트 선택
  2. 좌측 메뉴 → API 및 서비스 → OAuth 동의 화면
  3. User Type: 외부 선택 → 만들기
  4. 앱 이름: tistory-indexer (아무거나 가능)
  5. 사용자 지원 이메일: 본인 Gmail 입력
  6. 개발자 연락처 이메일: 본인 Gmail 입력
  7. 저장 후 계속 → 범위 설정은 건너뜀 → 저장 후 계속
  8. 테스트 사용자 → + ADD USERS → 본인 Gmail 추가
  9. 저장 후 계속 → 완료

  ---
  ### 4. OAuth2 클라이언트 ID 생성

  1. 좌측 메뉴 → API 및 서비스 → 사용자 인증 정보
  2. 상단 + 사용자 인증 정보 만들기 → OAuth 클라이언트 ID
  3. 애플리케이션 유형: 데스크톱 앱 선택
  4. 이름: tistory-indexer-desktop (아무거나 가능)
  5. 만들기
  6. 좌측 메뉴 → API 및 서비스 → 라이브러리 검색 후 Google Search Console API 검색 → '사용' 클릭

  ---
  ### 5. Client ID / Secret 확인

  1. 생성 완료 팝업에서 바로 확인 <br>
  또는 사용자 인증 정보 목록에서 방금 만든 항목 클릭 시 확인

  ```
  - 클라이언트 ID:     xxxx.apps.googleusercontent.com
  - 클라이언트 보안 비밀번호: GOCSPX-xxxx
  ```

  ---
  ### 6. Refresh Token 발급
  1. ```go run cmd/authorize/main.go``` 수행
  2. 출력된 URL을 브라우저로 열어서 로그인
  3. 승인하면 refresh token 발급

  ```GOOGLE_REFRESH_TOKEN=1//xxx```

  ---
  ### 7. .env 파일 생성

  ```sh
  $ vi ./tistory-indexer/.env
  GOOGLE_CLIENT_ID=xxxx.apps.googleusercontent.com
  GOOGLE_CLIENT_SECRET=GOCSPX-xxxx
  GOOGLE_REFRESH_TOKEN=xxx
  ```
  ---
  
  ### 8. 실행 테스트

  ```sh
  $ go run cmd/indexer/main.go
  ```

  ### 9. 실행
  ```sh
  # 빌드
  $ go build -o server.exe ./cmd/server/
  
  # INFO 레벨로 실행
  $ ./server

  # Debug 모드로 실행
  $ $env:LOG_LEVEL="debug"; ./server
  ```
  - 브라우저에서 'http://localhost:8090'으로 접속
