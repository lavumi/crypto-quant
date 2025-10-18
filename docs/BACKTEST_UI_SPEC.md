# 백테스트 UI 기획서

> **Version**: 1.0  
> **Last Updated**: 2025-10-18  
> **Status**: In Development

## 🎯 목표

사용자가 손쉽게 트레이딩 전략을 백테스트하고 결과를 분석할 수 있는 직관적인 인터페이스 제공

---

## 📱 화면 구성

### 페이지 1: 백테스트 실행 페이지

**Route**: `/backtest/new`

#### 레이아웃 구조

```
┌─────────────────────────────────────────────────────┐
│  Header: 백테스트 실행                                 │
├─────────────────────────────────────────────────────┤
│                                                     │
│  ┌─────────────────┐  ┌──────────────────────────┐ │
│  │   기본 설정      │  │   전략 파라미터           │ │
│  │                 │  │                          │ │
│  │ - Symbol        │  │  전략 선택에 따라         │ │
│  │ - Interval      │  │  동적으로 변경됨          │ │
│  │ - 날짜 범위      │  │                          │ │
│  │ - 초기 자금      │  │  [MA Cross]             │ │
│  │ - 수수료        │  │  - Fast Period: 10      │ │
│  │                 │  │  - Slow Period: 30      │ │
│  └─────────────────┘  │                          │ │
│                       │  [RSI]                   │ │
│  ┌─────────────────┐  │  - Period: 14           │ │
│  │  전략 선택       │  │  - Oversold: 30         │ │
│  │                 │  │  - Overbought: 70       │ │
│  │  ○ MA Cross     │  │                          │ │
│  │  ○ RSI          │  │  [BB + RSI]             │ │
│  │  ○ BB + RSI     │  │  - BB Period: 20        │ │
│  │                 │  │  - BB StdDev: 2         │ │
│  └─────────────────┘  └──────────────────────────┘ │
│                                                     │
│              ┌──────────────────┐                   │
│              │  백테스트 실행    │                   │
│              └──────────────────┘                   │
└─────────────────────────────────────────────────────┘
```

#### 입력 필드 상세

##### 1. 기본 설정 (왼쪽 상단 카드)

| Field | Type | Required | Default | Validation |
|-------|------|----------|---------|------------|
| **Symbol** | Select/Combobox | ✅ | BTCUSDT | Options: BTCUSDT, ETHUSDT, BNBUSDT, etc. |
| **Interval** | Select | ✅ | 1h | Options: 1m, 5m, 15m, 30m, 1h, 4h, 1d |
| **Start Date** | Date Picker | ✅ | 3 months ago | Must be < End Date |
| **End Date** | Date Picker | ✅ | Today | Must be > Start Date |
| **Initial Balance** | Number Input | ❌ | 10000 | Min: 0, Step: 1000, Unit: USDT |
| **Commission** | Number Input | ❌ | 0.001 | Min: 0, Max: 1, Step: 0.0001, Display: % |

##### 2. 전략 선택 (왼쪽 하단 카드)

**Type**: Radio Group (Card Style)

- ○ **MA Cross** (이동평균 교차)
- ○ **RSI** (상대강도지수)
- ○ **BB + RSI** (볼린저밴드 + RSI)

##### 3. 전략 파라미터 (오른쪽 카드)

동적으로 변경되는 입력 필드

###### MA Cross 선택 시

```typescript
{
  fast_period: number    // Default: 10, Min: 1, Max: 100
  slow_period: number    // Default: 30, Min: 1, Max: 200
                        // Validation: slow_period > fast_period
}
```

###### RSI 선택 시

```typescript
{
  rsi_period: number        // Default: 14, Min: 2, Max: 50
  rsi_oversold: number      // Default: 30, Min: 0, Max: 50
  rsi_overbought: number    // Default: 70, Min: 50, Max: 100
                           // Validation: overbought > oversold
}
```

###### BB + RSI 선택 시

```typescript
{
  bb_period: number         // Default: 20, Min: 2, Max: 100
  bb_std_dev: number        // Default: 2, Min: 0.5, Max: 5, Step: 0.1
  rsi_period: number        // Default: 14, Min: 2, Max: 50
  rsi_oversold: number      // Default: 30
  rsi_overbought: number    // Default: 70
}
```

##### 4. 포지션 사이즈 (선택)

```typescript
{
  position_size: number     // Default: 0.01 (1%)
                           // Min: 0.001, Max: 1, Step: 0.001
}
```

