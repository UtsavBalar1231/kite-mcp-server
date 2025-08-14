package mcp

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// TechnicalIndicators holds all calculated technical analysis values
type TechnicalIndicators struct {
	// Price Action
	Support           []float64
	Resistance        []float64
	Trend             string // "bullish", "bearish", "neutral"
	TrendStrength     float64
	
	// Moving Averages
	SMA20             float64
	SMA50             float64
	SMA200            float64
	EMA9              float64
	EMA21             float64
	VWAP              float64
	
	// Momentum Indicators
	RSI               float64
	RSIDivergence     bool
	MACD              MACDValues
	Stochastic        StochasticValues
	
	// Volatility
	BollingerBands    BollingerValues
	ATR               float64
	VolumeProfile     VolumeProfileData
	
	// Patterns
	CandlePattern     string
	ChartPattern      string
	
	// Strength Scores
	BullishScore      float64 // 0-100
	BearishScore      float64 // 0-100
}

type MACDValues struct {
	MACD      float64
	Signal    float64
	Histogram float64
	Crossover string // "bullish", "bearish", "none"
}

type StochasticValues struct {
	K float64
	D float64
	Oversold bool
	Overbought bool
}

type BollingerValues struct {
	Upper  float64
	Middle float64
	Lower  float64
	Width  float64
}

type VolumeProfileData struct {
	POC              float64 // Point of Control
	VAH              float64 // Value Area High
	VAL              float64 // Value Area Low
	VolumeIncrease   bool
	AccumulationDist float64
}

// MarketAnalysis contains comprehensive market analysis
type MarketAnalysis struct {
	Symbol           string
	Technical        TechnicalIndicators
	Fundamental      FundamentalData
	Sentiment        SentimentData
	RiskReward       RiskRewardAnalysis
	TradeSignal      TradeSignal
	Confidence       float64 // 0-100
	TimeAnalyzed     time.Time
}

type FundamentalData struct {
	PE               float64
	PB               float64
	DebtToEquity     float64
	ROE              float64
	QuarterlyGrowth  float64
	IndustryPE       float64
	RelativeStrength float64
	FundamentalScore float64 // 0-100
}

type SentimentData struct {
	DeliveryPercent  float64
	BulkDeals        int
	FIIActivity      string // "buying", "selling", "neutral"
	OptionsPCR       float64
	OpenInterest     float64
	SentimentScore   float64 // 0-100
}

type RiskRewardAnalysis struct {
	EntryPrice       float64
	StopLoss         float64
	Target1          float64
	Target2          float64
	Target3          float64
	RiskAmount       float64
	RewardAmount     float64
	RiskRewardRatio  float64
	PositionSize     int
	MaxLoss          float64
	MaxProfit        float64
}

type TradeSignal struct {
	Action           string // "BUY", "SELL", "HOLD"
	Strength         string // "strong", "moderate", "weak"
	Timeframe        string // "intraday", "swing", "positional"
	Strategy         string // Description of strategy
	Reasons          []string
	Warnings         []string
	ExpectedReturn   float64
	HoldingPeriod    string
	Priority         int // 1-10, higher is better
}

// CalculateTechnicalIndicators performs comprehensive technical analysis
func CalculateTechnicalIndicators(prices []float64, volumes []float64) TechnicalIndicators {
	if len(prices) < 200 {
		return TechnicalIndicators{}
	}
	
	indicators := TechnicalIndicators{}
	
	// Calculate Moving Averages
	indicators.SMA20 = calculateSMA(prices, 20)
	indicators.SMA50 = calculateSMA(prices, 50)
	indicators.SMA200 = calculateSMA(prices, 200)
	indicators.EMA9 = calculateEMA(prices, 9)
	indicators.EMA21 = calculateEMA(prices, 21)
	
	// Calculate RSI
	indicators.RSI = calculateRSI(prices, 14)
	indicators.RSIDivergence = detectRSIDivergence(prices, indicators.RSI)
	
	// Calculate MACD
	indicators.MACD = calculateMACD(prices)
	
	// Calculate Stochastic
	indicators.Stochastic = calculateStochastic(prices, 14, 3, 3)
	
	// Calculate Bollinger Bands
	indicators.BollingerBands = calculateBollingerBands(prices, 20, 2)
	
	// Calculate ATR
	indicators.ATR = calculateATR(prices, 14)
	
	// Calculate VWAP
	indicators.VWAP = calculateVWAP(prices, volumes)
	
	// Detect Support and Resistance
	indicators.Support, indicators.Resistance = findSupportResistance(prices)
	
	// Determine Trend
	indicators.Trend, indicators.TrendStrength = determineTrend(prices, indicators)
	
	// Detect Patterns
	indicators.CandlePattern = detectCandlePattern(prices)
	indicators.ChartPattern = detectChartPattern(prices)
	
	// Calculate Volume Profile
	indicators.VolumeProfile = calculateVolumeProfile(prices, volumes)
	
	// Calculate Overall Scores
	indicators.BullishScore = calculateBullishScore(indicators)
	indicators.BearishScore = calculateBearishScore(indicators)
	
	return indicators
}

