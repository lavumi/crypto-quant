# 백테스트 API 가이드

## 개요

백테스트 기능을 REST API로 사용할 수 있습니다. 웹 인터페이스나 외부 도구에서 백테스트를 실행할 수 있습니다.

## API 엔드포인트

### 1. 사용 가능한 전략 조회

**GET** `/api/v1/backtest/strategies`

사용 가능한 모든 백테스트 전략과 파라미터 정보를 조회합니다.

#### 예시

```bash
curl http://localhost:8080/api/v1/backtest/strategies
```

#### 응답

```json
{
  "success": true,
  "data": {
    "ma_cross": {
      "name": "Moving Average Crossover",
      "description": "Buy when fast MA crosses above slow MA, sell when it crosses below",
      "parameters": {
        "fast_period": {
          "type": "integer",
          "default": 10,
          "description": "Fast moving average period"
        },
        "slow_period": {
          "type": "integer",
          "default": 30,
          "description": "Slow moving average period"
        }
      }
    },
    "rsi": {
      "name": "RSI Strategy",
      "description": "Buy when RSI is oversold, sell when overbought",
      "parameters": {
        "rsi_period": {
          "type": "integer",
          "default": 14,
          "description": "RSI calculation period"
        },
        "rsi_oversold": {
          "type": "float",
          "default": 30,
          "description": "Oversold threshold (buy signal)"
        },
        "rsi_overbought": {
          "type": "float",
          "default": 70,
          "description": "Overbought threshold (sell signal)"
        },
        "position_size": {
          "type": "float",
          "default": 0.01,
          "description": "Position size per trade"
        }
      }
    },
    "bb_rsi": {
      "name": "Bollinger Bands + RSI",
      "description": "Combined strategy using BB and RSI confirmation",
      "parameters": {
        "rsi_period": {
          "type": "integer",
          "default": 14,
          "description": "RSI calculation period"
        },
        "rsi_oversold": {
          "type": "float",
          "default": 30,
          "description": "Oversold threshold"
        },
        "rsi_overbought": {
          "type": "float",
          "default": 70,
          "description": "Overbought threshold"
        },
        "position_size": {
          "type": "float",
          "default": 0.01,
          "description": "Position size per trade"
        }
      }
    }
  }
}
```

---

### 2. 백테스트 실행

**POST** `/api/v1/backtest/run`

지정한 파라미터로 백테스트를 실행합니다.

#### 요청 Body

```json
{
  "symbol": "BTCUSDT",
  "interval": "1h",
  "start_date": "2025-07-19",
  "end_date": "2025-10-17",
  "initial_balance": 10000,
  "commission": 0.001,
  "strategy": "ma_cross",
  "fast_period": 10,
  "slow_period": 30
}
```

#### 파라미터 설명

| 필드 | 필수 | 타입 | 기본값 | 설명 |
|------|------|------|--------|------|
| `symbol` | ✓ | string | - | 거래 페어 (예: BTCUSDT, ETHUSDT) |
| `interval` | ✓ | string | - | 캔들 간격 (1m, 5m, 15m, 1h, 4h, 1d) |
| `start_date` | ✓ | string | - | 시작 날짜 (YYYY-MM-DD) |
| `end_date` | ✓ | string | - | 종료 날짜 (YYYY-MM-DD) |
| `initial_balance` | | float | 10000 | 초기 자금 |
| `commission` | | float | 0.001 | 수수료율 (0.001 = 0.1%) |
| `strategy` | | string | "ma_cross" | 전략 (ma_cross, rsi, bb_rsi) |
| `fast_period` | | int | 10 | 빠른 MA 기간 (MA Cross용) |
| `slow_period` | | int | 30 | 느린 MA 기간 (MA Cross용) |
| `rsi_period` | | int | 14 | RSI 기간 (RSI, BB+RSI용) |
| `rsi_oversold` | | float | 30 | 과매도 임계값 (RSI, BB+RSI용) |
| `rsi_overbought` | | float | 70 | 과매수 임계값 (RSI, BB+RSI용) |
| `position_size` | | float | 0.01 | 포지션 크기 (RSI, BB+RSI용) |

#### 예시

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

#### 응답