#### UX 흐름

1. 사용자가 페이지 진입
2. 기본값이 모두 채워진 상태 (즉시 실행 가능)
3. Symbol과 날짜만 선택하면 바로 실행 가능
4. 전략 선택 시 오른쪽 파라미터 카드가 부드럽게 전환
5. "백테스트 실행" 버튼 클릭
6. 로딩 스피너 표시 (예상 시간: 2-10초)
7. 완료 후 결과 페이지로 이동

#### 추가 기능 (Future)

- **프리셋 저장**: 자주 쓰는 설정 저장
- **빠른 설정**: "보수적", "공격적", "균형" 프리셋
- **데이터 확인**: 선택한 기간의 데이터 존재 여부 확인

---

### 페이지 2: 백테스트 결과 페이지

**Route**: `/backtest/result`

#### 레이아웃 구조

```
┌─────────────────────────────────────────────────────┐
│  Header: 백테스트 결과                                 │
│  [< 뒤로] BTCUSDT 1h MA Cross | 2025-07-01 ~ 10-17  │
├─────────────────────────────────────────────────────┤
│                                                     │
│  📊 성과 지표 (4개 카드 그리드)                        │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐              │
│  │수익률│ │Sharpe│ │ MDD  │ │승률  │              │
│  │+25%  │ │ 1.8  │ │-15%  │ │65%   │              │
│  └──────┘ └──────┘ └──────┘ └──────┘              │
│                                                     │
│  📈 자산 곡선 (Equity Curve)                         │
│  ┌─────────────────────────────────────────────┐   │
│  │                        ╱─╲                  │   │
│  │                  ╱────╱   ╲                 │   │
│  │            ╱────╱           ╲╱─╲            │   │
│  │      ╱────╱                      ╲          │   │
│  │ ────╱                             ╲──       │   │
│  └─────────────────────────────────────────────┘   │
│                                                     │
│  📋 거래 내역                                        │
│  ┌─────────────────────────────────────────────┐   │
│  │ Date      │ Type │ Price  │ Amount │ PnL   │   │
│  │ 10-01 ... │ BUY  │ 45000  │ 0.1    │ -     │   │
│  │ 10-05 ... │ SELL │ 47000  │ 0.1    │ +4.4% │   │
│  └─────────────────────────────────────────────┘   │
│                                                     │
│  📊 추가 분석 (Tabs)                                 │
│  [월별 수익] [드로우다운] [통계]                       │
│                                                     │
└─────────────────────────────────────────────────────┘
```

#### 표시할 데이터

##### 1. 성과 지표 카드 (상단)

**주요 지표 (4개 카드)**

```
┌─ 총 수익률 ─────┐  ┌─ Sharpe Ratio ─┐
│  +25.4%        │  │    1.85        │
│  ↑ $2,540      │  │    위험조정수익  │
└────────────────┘  └────────────────┘

┌─ Max Drawdown ─┐  ┌─ 승률 ─────────┐
│  -15.3%        │  │    65.2%       │
│  최대 낙폭      │  │    15승 8패     │
└────────────────┘  └────────────────┘
```

**추가 지표 (접기/펼치기 가능)**

- 총 거래 횟수: 23회
- 평균 수익: +4.2%
- 평균 손실: -2.8%
- Profit Factor: 1.85
- 최대 연속 승: 5회
- 최대 연속 패: 3회

##### 2. 자산 곡선 차트

- **Type**: Line Chart
- **X축**: 날짜/시간
- **Y축**: 계좌 잔액 (USDT)
- **기준선**: 초기 자금 표시
- **Interaction**:
  - Hover: 해당 시점의 잔액, 수익률 표시
  - Buy/Sell 마커 표시 (토글 가능)
  - Zoom/Pan 지원

##### 3. 거래 내역 테이블

**Columns**

| Column | Description | Sortable |
|--------|-------------|----------|
| Date & Time | 거래 체결 시각 | ✅ |
| Side | BUY/SELL | ❌ |
| Price | 체결 가격 | ✅ |
| Amount | 거래량 | ✅ |
| PnL | 개별 거래 손익 (%) | ✅ |
| Cumulative PnL | 누적 손익 ($) | ✅ |

**Features**
- 정렬: 날짜, 수익률 등
- 필터: Buy만, Sell만, 수익만, 손실만
- 페이지네이션: 20개씩
- Export: CSV 다운로드

