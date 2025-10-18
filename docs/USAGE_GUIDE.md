# 사용 가이드 (Usage Guide)

## 목차
1. [데이터 수집](#1-데이터-수집)
2. [백테스트 실행 (CLI)](#2-백테스트-실행-cli)
3. [백테스트 실행 (API)](#3-백테스트-실행-api)
4. [API 서버 실행](#4-api-서버-실행)
5. [자주 사용하는 명령어](#5-자주-사용하는-명령어)

---

## 1. 데이터 수집

### 기본 사용법

```bash
./bin/collector -symbol BTCUSDT -interval 1h -days 90
```

### 파라미터 설명

| 파라미터 | 설명 | 기본값 | 예시 |
|---------|------|--------|------|
| `-symbol` | 거래 페어 | BTCUSDT | BTCUSDT, ETHUSDT, BNBUSDT |
| `-interval` | 캔들 간격 | 1h | 1m, 5m, 15m, 1h, 4h, 1d |
| `-days` | 수집 일수 (과거부터) | 30 | 90, 180, 365 |
| `-db` | 데이터베이스 경로 | data/trading.db | data/custom.db |

### 추천 데이터 수집 명령어

```bash
# BTC 데이터 수집
./bin/collector -symbol BTCUSDT -interval 1h -days 90   # 1시간봉 3개월
./bin/collector -symbol BTCUSDT -interval 4h -days 180  # 4시간봉 6개월
./bin/collector -symbol BTCUSDT -interval 1d -days 365  # 일봉 1년

# ETH 데이터 수집
./bin/collector -symbol ETHUSDT -interval 1h -days 90
./bin/collector -symbol ETHUSDT -interval 1d -days 365

# BNB 데이터 수집
./bin/collector -symbol BNBUSDT -interval 1h -days 90
./bin/collector -symbol BNBUSDT -interval 1d -days 365
```

### 출력 예시

```
=== Historical Data Collector ===
Symbol: BTCUSDT
Interval: 1h
Days: 90
Database: data/trading.db
Database migrations completed successfully
Collecting historical data for BTCUSDT (1h) from 2025-07-19 to 2025-10-17
Saved 1000 candles (total: 1000)
Saved 1000 candles (total: 2000)
Saved 159 candles (total: 2159)
Historical data collection completed successfully!
```

### 주의사항

- 데이터 수집은 Binance API를 사용합니다 (API 키 불필요)
- 너무 많은 요청을 보내면 rate limit에 걸릴 수 있습니다
- 한 번에 1000개 캔들씩 수집됩니다
- 이미 존재하는 데이터는 업데이트됩니다 (중복 방지)

---

## 2. 백테스트 실행 (CLI)

### 기본 사용법

```bash
./bin/backtest -symbol BTCUSDT -interval 1h -start 2024-07-01 -end 2024-10-17
```

### 파라미터 설명

| 파라미터 | 설명 | 기본값 | 예시 |
|---------|------|--------|------|
| `-symbol` | 거래 페어 | BTCUSDT | BTCUSDT, ETHUSDT |
| `-interval` | 캔들 간격 | 1h | 1m, 5m, 15m, 1h, 4h, 1d |
| `-start` | 시작 날짜 | 3개월 전 | 2024-01-01 |
| `-end` | 종료 날짜 | 현재 | 2024-10-17 |
| `-balance` | 초기 자금 | 10000 | 10000, 50000, 100000 |
| `-commission` | 수수료율 | 0.001 | 0.001 (0.1%), 0.002 (0.2%) |
| `-fast` | 빠른 MA 기간 | 10 | 5, 7, 10, 20 |
| `-slow` | 느린 MA 기간 | 30 | 20, 30, 50, 200 |

### 다양한 백테스트 시나리오

#### 1. 기본 백테스트
```bash
# 기본 설정으로 빠르게 테스트
./bin/backtest
```

#### 2. 단기 트레이딩 전략
```bash
# 빠른 이동평균으로 자주 거래
./bin/backtest -symbol BTCUSDT -interval 1h -fast 5 -slow 15 -start 2024-08-01
```

#### 3. 장기 투자 전략
```bash
# 느린 이동평균으로 큰 추세 포착
./bin/backtest -symbol BTCUSDT -interval 1d -fast 20 -slow 50 -start 2024-01-01
```

#### 4. 골든 크로스 전략
```bash
# 유명한 50/200 골든 크로스
./bin/backtest -symbol BTCUSDT -interval 1d -fast 50 -slow 200 -start 2024-01-01
```

#### 5. 알트코인 백테스트
```bash
# ETH로 테스트
./bin/backtest -symbol ETHUSDT -interval 4h -fast 10 -slow 30
```

#### 6. 높은 수수료 환경 테스트
```bash
# 현실적인 높은 수수료로 테스트
./bin/backtest -commission 0.002 -balance 10000
```

#### 7. 큰 자금으로 테스트
```bash
# 10만 달러로 시뮬레이션
./bin/backtest -balance 100000 -symbol BTCUSDT -interval 1d
```

### 결과 해석

```
========== Backtest Results ==========
Time Period:
  Start: 2024-07-19 09:00:00
  End:   2024-10-16 23:00:00
  Duration: 2158h0m0s

Financial Performance:
  Initial Balance: $10000.00    ← 초기 자금
  Final Equity:    $12450.75     ← 최종 자산 (잔고 + 포지션 가치)
  Total Return:    24.51%        ← 총 수익률

Risk Metrics:
  Sharpe Ratio:    1.85          ← 위험 대비 수익 (>1 좋음, >2 매우 좋음)
  Max Drawdown:    $850.25 (8.50%) ← 최대 손실폭 (작을수록 좋음)

Trade Statistics:
  Total Trades:    45            ← 총 거래 횟수
  Winning Trades:  28            ← 수익 거래
  Losing Trades:   17            ← 손실 거래
  Win Rate:        62.22%        ← 승률
======================================
```

### 성능 지표 설명

#### 📊 Total Return (총 수익률)
- 초기 자금 대비 수익률
- **좋은 기준**: 연 10% 이상
- 예: 24.51% = $10,000 → $12,451

#### 📈 Sharpe Ratio (샤프 비율)
- 위험(변동성) 대비 수익률
- **평가 기준**:
  - < 0: 나쁨
  - 0-1: 평범
  - 1-2: 좋음
  - \> 2: 매우 우수
- 예: 1.85 = 좋은 위험 대비 수익

#### 📉 Maximum Drawdown (최대 낙폭)
- 최고점에서 최저점까지 최대 하락폭
- **좋은 기준**: 10% 이하
- 예: 8.50% = 심리적으로 견딜 만한 수준

#### 🎯 Win Rate (승률)
- 수익 거래 비율
- **주의**: 높은 승률 ≠ 높은 수익
- 평균 수익/손실 크기도 함께 봐야 함
- 예: 62.22% = 100번 중 62번 수익

---

## 3. 백테스트 실행 (API)

### 서버 시작

```bash
./bin/api --port 8080
```

### API로 백테스트 실행

```bash
curl -X POST http://localhost:8080/api/v1/backtest/run \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTCUSDT",
    "interval": "1h",
    "start_date": "2025-07-19",
    "end_date": "2025-10-17",
    "initial_balance": 10000,
    "commission": 0.001,
    "strategy": "ma_cross",
    "fast_period": 10,
    "slow_period": 30
  }'
```

### 사용 가능한 전략 조회

```bash
curl http://localhost:8080/api/v1/backtest/strategies
```

### Python으로 사용

```python
import requests

url = "http://localhost:8080/api/v1/backtest/run"
payload = {
    "symbol": "BTCUSDT",
    "interval": "1h",
    "start_date": "2025-07-19",
    "end_date": "2025-10-17",
    "strategy": "ma_cross",
    "fast_period": 10,
    "slow_period": 30
}

response = requests.post(url, json=payload)
result = response.json()

if result["success"]:
    data = result["data"]
    print(f"Total Return: {data['total_return_pct']}")
    print(f"Sharpe Ratio: {data['sharpe_ratio']:.2f}")
    print(f"Win Rate: {data['win_rate_pct']}")
    print(f"Total Trades: {data['total_trades']}")
```

> 📖 **상세 가이드**: [백테스트 API 가이드](./API_BACKTEST.md) 참고

---

## 4. API 서버 실행

### 기본 사용법

```bash
./bin/api --port 8080 --db data/trading.db
```

### 파라미터 설명

| 파라미터 | 설명 | 기본값 |
|---------|------|--------|
| `--port` | 서버 포트 | 8080 |
| `--db` | 데이터베이스 경로 | data/trading.db |
| `--api-key` | Binance API 키 | - |
| `--secret-key` | Binance Secret 키 | - |
| `--testnet` | 테스트넷 사용 | false |

### API 엔드포인트

서버 실행 후 `http://localhost:8080/swagger/index.html` 에서 API 문서 확인 가능

주요 엔드포인트:
- `GET /api/v1/market/price/{symbol}` - 현재 가격 조회
- `GET /api/v1/market/stream/{symbol}` - 실시간 가격 스트림 (SSE)
- `GET /api/v1/data/candles` - 캔들 데이터 조회
- `GET /api/v1/wallet/balance` - 잔고 조회
- `POST /api/v1/order` - 주문 생성

---

## 5. 자주 사용하는 명령어

### 빌드 관련

```bash
# 전체 빌드
make build

# 개별 빌드
make build-collector
make build-backtest
make build-api

# 빌드 파일 삭제
make clean
```

### 개발 모드 실행

```bash
# 빌드 없이 바로 실행 (개발용)
make dev-collector
make dev-backtest
make dev-api
```

### 데이터베이스 확인

```bash
# SQLite CLI로 데이터 확인
sqlite3 data/trading.db

# 테이블 목록
.tables

# 캔들 데이터 조회
SELECT symbol, interval, COUNT(*) as count 
FROM candles 
GROUP BY symbol, interval;

# 최근 캔들 조회
SELECT * FROM candles 
WHERE symbol='BTCUSDT' AND interval='1h' 
ORDER BY open_time DESC 
LIMIT 10;

# 종료
.quit
```

---

## 5. 일반적인 워크플로우

### 새 프로젝트 시작 시

```bash
# 1. 의존성 설치
make deps

# 2. 빌드
make build

# 3. 데이터 수집
./bin/collector -symbol BTCUSDT -interval 1h -days 90
./bin/collector -symbol BTCUSDT -interval 1d -days 365

# 4. 백테스트 실행
./bin/backtest -symbol BTCUSDT -interval 1h -start 2024-07-01

# 5. 결과 분석 및 전략 조정
```

### 전략 파라미터 최적화

```bash
# 여러 파라미터 조합 테스트
for fast in 5 10 15 20; do
  for slow in 20 30 40 50; do
    echo "Testing fast=$fast slow=$slow"
    ./bin/backtest -fast $fast -slow $slow -symbol BTCUSDT -interval 1h
  done
done
```

### 여러 심볼 테스트

```bash
# 스크립트로 자동화
for symbol in BTCUSDT ETHUSDT BNBUSDT; do
  echo "=== Testing $symbol ==="
  ./bin/backtest -symbol $symbol -interval 1d -start 2024-01-01
done
```

---

## 6. 트러블슈팅

### 문제: 데이터가 없다고 나옴

**해결책**: 먼저 collector로 데이터 수집
```bash
./bin/collector -symbol BTCUSDT -interval 1h -days 90
```

### 문제: API 에러 발생

**해결책**: Binance API rate limit일 수 있음. 잠시 후 재시도
```bash
# 10초 대기 후 재시도
sleep 10 && ./bin/collector -symbol BTCUSDT -interval 1h -days 30
```

### 문제: 빌드 실패

**해결책**: 의존성 재설치
```bash
go mod tidy
go mod download
make build
```

### 문제: 데이터베이스 락 에러

**해결책**: 다른 프로세스가 DB를 사용 중인지 확인
```bash
# 실행 중인 프로세스 확인
ps aux | grep crypto-quant

# 필요시 프로세스 종료
killall api collector backtest
```

---

## 7. 팁과 베스트 프랙티스

### 💡 데이터 수집 팁

1. **다양한 타임프레임 수집**
   ```bash
   # 같은 심볼의 여러 타임프레임
   ./bin/collector -symbol BTCUSDT -interval 1h -days 90
   ./bin/collector -symbol BTCUSDT -interval 4h -days 180
   ./bin/collector -symbol BTCUSDT -interval 1d -days 365
   ```

2. **정기적으로 데이터 업데이트**
   ```bash
   # cron job으로 매일 실행
   0 0 * * * /path/to/bin/collector -symbol BTCUSDT -interval 1d -days 7
   ```

### 💡 백테스트 팁

1. **아웃오브샘플 테스트**
   ```bash
   # 학습 기간
   ./bin/backtest -start 2024-01-01 -end 2024-06-30 -fast 10 -slow 30
   
   # 검증 기간 (파라미터 동일하게)
   ./bin/backtest -start 2024-07-01 -end 2024-10-17 -fast 10 -slow 30
   ```

2. **현실적인 조건 설정**
   ```bash
   # 실제 거래 수수료 적용
   ./bin/backtest -commission 0.001  # Binance 메이커 수수료
   ./bin/backtest -commission 0.002  # 보수적 추정
   ```

3. **여러 시장 상황 테스트**
   ```bash
   # 상승장
   ./bin/backtest -start 2024-01-01 -end 2024-03-31
   
   # 하락장
   ./bin/backtest -start 2024-04-01 -end 2024-06-30
   
   # 횡보장
   ./bin/backtest -start 2024-07-01 -end 2024-09-30
   ```

### 💡 성능 최적화

1. **큰 데이터셋은 높은 타임프레임 사용**
   ```bash
   # 빠름: 일봉 1년
   ./bin/backtest -interval 1d -days 365
   
   # 느림: 1분봉 1년
   # ./bin/backtest -interval 1m -days 365  # 피하기
   ```

2. **필요한 만큼만 데이터 수집**
   ```bash
   # 전략 개발: 짧은 기간
   ./bin/collector -days 30
   
   # 최종 검증: 긴 기간
   ./bin/collector -days 365
   ```

---

## 8. 추가 리소스

- [백테스트 상세 가이드](./BACKTEST.md)
- [Phase 2 완료 요약](./PHASE2_SUMMARY.md)
- [변경사항 로그](../CHANGELOG.md)
- [프로젝트 README](../README.md)

---

**마지막 업데이트**: 2024-10-17