// Helper functions for technical calculations

func calculateSMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}
	sum := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		sum += prices[i]
	}
	return sum / float64(period)
}

func calculateEMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}
	
	multiplier := 2.0 / float64(period+1)
	ema := calculateSMA(prices[:period], period)
	
	for i := period; i < len(prices); i++ {
		ema = (prices[i]-ema)*multiplier + ema
	}
	
	return ema
}

func calculateRSI(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 50
	}
	
	gains := 0.0
	losses := 0.0
	
	// Calculate initial average gain/loss
	for i := len(prices) - period; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains += change
		} else {
			losses += math.Abs(change)
		}
	}
	
	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)
	
	if avgLoss == 0 {
		return 100
	}
	
	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))
	
	return rsi
}

func detectRSIDivergence(prices []float64, rsi float64) bool {
	if len(prices) < 20 {
		return false
	}
	
	// Check for bullish divergence (price making lower lows, RSI making higher lows)
	recentPrices := prices[len(prices)-20:]
	priceMin := recentPrices[0]
	priceMinIdx := 0
	
	for i, p := range recentPrices {
		if p < priceMin {
			priceMin = p
			priceMinIdx = i
		}
	}
	
	// Simplified divergence detection
	if priceMinIdx > 10 && rsi < 40 {
		return true // Potential bullish divergence
	}
	
	return false
}

func calculateMACD(prices []float64) MACDValues {
	if len(prices) < 26 {
		return MACDValues{}
	}
	
	ema12 := calculateEMA(prices, 12)
	ema26 := calculateEMA(prices, 26)
	
	macd := ema12 - ema26
	signal := calculateEMA([]float64{macd}, 9) // Simplified
	histogram := macd - signal
	
	crossover := "none"
	if histogram > 0 && histogram > signal*0.01 {
		crossover = "bullish"
	} else if histogram < 0 && histogram < signal*-0.01 {
		crossover = "bearish"
	}
	
	return MACDValues{
		MACD:      macd,
		Signal:    signal,
		Histogram: histogram,
		Crossover: crossover,
	}
}

func calculateStochastic(prices []float64, period, kSmooth, dSmooth int) StochasticValues {
	if len(prices) < period {
		return StochasticValues{}
	}
	
	recentPrices := prices[len(prices)-period:]
	lowest := recentPrices[0]
	highest := recentPrices[0]
	
	for _, p := range recentPrices {
		if p < lowest {
			lowest = p
		}
		if p > highest {
			highest = p
		}
	}
	
	current := prices[len(prices)-1]
	k := 100 * ((current - lowest) / (highest - lowest))
	d := k // Simplified - should be SMA of K values
	
	return StochasticValues{
		K:          k,
		D:          d,
		Oversold:   k < 20,
		Overbought: k > 80,
	}
}

func calculateBollingerBands(prices []float64, period int, stdDev float64) BollingerValues {
	if len(prices) < period {
		return BollingerValues{}
	}
	
	sma := calculateSMA(prices, period)
	
	// Calculate standard deviation
	variance := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		diff := prices[i] - sma
		variance += diff * diff
	}
	std := math.Sqrt(variance / float64(period))
	
	return BollingerValues{
		Upper:  sma + (std * stdDev),
		Middle: sma,
		Lower:  sma - (std * stdDev),
		Width:  (std * stdDev * 2) / sma * 100, // Width as percentage
	}
}