##### 4. 추가 분석 (탭)

**Tab 1: 월별 수익**
- 히트맵 또는 바 차트
- 각 월의 수익률 표시
- 월간 평균 계산

**Tab 2: 드로우다운 차트**
- 시간에 따른 낙폭 표시
- 최대 낙폭 구간 하이라이트
- 회복 기간 표시

**Tab 3: 통계 요약**
- 거래 통계 (평균 보유 기간, 거래 빈도 등)
- 시간대별 분석
- 승/패 분포 히스토그램

#### 액션 버튼

```
[다른 파라미터로 재실행] [결과 저장] [PDF 다운로드]
```

---

## 🎨 디자인 시스템

### 색상 스키마

```css
/* Primary Colors */
--primary-gradient: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
--primary: #667eea;
--primary-dark: #764ba2;

/* Semantic Colors */
--success: #10b981;      /* 상승, 매수, 수익 */
--danger: #ef4444;       /* 하락, 매도, 손실 */
--warning: #f59e0b;      /* 경고 */
--info: #3b82f6;         /* 정보 */

/* Neutral Colors */
--background-light: #ffffff;
--background-dark: #1a1a1a;
--card-light: #ffffff;
--card-dark: #2d2d2d;
--text-primary: #1f2937;
--text-secondary: #6b7280;

/* Chart Colors */
--chart-line: #667eea;
--chart-grid: #e5e7eb;
--chart-buy: #10b981;
--chart-sell: #ef4444;
```

### Typography

```
- Heading 1: 3rem (48px), Bold
- Heading 2: 2rem (32px), SemiBold
- Heading 3: 1.5rem (24px), SemiBold
- Body: 1rem (16px), Regular
- Small: 0.875rem (14px), Regular
- Caption: 0.75rem (12px), Regular
```

### Spacing

```
- xs: 0.25rem (4px)
- sm: 0.5rem (8px)
- md: 1rem (16px)
- lg: 1.5rem (24px)
- xl: 2rem (32px)
- 2xl: 3rem (48px)
```

### 필요한 컴포넌트 (우선순위)

1. ⭐ Button (Primary, Secondary, Outline, Ghost)
2. ⭐ Card (Container with shadow)
3. ⭐ Input (Text, Number)
4. ⭐ Select / Combobox
5. ⭐ Label
6. Date Picker (Range)
7. Radio Group (Card style)
8. Table (Sortable, Filterable)
9. Tabs
10. Badge (수익률 표시)
11. Spinner / Loading
12. Toast (알림)

---

## 🔌 API 연동

### 1. 백테스트 실행

```http
POST /api/v1/backtest/run
Content-Type: application/json
```

**Request Body**

```typescript
interface BacktestRequest {
  symbol: string;           // "BTCUSDT"
  interval: string;         // "1h"
  start_date: string;       // "2025-07-01"
  end_date: string;         // "2025-10-17"
  initial_balance?: number; // 10000
  commission?: number;      // 0.001
  strategy: string;         // "ma_cross" | "rsi" | "bb_rsi"
  position_size?: number;   // 0.01
  
  // Strategy Parameters (conditional)
  fast_period?: number;     // MA Cross
  slow_period?: number;     // MA Cross
  rsi_period?: number;      // RSI, BB+RSI
  rsi_oversold?: number;    // RSI, BB+RSI
  rsi_overbought?: number;  // RSI, BB+RSI
  bb_period?: number;       // BB+RSI
  bb_std_dev?: number;      // BB+RSI
}
```

**Response**

```typescript
interface BacktestResponse {
  metrics: {
    total_return: number;           // 0.254 (25.4%)
    sharpe_ratio: number;           // 1.85
    max_drawdown: number;           // -0.153 (-15.3%)
    win_rate: number;               // 0.652 (65.2%)
    total_trades: number;           // 23
    winning_trades: number;         // 15
    losing_trades: number;          // 8
    avg_profit: number;             // 0.042 (4.2%)
    avg_loss: number;               // -0.028 (-2.8%)
    profit_factor: number;          // 1.85
    max_consecutive_wins: number;   // 5
    max_consecutive_losses: number; // 3
  };
  
  trades: Array<{
    timestamp: string;      // ISO 8601
    side: "BUY" | "SELL";
    price: number;
    amount: number;
    pnl?: number;           // Individual trade PnL
    cumulative_pnl: number;
  }>;
  
  equity_curve: Array<{
    timestamp: string;
    balance: number;
    return: number;         // % from initial
  }>;
}
```