```json
{
  "success": true,
  "data": {
    "start_time": "2025-07-19T17:00:00+09:00",
    "end_time": "2025-10-17T08:00:00+09:00",
    "duration": "2151h0m0s",
    "initial_balance": 10000,
    "final_equity": 9866.65,
    "total_return": -1.33,
    "total_return_pct": "-1.33%",
    "sharpe_ratio": -2.16,
    "max_drawdown": 206.59,
    "max_drawdown_pct": 2.07,
    "total_trades": 82,
    "winning_trades": 12,
    "losing_trades": 29,
    "win_rate": 29.27,
    "win_rate_pct": "29.27%",
    "strategy": "MA_Cross_10_30",
    "symbol": "BTCUSDT",
    "interval": "1h",
    "commission": 0.001,
    "candles_used": 2152,
    "recent_trades": [
      {
        "timestamp": "2025-07-21T16:00:00+09:00",
        "side": "BUY",
        "price": 119362.67,
        "quantity": 0.01,
        "fee": 1.19,
        "balance": 8805.18,
        "position": 0.01,
        "reason": "Golden Cross: Fast MA (118128.46) > Slow MA (118117.20)"
      },
      {
        "timestamp": "2025-07-22T03:00:00+09:00",
        "side": "SELL",
        "price": 117162.28,
        "quantity": 0.01,
        "fee": 1.17,
        "balance": 9975.63,
        "position": 0,
        "reason": "Death Cross: Fast MA (118223.59) < Slow MA (118242.43)"
      }
      // ... 최대 20개 거래까지 표시
    ]
  }
}
```

#### 응답 필드 설명

| 필드 | 설명 |
|------|------|
| `start_time` | 백테스트 시작 시간 |
| `end_time` | 백테스트 종료 시간 |
| `duration` | 백테스트 기간 |
| `initial_balance` | 초기 자금 |
| `final_equity` | 최종 자산 (잔고 + 포지션 가치) |
| `total_return` | 총 수익률 (%) |
| `sharpe_ratio` | 샤프 비율 (위험 조정 수익률) |
| `max_drawdown` | 최대 낙폭 (달러) |
| `max_drawdown_pct` | 최대 낙폭 (%) |
| `total_trades` | 총 거래 횟수 |
| `winning_trades` | 수익 거래 횟수 |
| `losing_trades` | 손실 거래 횟수 |
| `win_rate` | 승률 (%) |
| `strategy` | 사용된 전략 이름 |
| `candles_used` | 사용된 캔들 수 |
| `recent_trades` | 최근 거래 내역 (최대 20개) |

---

## 사용 예시

### 1. MA Crossover 전략

```bash
curl -X POST http://localhost:8080/api/v1/backtest/run \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTCUSDT",
    "interval": "1h",
    "start_date": "2025-07-19",
    "end_date": "2025-10-17",
    "strategy": "ma_cross",
    "fast_period": 10,
    "slow_period": 30
  }'
```

### 2. RSI 전략

```bash
curl -X POST http://localhost:8080/api/v1/backtest/run \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "ETHUSDT",
    "interval": "4h",
    "start_date": "2025-07-01",
    "end_date": "2025-10-17",
    "strategy": "rsi",
    "rsi_period": 14,
    "rsi_oversold": 30,
    "rsi_overbought": 70,
    "position_size": 0.1
  }'
```

### 3. BB + RSI 전략

```bash
curl -X POST http://localhost:8080/api/v1/backtest/run \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTCUSDT",
    "interval": "1d",
    "start_date": "2025-01-01",
    "end_date": "2025-10-17",
    "strategy": "bb_rsi",
    "rsi_period": 14,
    "rsi_oversold": 25,
    "rsi_overbought": 75,
    "position_size": 0.01
  }'
```

### 4. Python으로 사용

```python
import requests
import json

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
else:
    print(f"Error: {result['error']['message']}")
```

### 5. JavaScript (Node.js)로 사용

```javascript
const axios = require('axios');

async function runBacktest() {
  try {
    const response = await axios.post('http://localhost:8080/api/v1/backtest/run', {
      symbol: 'BTCUSDT',
      interval: '1h',
      start_date: '2025-07-19',
      end_date: '2025-10-17',
      strategy: 'ma_cross',
      fast_period: 10,
      slow_period: 30
    });

    const data = response.data.data;
    console.log(`Total Return: ${data.total_return_pct}`);
    console.log(`Sharpe Ratio: ${data.sharpe_ratio.toFixed(2)}`);
    console.log(`Win Rate: ${data.win_rate_pct}`);
    console.log(`Total Trades: ${data.total_trades}`);
  } catch (error) {
    console.error('Error:', error.response.data.error.message);
  }
}

runBacktest();
```