func calculateATR(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 0
	}
	
	trValues := make([]float64, 0)
	for i := len(prices) - period; i < len(prices); i++ {
		if i == 0 {
			continue
		}
		
		high := prices[i]
		low := prices[i] * 0.98 // Simulated low
		prevClose := prices[i-1]
		
		tr := math.Max(high-low, math.Max(math.Abs(high-prevClose), math.Abs(low-prevClose)))
		trValues = append(trValues, tr)
	}
	
	// Calculate average
	sum := 0.0
	for _, tr := range trValues {
		sum += tr
	}
	
	return sum / float64(len(trValues))
}

func calculateVWAP(prices []float64, volumes []float64) float64 {
	if len(prices) == 0 || len(volumes) == 0 || len(prices) != len(volumes) {
		return 0
	}
	
	totalPV := 0.0
	totalVolume := 0.0
	
	for i := 0; i < len(prices); i++ {
		totalPV += prices[i] * volumes[i]
		totalVolume += volumes[i]
	}
	
	if totalVolume == 0 {
		return 0
	}
	
	return totalPV / totalVolume
}

func findSupportResistance(prices []float64) ([]float64, []float64) {
	if len(prices) < 20 {
		return []float64{}, []float64{}
	}
	
	// Find local minima and maxima
	support := make([]float64, 0)
	resistance := make([]float64, 0)
	
	for i := 10; i < len(prices)-10; i++ {
		isSupport := true
		isResistance := true
		
		// Check if local minimum (support)
		for j := i - 5; j <= i+5; j++ {
			if j != i && prices[j] < prices[i] {
				isSupport = false
			}
			if j != i && prices[j] > prices[i] {
				isResistance = false
			}
		}
		
		if isSupport && len(support) < 3 {
			support = append(support, prices[i])
		}
		if isResistance && len(resistance) < 3 {
			resistance = append(resistance, prices[i])
		}
	}
	
	sort.Float64s(support)
	sort.Float64s(resistance)
	
	return support, resistance
}

func determineTrend(prices []float64, indicators TechnicalIndicators) (string, float64) {
	if len(prices) < 50 {
		return "neutral", 0
	}
	
	current := prices[len(prices)-1]
	
	// Multiple trend confirmations
	trendPoints := 0.0
	
	// Moving average alignment
	if current > indicators.SMA20 && indicators.SMA20 > indicators.SMA50 && indicators.SMA50 > indicators.SMA200 {
		trendPoints += 3
	} else if current < indicators.SMA20 && indicators.SMA20 < indicators.SMA50 && indicators.SMA50 < indicators.SMA200 {
		trendPoints -= 3
	}
	
	// Price action
	recent20 := calculateSMA(prices[len(prices)-20:], 20)
	older20 := calculateSMA(prices[len(prices)-40:len(prices)-20], 20)
	if recent20 > older20 {
		trendPoints += 2
	} else {
		trendPoints -= 2
	}
	
	// MACD confirmation
	if indicators.MACD.Histogram > 0 {
		trendPoints += 1
	} else {
		trendPoints -= 1
	}
	
	// Determine trend and strength
	trend := "neutral"
	strength := math.Abs(trendPoints) / 6.0 * 100
	
	if trendPoints > 2 {
		trend = "bullish"
	} else if trendPoints < -2 {
		trend = "bearish"
	}
	
	return trend, strength
}

func detectCandlePattern(prices []float64) string {
	if len(prices) < 5 {
		return "none"
	}
	
	// Simplified candlestick pattern detection
	recent := prices[len(prices)-3:]
	
	// Bullish patterns
	if recent[0] < recent[1] && recent[1] < recent[2] && recent[2] > recent[1]*1.01 {
		return "bullish_engulfing"
	}
	
	// Bearish patterns
	if recent[0] > recent[1] && recent[1] > recent[2] && recent[2] < recent[1]*0.99 {
		return "bearish_engulfing"
	}
	
	// Doji
	if math.Abs(recent[2]-recent[1]) < recent[1]*0.001 {
		return "doji"
	}
	
	return "none"
}

