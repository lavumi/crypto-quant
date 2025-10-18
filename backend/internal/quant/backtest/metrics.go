package backtest

import (
	"fmt"
	"math"
	"time"
)

// Result holds backtest results and performance metrics
type Result struct {
	// Basic metrics
	InitialBalance float64
	FinalEquity    float64
	TotalReturn    float64 // Percentage return

	// Trade statistics
	TotalTrades   int
	WinningTrades int
	LosingTrades  int
	WinRate       float64

	// Performance metrics
	SharpeRatio    float64
	MaxDrawdown    float64
	MaxDrawdownPct float64

	// Time metrics
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration

	// Detailed data
	Trades      []*Trade
	EquityCurve []EquityPoint
}

// calculateResult computes backtest results and metrics
func (e *Engine) calculateResult() *Result {
	if len(e.equity) == 0 {
		return &Result{
			InitialBalance: e.initialBalance,
			FinalEquity:    e.balance,
			TotalReturn:    0,
		}
	}

	finalEquity := e.equity[len(e.equity)-1].Equity
	totalReturn := (finalEquity - e.initialBalance) / e.initialBalance

	result := &Result{
		InitialBalance: e.initialBalance,
		FinalEquity:    finalEquity,
		TotalReturn:    totalReturn,
		TotalTrades:    len(e.trades),
		StartTime:      e.equity[0].Timestamp,
		EndTime:        e.equity[len(e.equity)-1].Timestamp,
		Trades:         e.trades,
		EquityCurve:    e.equity,
	}

	result.Duration = result.EndTime.Sub(result.StartTime)

	// Calculate trade statistics
	result.calculateTradeStats()

	// Calculate Sharpe Ratio
	result.SharpeRatio = result.calculateSharpeRatio()

	// Calculate Maximum Drawdown
	result.MaxDrawdown, result.MaxDrawdownPct = result.calculateMaxDrawdown()

	return result
}

// calculateTradeStats calculates trade statistics
func (r *Result) calculateTradeStats() {
	if len(r.Trades) < 2 {
		return
	}

	// Group trades into complete round trips (buy-sell pairs)
	var buyPrice float64
	var buyQty float64

	for _, trade := range r.Trades {
		if trade.Side == "BUY" {
			buyPrice = trade.Price
			buyQty = trade.Quantity
		} else if trade.Side == "SELL" && buyPrice > 0 {
			// Calculate P&L for this round trip
			pnl := (trade.Price - buyPrice) * buyQty

			if pnl > 0 {
				r.WinningTrades++
			} else if pnl < 0 {
				r.LosingTrades++
			}

			buyPrice = 0
			buyQty = 0
		}
	}

	totalCompletedTrades := r.WinningTrades + r.LosingTrades
	if totalCompletedTrades > 0 {
		r.WinRate = float64(r.WinningTrades) / float64(totalCompletedTrades)
	}
}

// calculateSharpeRatio calculates the Sharpe ratio
// Sharpe Ratio = (Average Return - Risk Free Rate) / Standard Deviation of Returns
// Assuming risk-free rate = 0 for simplicity
func (r *Result) calculateSharpeRatio() float64 {
	if len(r.EquityCurve) < 2 {
		return 0
	}

	// Calculate daily returns
	returns := make([]float64, 0, len(r.EquityCurve)-1)
	for i := 1; i < len(r.EquityCurve); i++ {
		prevEquity := r.EquityCurve[i-1].Equity
		currEquity := r.EquityCurve[i].Equity

		if prevEquity > 0 {
			ret := (currEquity - prevEquity) / prevEquity
			returns = append(returns, ret)
		}
	}

	if len(returns) == 0 {
		return 0
	}

	// Calculate mean return
	var sumReturns float64
	for _, ret := range returns {
		sumReturns += ret
	}
	meanReturn := sumReturns / float64(len(returns))

	// Calculate standard deviation
	var sumSquaredDiff float64
	for _, ret := range returns {
		diff := ret - meanReturn
		sumSquaredDiff += diff * diff
	}
	stdDev := math.Sqrt(sumSquaredDiff / float64(len(returns)))

	if stdDev == 0 {
		return 0
	}

	// Annualize Sharpe Ratio (assuming returns are in the same frequency as data)
	// For simplicity, we'll use sqrt(252) for daily data (252 trading days per year)
	sharpe := meanReturn / stdDev

	// Estimate annualization factor based on data frequency
	totalDays := r.Duration.Hours() / 24
	periodsPerYear := 365 / (totalDays / float64(len(returns)))
	annualizationFactor := math.Sqrt(periodsPerYear)

	return sharpe * annualizationFactor
}

// calculateMaxDrawdown calculates maximum drawdown
// Maximum Drawdown = (Peak - Trough) / Peak
func (r *Result) calculateMaxDrawdown() (float64, float64) {
	if len(r.EquityCurve) == 0 {
		return 0, 0
	}

	var maxDrawdown float64
	var maxDrawdownPct float64
	peak := r.EquityCurve[0].Equity

	for _, point := range r.EquityCurve {
		if point.Equity > peak {
			peak = point.Equity
		}

		drawdown := peak - point.Equity
		drawdownPct := drawdown / peak

		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
			maxDrawdownPct = drawdownPct
		}
	}

	return maxDrawdown, maxDrawdownPct
}

// Print prints the backtest results
func (r *Result) Print() {
	println("\n========== Backtest Results ==========")
	println("Time Period:")
	println("  Start:", r.StartTime.Format("2006-01-02 15:04:05"))
	println("  End:  ", r.EndTime.Format("2006-01-02 15:04:05"))
	println("  Duration:", r.Duration.String())
	println()
	println("Financial Performance:")
	println("  Initial Balance:", formatMoney(r.InitialBalance))
	println("  Final Equity:   ", formatMoney(r.FinalEquity))
	println("  Total Return:   ", formatPercent(r.TotalReturn))
	println()
	println("Risk Metrics:")
	println("  Sharpe Ratio:   ", formatFloat(r.SharpeRatio, 2))
	println("  Max Drawdown:   ", formatMoney(r.MaxDrawdown), formatPercent(r.MaxDrawdownPct))
	println()
	println("Trade Statistics:")
	println("  Total Trades:   ", r.TotalTrades)
	println("  Winning Trades: ", r.WinningTrades)
	println("  Losing Trades:  ", r.LosingTrades)
	println("  Win Rate:       ", formatPercent(r.WinRate))
	println("======================================\n")
}

func formatMoney(value float64) string {
	return fmt.Sprintf("$%.2f", value)
}

func formatPercent(value float64) string {
	return fmt.Sprintf("%.2f%%", value*100)
}

func formatFloat(value float64, decimals int) string {
	format := fmt.Sprintf("%%.%df", decimals)
	return fmt.Sprintf(format, value)
}