---

## 파라미터 최적화 예시

여러 파라미터 조합을 테스트하여 최적의 전략을 찾을 수 있습니다:

### Bash 스크립트

```bash
#!/bin/bash

# 여러 MA 조합 테스트
for fast in 5 10 15 20; do
  for slow in 20 30 40 50; do
    if [ $fast -lt $slow ]; then
      echo "Testing MA($fast/$slow)..."
      curl -s -X POST http://localhost:8080/api/v1/backtest/run \
        -H "Content-Type: application/json" \
        -d "{
          \"symbol\": \"BTCUSDT\",
          \"interval\": \"1h\",
          \"start_date\": \"2025-07-19\",
          \"end_date\": \"2025-10-17\",
          \"strategy\": \"ma_cross\",
          \"fast_period\": $fast,
          \"slow_period\": $slow
        }" | jq -r '.data | "MA(\($fast)/\($slow)): Return=\(.total_return_pct), Sharpe=\(.sharpe_ratio | tostring | .[0:5]), Trades=\(.total_trades)"'
    fi
  done
done
```

### Python 스크립트

```python
import requests
import pandas as pd

url = "http://localhost:8080/api/v1/backtest/run"
results = []

for fast in [5, 10, 15, 20]:
    for slow in [20, 30, 40, 50]:
        if fast < slow:
            payload = {
                "symbol": "BTCUSDT",
                "interval": "1h",
                "start_date": "2025-07-19",
                "end_date": "2025-10-17",
                "strategy": "ma_cross",
                "fast_period": fast,
                "slow_period": slow
            }
            
            response = requests.post(url, json=payload)
            if response.json()["success"]:
                data = response.json()["data"]
                results.append({
                    "Fast MA": fast,
                    "Slow MA": slow,
                    "Return (%)": round(data["total_return"], 2),
                    "Sharpe": round(data["sharpe_ratio"], 2),
                    "Trades": data["total_trades"],
                    "Win Rate (%)": round(data["win_rate"], 2)
                })

# 결과를 DataFrame으로 정리
df = pd.DataFrame(results)
df_sorted = df.sort_values("Return (%)", ascending=False)
print(df_sorted.to_string(index=False))
```

---

## 에러 처리

API는 표준 HTTP 상태 코드와 에러 메시지를 반환합니다:

### 400 Bad Request

```json
{
  "success": false,
  "error": {
    "code": "BAD_REQUEST",
    "message": "Invalid request: start_date is required"
  }
}
```

### 404 Not Found

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "No candles found for the specified period. Please collect data first."
  }
}
```

### 500 Internal Server Error

```json
{
  "success": false,
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Backtest failed: strategy error"
  }
}
```

---

## 주의사항

1. **데이터 수집 먼저**: 백테스트를 실행하기 전에 먼저 collector로 데이터를 수집해야 합니다.
   ```bash
   ./bin/collector -symbol BTCUSDT -interval 1h -days 90
   ```

2. **처리 시간**: 많은 캔들 데이터로 백테스트할 경우 시간이 걸릴 수 있습니다.

3. **날짜 형식**: 날짜는 반드시 `YYYY-MM-DD` 형식을 사용해야 합니다.

4. **전략 파라미터**: 각 전략에 맞는 파라미터를 사용하세요.
   - MA Cross: `fast_period`, `slow_period`
   - RSI: `rsi_period`, `rsi_oversold`, `rsi_overbought`, `position_size`
   - BB+RSI: `rsi_period`, `rsi_oversold`, `rsi_overbought`, `position_size`

---

## Swagger 문서

더 자세한 API 문서는 Swagger UI에서 확인할 수 있습니다:

```
http://localhost:8080/swagger/index.html
```

---

## 다음 단계

- [전체 사용 가이드](./USAGE_GUIDE.md)
- [백테스팅 상세 가이드](./BACKTEST.md)
- [빠른 시작](../QUICKSTART.md)