func detectChartPattern(prices []float64) string {
	if len(prices) < 50 {
		return "none"
	}
	
	// Simplified chart pattern detection
	recent50 := prices[len(prices)-50:]
	
	// Check for triangle pattern
	highs := make([]float64, 0)
	lows := make([]float64, 0)
	
	for i := 0; i < len(recent50); i += 5 {
		segment := recent50[i:min(i+5, len(recent50))]
		high := segment[0]
		low := segment[0]
		
		for _, p := range segment {
			if p > high {
				high = p
			}
			if p < low {
				low = p
			}
		}
		
		highs = append(highs, high)
		lows = append(lows, low)
	}
	
	// Check if highs are descending and lows are ascending (triangle)
	if len(highs) >= 3 && len(lows) >= 3 {
		highsDescending := highs[0] > highs[len(highs)-1]
		lowsAscending := lows[0] < lows[len(lows)-1]
		
		if highsDescending && lowsAscending {
			return "triangle"
		}
		
		if !highsDescending && !lowsAscending {
			return "channel"
		}
	}
	
	return "none"
}

func calculateVolumeProfile(prices []float64, volumes []float64) VolumeProfileData {
	if len(prices) == 0 || len(volumes) == 0 {
		return VolumeProfileData{}
	}
	
	// Calculate volume-weighted price levels
	priceVolume := make(map[float64]float64)
	
	for i := 0; i < len(prices) && i < len(volumes); i++ {
		// Round price to nearest 0.5 for grouping
		roundedPrice := math.Round(prices[i]*2) / 2
		priceVolume[roundedPrice] += volumes[i]
	}
	
	// Find Point of Control (price with highest volume)
	poc := 0.0
	maxVolume := 0.0
	
	for price, vol := range priceVolume {
		if vol > maxVolume {
			maxVolume = vol
			poc = price
		}
	}
	
	// Calculate value area (70% of volume)
	totalVolume := 0.0
	for _, vol := range volumes {
		totalVolume += vol
	}
	
	// Simplified VAH and VAL
	vah := poc * 1.01
	val := poc * 0.99
	
	// Check for volume increase
	recentVolume := calculateSMA(volumes[max(0, len(volumes)-10):], min(10, len(volumes)))
	olderVolume := calculateSMA(volumes[max(0, len(volumes)-20):max(0, len(volumes)-10)], min(10, len(volumes)))
	
	volumeIncrease := recentVolume > olderVolume*1.2
	
	// Calculate Accumulation/Distribution
	ad := 0.0
	if len(prices) > 1 && len(volumes) > 0 {
		moneyFlow := ((prices[len(prices)-1] - prices[len(prices)-2]) / prices[len(prices)-2]) * volumes[len(volumes)-1]
		ad = moneyFlow / totalVolume * 100
	}
	
	return VolumeProfileData{
		POC:              poc,
		VAH:              vah,
		VAL:              val,
		VolumeIncrease:   volumeIncrease,
		AccumulationDist: ad,
	}
}

func calculateBullishScore(indicators TechnicalIndicators) float64 {
	score := 0.0
	factors := 0.0
	
	// Trend (weight: 25%)
	if indicators.Trend == "bullish" {
		score += 25 * (indicators.TrendStrength / 100)
		factors += 25
	} else {
		factors += 25
	}
	
	// RSI (weight: 15%)
	if indicators.RSI > 30 && indicators.RSI < 70 {
		score += 15 * ((indicators.RSI - 30) / 40)
		factors += 15
	} else if indicators.RSI <= 30 {
		score += 15 // Oversold = bullish
		factors += 15
	} else {
		factors += 15
	}
	
	// MACD (weight: 20%)
	if indicators.MACD.Crossover == "bullish" {
		score += 20
		factors += 20
	} else if indicators.MACD.Histogram > 0 {
		score += 10
		factors += 20
	} else {
		factors += 20
	}
	
	// Stochastic (weight: 10%)
	if indicators.Stochastic.Oversold {
		score += 10
		factors += 10
	} else if indicators.Stochastic.K > 20 && indicators.Stochastic.K < 80 {
		score += 5
		factors += 10
	} else {
		factors += 10
	}
	
	// Volume (weight: 15%)
	if indicators.VolumeProfile.VolumeIncrease && indicators.VolumeProfile.AccumulationDist > 0 {
		score += 15
		factors += 15
	} else if indicators.VolumeProfile.AccumulationDist > 0 {
		score += 7.5
		factors += 15
	} else {
		factors += 15
	}
	
	// Patterns (weight: 15%)
	if indicators.CandlePattern == "bullish_engulfing" {
		score += 15
		factors += 15
	} else if indicators.ChartPattern == "triangle" || indicators.ChartPattern == "channel" {
		score += 7.5
		factors += 15
	} else {
		factors += 15
	}
	
	if factors == 0 {
		return 50
	}
	
	return (score / factors) * 100
}