### 2. 전략 목록 조회

```http
GET /api/v1/backtest/strategies
```

**Response**

```typescript
interface StrategiesResponse {
  strategies: Array<{
    name: string;              // "ma_cross"
    display_name: string;      // "MA Cross"
    description: string;       // "이동평균선 교차 전략"
    parameters: Array<{
      name: string;
      display_name: string;
      type: "number" | "string";
      default: any;
      min?: number;
      max?: number;
      step?: number;
    }>;
  }>;
}
```

---

## ✅ 구현 체크리스트

### Phase 1: 환경 설정 ⏳
- [ ] Shadcn-svelte 설치 및 설정
- [ ] Tailwind CSS 설정
- [ ] 라우팅 설정 (`/backtest/new`, `/backtest/result`)
- [ ] 레이아웃 컴포넌트 (Header, Sidebar)

### Phase 2: 백테스트 실행 페이지 🎯
- [ ] 기본 설정 폼 컴포넌트
  - [ ] Symbol Select
  - [ ] Interval Select
  - [ ] Date Range Picker
  - [ ] Number Inputs (Balance, Commission)
- [ ] 전략 선택 Radio Group
- [ ] 동적 파라미터 폼
  - [ ] MA Cross 파라미터
  - [ ] RSI 파라미터
  - [ ] BB+RSI 파라미터
- [ ] Form Validation
- [ ] API 연동
  - [ ] `/backtest/run` POST 요청
  - [ ] 로딩 상태 관리
  - [ ] 에러 핸들링
- [ ] 결과 페이지로 데이터 전달

### Phase 3: 결과 페이지 📊
- [ ] 성과 지표 카드 컴포넌트
  - [ ] 수익률 카드
  - [ ] Sharpe Ratio 카드
  - [ ] MDD 카드
  - [ ] 승률 카드
  - [ ] 추가 지표 (접기/펼치기)
- [ ] 자산 곡선 차트
  - [ ] Chart 라이브러리 선택 (Chart.js / Recharts)
  - [ ] Line Chart 구현
  - [ ] Buy/Sell 마커
  - [ ] Tooltip/Hover
- [ ] 거래 내역 테이블
  - [ ] Table 컴포넌트
  - [ ] Sorting
  - [ ] Filtering
  - [ ] Pagination
- [ ] 추가 분석 탭
  - [ ] Tab 컴포넌트
  - [ ] 월별 수익 차트
  - [ ] 드로우다운 차트
  - [ ] 통계 요약
- [ ] 액션 버튼
  - [ ] 재실행 버튼 (파라미터 유지)
  - [ ] PDF 다운로드 (Future)

### Phase 4: 최적화 & 개선 🚀
- [ ] 반응형 디자인 (모바일 대응)
- [ ] 다크 모드 지원
- [ ] 애니메이션 & 트랜지션
- [ ] 접근성 개선 (A11y)
- [ ] 성능 최적화
- [ ] 에러 바운더리
- [ ] 로딩 스켈레톤

---

## 📦 필요한 패키지

```json
{
  "dependencies": {
    "@sveltejs/kit": "^2.x",
    "svelte": "^5.x"
  },
  "devDependencies": {
    "tailwindcss": "^3.x",
    "autoprefixer": "^10.x",
    "postcss": "^8.x"
  },
  "추가 예정": {
    "chart.js": "차트 라이브러리",
    "date-fns": "날짜 처리",
    "zod": "Form validation"
  }
}
```

---

## 🎯 성공 기준

1. **사용성**: 5분 내에 첫 백테스트 실행 가능
2. **직관성**: 도움말 없이 모든 기능 사용 가능
3. **성능**: 백테스트 결과 로딩 < 3초
4. **반응성**: 모바일/태블릿에서도 사용 가능
5. **신뢰성**: 에러 상황에서도 명확한 피드백

---

## 📝 참고 사항

- 모든 금액은 USDT 기준
- 날짜/시간은 UTC 기준 (또는 사용자 로컬 시간대)
- 퍼센트는 소수점 2자리까지 표시
- 큰 숫자는 천 단위 구분 (45,000)
- 색상은 일관성 유지 (빨강=손실, 초록=수익)

---

## 🔄 버전 히스토리

- **v1.0** (2025-10-18): 초기 기획서 작성