func calculateBearishScore(indicators TechnicalIndicators) float64 {
	score := 0.0
	factors := 0.0
	
	// Trend (weight: 25%)
	if indicators.Trend == "bearish" {
		score += 25 * (indicators.TrendStrength / 100)
		factors += 25
	} else {
		factors += 25
	}
	
	// RSI (weight: 15%)
	if indicators.RSI > 70 {
		score += 15 // Overbought = bearish
		factors += 15
	} else if indicators.RSI > 50 && indicators.RSI <= 70 {
		score += 15 * ((70 - indicators.RSI) / 20)
		factors += 15
	} else {
		factors += 15
	}
	
	// MACD (weight: 20%)
	if indicators.MACD.Crossover == "bearish" {
		score += 20
		factors += 20
	} else if indicators.MACD.Histogram < 0 {
		score += 10
		factors += 20
	} else {
		factors += 20
	}
	
	// Stochastic (weight: 10%)
	if indicators.Stochastic.Overbought {
		score += 10
		factors += 10
	} else if indicators.Stochastic.K > 50 {
		score += 5
		factors += 10
	} else {
		factors += 10
	}
	
	// Volume (weight: 15%)
	if indicators.VolumeProfile.VolumeIncrease && indicators.VolumeProfile.AccumulationDist < 0 {
		score += 15
		factors += 15
	} else if indicators.VolumeProfile.AccumulationDist < 0 {
		score += 7.5
		factors += 15
	} else {
		factors += 15
	}
	
	// Patterns (weight: 15%)
	if indicators.CandlePattern == "bearish_engulfing" {
		score += 15
		factors += 15
	} else if indicators.CandlePattern == "doji" && indicators.Trend == "bearish" {
		score += 7.5
		factors += 15
	} else {
		factors += 15
	}
	
	if factors == 0 {
		return 50
	}
	
	return (score / factors) * 100
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GenerateTradeSignal creates actionable trade signals based on analysis
func GenerateTradeSignal(analysis MarketAnalysis, riskTolerance string) TradeSignal {
	signal := TradeSignal{
		Action:   "HOLD",
		Strength: "weak",
		Reasons:  make([]string, 0),
		Warnings: make([]string, 0),
	}
	
	// Determine action based on scores
	bullishScore := analysis.Technical.BullishScore
	bearishScore := analysis.Technical.BearishScore
	confidence := analysis.Confidence
	
	// Strong buy signal
	if bullishScore > 70 && confidence > 75 {
		signal.Action = "BUY"
		signal.Strength = "strong"
		signal.Timeframe = determineTimeframe(analysis)
		signal.Reasons = append(signal.Reasons, fmt.Sprintf("Strong bullish score: %.1f%%", bullishScore))
		
		if analysis.Technical.RSI < 40 {
			signal.Reasons = append(signal.Reasons, "RSI oversold - good entry point")
		}
		if analysis.Technical.MACD.Crossover == "bullish" {
			signal.Reasons = append(signal.Reasons, "MACD bullish crossover")
		}
		if analysis.Technical.Trend == "bullish" {
			signal.Reasons = append(signal.Reasons, fmt.Sprintf("Bullish trend with %.1f%% strength", analysis.Technical.TrendStrength))
		}
		
	} else if bullishScore > 60 && confidence > 65 {
		signal.Action = "BUY"
		signal.Strength = "moderate"
		signal.Timeframe = determineTimeframe(analysis)
		signal.Reasons = append(signal.Reasons, fmt.Sprintf("Moderate bullish score: %.1f%%", bullishScore))
		
	} else if bearishScore > 70 && confidence > 75 {
		signal.Action = "SELL"
		signal.Strength = "strong"
		signal.Reasons = append(signal.Reasons, fmt.Sprintf("Strong bearish score: %.1f%%", bearishScore))
		
		if analysis.Technical.RSI > 70 {
			signal.Warnings = append(signal.Warnings, "RSI overbought - potential reversal")
		}
	}
	
	// Add warnings
	if analysis.Technical.BollingerBands.Width > 5 {
		signal.Warnings = append(signal.Warnings, "High volatility detected")
	}
	
	if analysis.RiskReward.RiskRewardRatio < 2 {
		signal.Warnings = append(signal.Warnings, "Risk-reward ratio below optimal (< 1:2)")
	}
	
	// Set strategy based on signals
	signal.Strategy = determineStrategy(analysis, signal)
	
	// Calculate expected return
	if signal.Action == "BUY" {
		signal.ExpectedReturn = ((analysis.RiskReward.Target1 - analysis.RiskReward.EntryPrice) / analysis.RiskReward.EntryPrice) * 100
		signal.HoldingPeriod = estimateHoldingPeriod(signal.Timeframe)
	}
	
	// Set priority
	signal.Priority = calculatePriority(signal, analysis)
	
	return signal
}

func determineTimeframe(analysis MarketAnalysis) string {
	// Based on ATR and trend strength
	if analysis.Technical.ATR < analysis.RiskReward.EntryPrice*0.01 {
		return "intraday"
	} else if analysis.Technical.TrendStrength > 70 {
		return "positional"
	}
	return "swing"
}

func determineStrategy(analysis MarketAnalysis, signal TradeSignal) string {
	if signal.Action != "BUY" {
		return "Wait for better entry"
	}
	
	strategies := []string{}
	
	if analysis.Technical.RSI < 30 {
		strategies = append(strategies, "Oversold bounce play")
	}
	
	if analysis.Technical.MACD.Crossover == "bullish" {
		strategies = append(strategies, "MACD momentum trade")
	}
	
	if analysis.Technical.Trend == "bullish" && analysis.Technical.TrendStrength > 60 {
		strategies = append(strategies, "Trend following")
	}
	
	if len(analysis.Technical.Support) > 0 {
		currentPrice := analysis.RiskReward.EntryPrice
		for _, support := range analysis.Technical.Support {
			if math.Abs(currentPrice-support)/support < 0.02 {
				strategies = append(strategies, "Support level bounce")
				break
			}
		}
	}
	
	if analysis.Technical.ChartPattern == "triangle" {
		strategies = append(strategies, "Triangle breakout")
	}
	
	if len(strategies) == 0 {
		return "General momentum trade"
	}
	
	return strategies[0] // Return primary strategy
}

func estimateHoldingPeriod(timeframe string) string {
	switch timeframe {
	case "intraday":
		return "1-2 days"
	case "swing":
		return "3-10 days"
	case "positional":
		return "2-4 weeks"
	default:
		return "5-7 days"
	}
}

func calculatePriority(signal TradeSignal, analysis MarketAnalysis) int {
	priority := 5 // Base priority
	
	if signal.Strength == "strong" {
		priority += 2
	} else if signal.Strength == "moderate" {
		priority += 1
	}
	
	if analysis.Confidence > 80 {
		priority += 2
	} else if analysis.Confidence > 70 {
		priority += 1
	}
	
	if analysis.RiskReward.RiskRewardRatio > 3 {
		priority += 1
	}
	
	if len(signal.Warnings) > 2 {
		priority -= 2
	} else if len(signal.Warnings) > 0 {
		priority -= 1
	}
	
	// Cap priority between 1 and 10
	if priority > 10 {
		priority = 10
	} else if priority < 1 {
		priority = 1
	}
	
	return priority
}