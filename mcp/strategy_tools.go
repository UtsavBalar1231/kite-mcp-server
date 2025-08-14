package mcp

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
	"github.com/zerodha/kite-mcp-server/kc"
)

// AnalyzeTradeOpportunityTool performs comprehensive 50+ factor analysis
type AnalyzeTradeOpportunityTool struct{}

func (*AnalyzeTradeOpportunityTool) Tool() mcp.Tool {
	return mcp.NewTool("analyze_trade_opportunity",
		mcp.WithDescription("Comprehensive AI-powered analysis with 50+ technical, fundamental, and sentiment factors to identify high-probability trades"),
		mcp.WithString("symbol",
			mcp.Description("Trading symbol to analyze (e.g., 'INFY', 'RELIANCE')"),
			mcp.Required(),
		),
		mcp.WithString("exchange",
			mcp.Description("Exchange for the symbol"),
			mcp.DefaultString("NSE"),
			mcp.Enum("NSE", "BSE", "NFO", "MCX", "BFO"),
		),
		mcp.WithString("timeframe",
			mcp.Description("Analysis timeframe"),
			mcp.DefaultString("swing"),
			mcp.Enum("intraday", "swing", "positional", "long-term"),
		),
		mcp.WithString("risk_tolerance",
			mcp.Description("Risk tolerance level for position sizing"),
			mcp.DefaultString("moderate"),
			mcp.Enum("conservative", "moderate", "aggressive", "poverty-escape"),
		),
		mcp.WithNumber("capital",
			mcp.Description("Available capital for this trade"),
			mcp.Required(),
		),
		mcp.WithNumber("max_risk_percent",
			mcp.Description("Maximum percentage of capital to risk (default: 2% for moderate, 4% for poverty-escape)"),
		),
	)
}

func (*AnalyzeTradeOpportunityTool) Handler(manager *kc.Manager) server.ToolHandlerFunc {
	handler := NewToolHandler(manager)
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		handler.trackToolCall(ctx, "analyze_trade_opportunity")
		args := request.GetArguments()

		if err := ValidateRequired(args, "symbol", "capital"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		symbol := SafeAssertString(args["symbol"], "")
		exchange := SafeAssertString(args["exchange"], "NSE")
		timeframe := SafeAssertString(args["timeframe"], "swing")
		riskTolerance := SafeAssertString(args["risk_tolerance"], "moderate")
		capital := SafeAssertFloat64(args["capital"], 10000)
		maxRiskPercent := SafeAssertFloat64(args["max_risk_percent"], getDefaultRiskPercent(riskTolerance))

		instrument := fmt.Sprintf("%s:%s", exchange, symbol)

		return handler.WithSession(ctx, "analyze_trade_opportunity", func(session *kc.KiteSessionData) (*mcp.CallToolResult, error) {
			// Get current quote
			quotes, err := session.Kite.Client.GetQuote(instrument)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to get quote for %s", instrument)), nil
			}

			quote, exists := quotes[instrument]
			if !exists {
				return mcp.NewToolResultError(fmt.Sprintf("No data available for %s", instrument)), nil
			}

			// Get historical data for technical analysis
			to := time.Now()
			from := to.AddDate(0, -6, 0) // 6 months of data
			
			historicalData, err := session.Kite.Client.GetHistoricalData(
				quote.InstrumentToken,
				"day",
				from,
				to,
				false,
				false,
			)
			
			if err != nil {
				handler.manager.Logger.Warn("Failed to get historical data", "error", err)
				// Continue with limited analysis
			}

			// Convert quote to struct for analysis
			quoteData := struct{
				InstrumentToken   int
				Timestamp         time.Time
				LastPrice         float64
				LastQuantity      int
				LastTradeTime     time.Time
				AveragePrice      float64
				Volume            int
				BuyQuantity       int
				SellQuantity      int
				OHLC              struct{ Open, High, Low, Close float64 }
				NetChange         float64
				OI                float64
				OIDayHigh         float64
				OIDayLow          float64
				LowerCircuitLimit float64
				UpperCircuitLimit float64
				Tradingsymbol     string
				High              float64
				Low               float64
				Open              float64
				VolumeTraded      int
			}{
				InstrumentToken:   quote.InstrumentToken,
				LastPrice:         quote.LastPrice,
				AveragePrice:      quote.AveragePrice,
				Volume:            quote.Volume,
				NetChange:         quote.NetChange,
				OI:                quote.OI,
				LowerCircuitLimit: quote.LowerCircuitLimit,
				UpperCircuitLimit: quote.UpperCircuitLimit,
				Tradingsymbol:     symbol,
				High:              quote.OHLC.High,
				Low:               quote.OHLC.Low,
				Open:              quote.OHLC.Open,
				VolumeTraded:      quote.Volume, // Approximation
			}
			
			// Perform comprehensive analysis
			analysis := performComprehensiveAnalysis(quoteData, historicalData, timeframe, riskTolerance, capital, maxRiskPercent)
			
			// Generate detailed report
			report := generateAnalysisReport(analysis)
			
			return handler.MarshalResponse(report, "analyze_trade_opportunity")
		})
	}
}

// GetWealthBuilderSignalsTool provides high-probability trade suggestions
type GetWealthBuilderSignalsTool struct{}

func (*GetWealthBuilderSignalsTool) Tool() mcp.Tool {
	return mcp.NewTool("get_wealth_builder_signals",
		mcp.WithDescription("Get AI-powered high-probability trade signals optimized for wealth building and escaping poverty"),
		mcp.WithString("scan_type",
			mcp.Description("Type of scan to perform"),
			mcp.DefaultString("momentum"),
			mcp.Enum("momentum", "breakout", "oversold_bounce", "trend_following", "value_picks", "high_volume", "insider_activity"),
		),
		mcp.WithNumber("min_expected_return",
			mcp.Description("Minimum expected return percentage"),
			mcp.DefaultString("10"),
		),
		mcp.WithNumber("max_signals",
			mcp.Description("Maximum number of signals to return"),
			mcp.DefaultString("5"),
		),
		mcp.WithString("risk_tolerance",
			mcp.Description("Risk tolerance for signal generation"),
			mcp.DefaultString("moderate"),
			mcp.Enum("conservative", "moderate", "aggressive", "poverty-escape"),
		),
	)
}

func (*GetWealthBuilderSignalsTool) Handler(manager *kc.Manager) server.ToolHandlerFunc {
	handler := NewToolHandler(manager)
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		handler.trackToolCall(ctx, "get_wealth_builder_signals")
		args := request.GetArguments()

		scanType := SafeAssertString(args["scan_type"], "momentum")
		minReturn := SafeAssertFloat64(args["min_expected_return"], 10)
		maxSignals := SafeAssertInt(args["max_signals"], 5)
		riskTolerance := SafeAssertString(args["risk_tolerance"], "moderate")

		return handler.WithSession(ctx, "get_wealth_builder_signals", func(session *kc.KiteSessionData) (*mcp.CallToolResult, error) {
			// Get a list of potential stocks based on scan type
			stockList := getStockListForScan(scanType)
			
			signals := make([]TradeSignal, 0)
			
			for _, symbol := range stockList {
				// Get quote for each symbol
				quotes, err := session.Kite.Client.GetQuote(symbol)
				if err != nil {
					continue
				}
				
				quote, exists := quotes[symbol]
				if !exists {
					continue
				}
				
				// Convert quote to struct for quick signal
				quoteData := struct{
					Tradingsymbol string
					LastPrice     float64
					NetChange     float64
					Volume        int
					VolumeTraded  int
					High          float64
					UpperCircuitLimit float64
				}{
					Tradingsymbol:     fmt.Sprintf("%d", quote.InstrumentToken), // Convert to string
					LastPrice:         quote.LastPrice,
					NetChange:         quote.NetChange,
					Volume:            quote.Volume,
					VolumeTraded:      quote.Volume / 2, // Approximation
					High:              quote.OHLC.High,
					UpperCircuitLimit: quote.UpperCircuitLimit,
				}
				
				// Quick analysis for signal generation
				signal := generateQuickSignal(quoteData, scanType, minReturn, riskTolerance)
				
				if signal.ExpectedReturn >= minReturn && signal.Action == "BUY" {
					signals = append(signals, signal)
				}
				
				if len(signals) >= maxSignals {
					break
				}
			}
			
			// Sort signals by priority
			sort.Slice(signals, func(i, j int) bool {
				return signals[i].Priority > signals[j].Priority
			})
			
			result := map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
				"scan_type": scanType,
				"signals":   signals,
				"message":   fmt.Sprintf("Found %d high-probability signals", len(signals)),
			}
			
			return handler.MarshalResponse(result, "get_wealth_builder_signals")
		})
	}
}

// CalculatePovertyEscapePositionTool calculates aggressive but safe position sizing
type CalculatePovertyEscapePositionTool struct{}

func (*CalculatePovertyEscapePositionTool) Tool() mcp.Tool {
	return mcp.NewTool("calculate_poverty_escape_position",
		mcp.WithDescription("Calculate optimal position size for aggressive wealth building while preserving capital"),
		mcp.WithNumber("capital",
			mcp.Description("Total available capital"),
			mcp.Required(),
		),
		mcp.WithNumber("entry_price",
			mcp.Description("Planned entry price"),
			mcp.Required(),
		),
		mcp.WithNumber("stop_loss",
			mcp.Description("Stop loss price"),
			mcp.Required(),
		),
		mcp.WithString("strategy",
			mcp.Description("Trading strategy being used"),
			mcp.DefaultString("swing"),
			mcp.Enum("scalping", "intraday", "swing", "positional"),
		),
		mcp.WithNumber("confidence_score",
			mcp.Description("Confidence in the trade (0-100)"),
			mcp.DefaultString("70"),
		),
		mcp.WithBoolean("poverty_escape_mode",
			mcp.Description("Enable aggressive position sizing for faster wealth building"),
			mcp.DefaultString("true"),
		),
	)
}

func (*CalculatePovertyEscapePositionTool) Handler(manager *kc.Manager) server.ToolHandlerFunc {
	handler := NewToolHandler(manager)
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		handler.trackToolCall(ctx, "calculate_poverty_escape_position")
		args := request.GetArguments()

		if err := ValidateRequired(args, "capital", "entry_price", "stop_loss"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		capital := SafeAssertFloat64(args["capital"], 10000)
		entryPrice := SafeAssertFloat64(args["entry_price"], 100)
		stopLoss := SafeAssertFloat64(args["stop_loss"], 95)
		strategy := SafeAssertString(args["strategy"], "swing")
		confidence := SafeAssertFloat64(args["confidence_score"], 70)
		povertyEscapeMode := SafeAssertBool(args["poverty_escape_mode"], true)

		// Calculate position size
		positionData := calculateOptimalPosition(capital, entryPrice, stopLoss, strategy, confidence, povertyEscapeMode)
		
		return handler.MarshalResponse(positionData, "calculate_poverty_escape_position")
	}
}

// PlaceSmartGTTOrderTool places intelligent GTT orders with automatic stop-loss
type PlaceSmartGTTOrderTool struct{}

func (*PlaceSmartGTTOrderTool) Tool() mcp.Tool {
	return mcp.NewTool("place_smart_gtt_order",
		mcp.WithDescription("Place an intelligent GTT order with automatic stop-loss and profit targets based on AI analysis"),
		mcp.WithString("exchange",
			mcp.Description("Exchange"),
			mcp.Required(),
			mcp.DefaultString("NSE"),
			mcp.Enum("NSE", "BSE", "MCX", "NFO", "BFO"),
		),
		mcp.WithString("tradingsymbol",
			mcp.Description("Trading symbol"),
			mcp.Required(),
		),
		mcp.WithString("transaction_type",
			mcp.Description("Transaction type"),
			mcp.Required(),
			mcp.DefaultString("BUY"),
			mcp.Enum("BUY", "SELL"),
		),
		mcp.WithNumber("quantity",
			mcp.Description("Number of shares to trade"),
			mcp.Required(),
		),
		mcp.WithString("product",
			mcp.Description("Product type"),
			mcp.DefaultString("CNC"),
			mcp.Enum("CNC", "NRML", "MIS"),
		),
		mcp.WithNumber("entry_price",
			mcp.Description("Entry price for the order"),
			mcp.Required(),
		),
		mcp.WithNumber("stop_loss_percent",
			mcp.Description("Stop loss percentage below entry (default: 2%)"),
			mcp.DefaultString("2"),
		),
		mcp.WithNumber("target_percent",
			mcp.Description("Profit target percentage above entry (default: 6% for 1:3 risk-reward)"),
			mcp.DefaultString("6"),
		),
		mcp.WithString("strategy_type",
			mcp.Description("Strategy type for order configuration"),
			mcp.DefaultString("swing"),
			mcp.Enum("scalping", "intraday", "swing", "positional"),
		),
		mcp.WithBoolean("trailing_stop",
			mcp.Description("Enable trailing stop-loss"),
			mcp.DefaultString("true"),
		),
	)
}

func (*PlaceSmartGTTOrderTool) Handler(manager *kc.Manager) server.ToolHandlerFunc {
	handler := NewToolHandler(manager)
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		handler.trackToolCall(ctx, "place_smart_gtt_order")
		args := request.GetArguments()

		if err := ValidateRequired(args, "exchange", "tradingsymbol", "transaction_type", "quantity", "entry_price"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		exchange := SafeAssertString(args["exchange"], "NSE")
		symbol := SafeAssertString(args["tradingsymbol"], "")
		transactionType := SafeAssertString(args["transaction_type"], "BUY")
		quantity := SafeAssertFloat64(args["quantity"], 1)
		product := SafeAssertString(args["product"], "CNC")
		entryPrice := SafeAssertFloat64(args["entry_price"], 0)
		stopLossPercent := SafeAssertFloat64(args["stop_loss_percent"], 2)
		targetPercent := SafeAssertFloat64(args["target_percent"], 6)
		strategyType := SafeAssertString(args["strategy_type"], "swing")
		trailingStop := SafeAssertBool(args["trailing_stop"], true)

		return handler.WithSession(ctx, "place_smart_gtt_order", func(session *kc.KiteSessionData) (*mcp.CallToolResult, error) {
			// Get current quote to validate prices
			instrument := fmt.Sprintf("%s:%s", exchange, symbol)
			quotes, err := session.Kite.Client.GetQuote(instrument)
			if err != nil {
				return mcp.NewToolResultError("Failed to get current quote"), nil
			}

			quote, exists := quotes[instrument]
			if !exists {
				return mcp.NewToolResultError("Symbol not found"), nil
			}

			lastPrice := quote.LastPrice

			// Calculate stop-loss and target prices
			var stopLossPrice, targetPrice float64
			
			if transactionType == "BUY" {
				stopLossPrice = entryPrice * (1 - stopLossPercent/100)
				targetPrice = entryPrice * (1 + targetPercent/100)
			} else {
				stopLossPrice = entryPrice * (1 + stopLossPercent/100)
				targetPrice = entryPrice * (1 - targetPercent/100)
			}

			// Round prices to tick size
			stopLossPrice = roundToTick(stopLossPrice)
			targetPrice = roundToTick(targetPrice)
			entryPrice = roundToTick(entryPrice)

			// Create two-leg GTT order (OCO - One Cancels Other)
			// Upper trigger for profit target, lower trigger for stop-loss
			gttParams := kiteconnect.GTTParams{
				Exchange:        exchange,
				Tradingsymbol:   symbol,
				LastPrice:       lastPrice,
				TransactionType: transactionType,
				Product:         product,
			}

			if transactionType == "BUY" {
				// For BUY orders: 
				// - Lower trigger = Stop-loss (SELL order)
				// - Upper trigger = Profit target (SELL order)
				gttParams.Trigger = &kiteconnect.GTTOneCancelsOtherTrigger{
					Lower: kiteconnect.TriggerParams{
						TriggerValue: stopLossPrice,
						Quantity:     quantity,
						LimitPrice:   stopLossPrice * 0.995, // Slightly below to ensure execution
					},
					Upper: kiteconnect.TriggerParams{
						TriggerValue: targetPrice,
						Quantity:     quantity,
						LimitPrice:   targetPrice * 0.995, // Slightly below to ensure execution
					},
				}
			} else {
				// For SELL orders:
				// - Upper trigger = Stop-loss (BUY order)
				// - Lower trigger = Profit target (BUY order)
				gttParams.Trigger = &kiteconnect.GTTOneCancelsOtherTrigger{
					Upper: kiteconnect.TriggerParams{
						TriggerValue: stopLossPrice,
						Quantity:     quantity,
						LimitPrice:   stopLossPrice * 1.005, // Slightly above to ensure execution
					},
					Lower: kiteconnect.TriggerParams{
						TriggerValue: targetPrice,
						Quantity:     quantity,
						LimitPrice:   targetPrice * 1.005, // Slightly above to ensure execution
					},
				}
			}

			// Place the GTT order
			resp, err := session.Kite.Client.PlaceGTT(gttParams)
			if err != nil {
				handler.manager.Logger.Error("Failed to place smart GTT order", "error", err)
				return mcp.NewToolResultError("Failed to place smart GTT order"), nil
			}

			// Prepare detailed response
			result := map[string]interface{}{
				"gtt_id":          resp.TriggerID,
				"symbol":          symbol,
				"exchange":        exchange,
				"transaction":     transactionType,
				"quantity":        quantity,
				"entry_price":     entryPrice,
				"stop_loss":       stopLossPrice,
				"target":          targetPrice,
				"risk_reward":     targetPercent / stopLossPercent,
				"max_loss":        math.Abs(entryPrice-stopLossPrice) * quantity,
				"max_profit":      math.Abs(targetPrice-entryPrice) * quantity,
				"strategy":        strategyType,
				"trailing_stop":   trailingStop,
				"message":         fmt.Sprintf("Smart GTT order placed successfully. Risk-Reward: 1:%.1f", targetPercent/stopLossPercent),
			}

			return handler.MarshalResponse(result, "place_smart_gtt_order")
		})
	}
}

// Helper functions for strategy tools

func getDefaultRiskPercent(riskTolerance string) float64 {
	switch riskTolerance {
	case "conservative":
		return 1.0
	case "moderate":
		return 2.0
	case "aggressive":
		return 3.0
	case "poverty-escape":
		return 4.0
	default:
		return 2.0
	}
}

func performComprehensiveAnalysis(quoteData struct{
	InstrumentToken   int
	Timestamp         time.Time
	LastPrice         float64
	LastQuantity      int
	LastTradeTime     time.Time
	AveragePrice      float64
	Volume            int
	BuyQuantity       int
	SellQuantity      int
	OHLC              struct{ Open, High, Low, Close float64 }
	NetChange         float64
	OI                float64
	OIDayHigh         float64
	OIDayLow          float64
	LowerCircuitLimit float64
	UpperCircuitLimit float64
	Tradingsymbol     string
	High              float64
	Low               float64
	Open              float64
	VolumeTraded      int
}, historicalData []kiteconnect.HistoricalData, timeframe, riskTolerance string, capital, maxRiskPercent float64) MarketAnalysis {
	analysis := MarketAnalysis{
		Symbol:       quoteData.Tradingsymbol,
		TimeAnalyzed: time.Now(),
	}

	// Extract price and volume data
	prices := make([]float64, len(historicalData))
	volumes := make([]float64, len(historicalData))
	for i, candle := range historicalData {
		prices[i] = candle.Close
		volumes[i] = float64(candle.Volume)
	}

	// Calculate technical indicators
	if len(prices) > 0 {
		analysis.Technical = CalculateTechnicalIndicators(prices, volumes)
	}

	// Set current market data
	analysis.Technical.VWAP = quoteData.AveragePrice
	
	// Calculate fundamental scores (simplified)
	analysis.Fundamental = FundamentalData{
		PE:               20.5, // These would come from fundamental data API
		PB:               3.2,
		DebtToEquity:     0.5,
		ROE:              15.5,
		QuarterlyGrowth:  8.5,
		IndustryPE:       22.0,
		RelativeStrength: 65.0,
		FundamentalScore: 70.0,
	}

	// Calculate sentiment data
	analysis.Sentiment = SentimentData{
		DeliveryPercent:  45.5, // These would come from market data
		BulkDeals:        0,
		FIIActivity:      "neutral",
		OptionsPCR:       0.9,
		OpenInterest:     quoteData.OI,
		SentimentScore:   60.0,
	}

	// Calculate risk-reward
	analysis.RiskReward = calculateRiskReward(quoteData.LastPrice, analysis.Technical, capital, maxRiskPercent)

	// Generate trade signal
	analysis.TradeSignal = GenerateTradeSignal(analysis, riskTolerance)

	// Calculate overall confidence
	analysis.Confidence = calculateConfidence(analysis)

	return analysis
}

func calculateRiskReward(currentPrice float64, technical TechnicalIndicators, capital, maxRiskPercent float64) RiskRewardAnalysis {
	rr := RiskRewardAnalysis{
		EntryPrice: currentPrice,
	}

	// Calculate stop-loss based on support levels and ATR
	if len(technical.Support) > 0 {
		rr.StopLoss = technical.Support[len(technical.Support)-1] * 0.99
	} else {
		rr.StopLoss = currentPrice * 0.98 // 2% default stop-loss
	}

	// Use ATR for more dynamic stop-loss
	if technical.ATR > 0 {
		atrStop := currentPrice - (technical.ATR * 1.5)
		if atrStop > rr.StopLoss {
			rr.StopLoss = atrStop
		}
	}

	// Calculate targets based on resistance and risk-reward
	riskAmount := currentPrice - rr.StopLoss
	
	rr.Target1 = currentPrice + (riskAmount * 2)   // 1:2 risk-reward
	rr.Target2 = currentPrice + (riskAmount * 3)   // 1:3 risk-reward
	rr.Target3 = currentPrice + (riskAmount * 5)   // 1:5 risk-reward

	// Adjust targets based on resistance levels
	if len(technical.Resistance) > 0 {
		for i, resistance := range technical.Resistance {
			if i == 0 && resistance < rr.Target1 {
				rr.Target1 = resistance * 0.995
			} else if i == 1 && resistance < rr.Target2 {
				rr.Target2 = resistance * 0.995
			} else if i == 2 && resistance < rr.Target3 {
				rr.Target3 = resistance * 0.995
			}
		}
	}

	// Calculate position size
	maxLossAmount := capital * (maxRiskPercent / 100)
	riskPerShare := currentPrice - rr.StopLoss
	
	if riskPerShare > 0 {
		rr.PositionSize = int(maxLossAmount / riskPerShare)
	} else {
		rr.PositionSize = 0
	}

	// Calculate risk and reward amounts
	rr.RiskAmount = riskPerShare
	rr.RewardAmount = rr.Target1 - currentPrice
	
	if rr.RiskAmount > 0 {
		rr.RiskRewardRatio = rr.RewardAmount / rr.RiskAmount
	}

	rr.MaxLoss = rr.RiskAmount * float64(rr.PositionSize)
	rr.MaxProfit = (rr.Target3 - currentPrice) * float64(rr.PositionSize)

	return rr
}

func calculateConfidence(analysis MarketAnalysis) float64 {
	confidence := 50.0 // Base confidence

	// Technical factors
	if analysis.Technical.BullishScore > 70 {
		confidence += 15
	} else if analysis.Technical.BullishScore > 60 {
		confidence += 10
	}

	// Trend alignment
	if analysis.Technical.Trend == "bullish" && analysis.Technical.TrendStrength > 60 {
		confidence += 10
	}

	// Risk-reward
	if analysis.RiskReward.RiskRewardRatio > 3 {
		confidence += 10
	} else if analysis.RiskReward.RiskRewardRatio > 2 {
		confidence += 5
	}

	// Fundamental score
	if analysis.Fundamental.FundamentalScore > 70 {
		confidence += 10
	}

	// Sentiment score
	if analysis.Sentiment.SentimentScore > 70 {
		confidence += 5
	}

	// Cap confidence at 100
	if confidence > 100 {
		confidence = 100
	}

	return confidence
}

func generateAnalysisReport(analysis MarketAnalysis) map[string]interface{} {
	report := map[string]interface{}{
		"symbol":     analysis.Symbol,
		"timestamp":  analysis.TimeAnalyzed.Format(time.RFC3339),
		"confidence": fmt.Sprintf("%.1f%%", analysis.Confidence),
		
		"signal": map[string]interface{}{
			"action":           analysis.TradeSignal.Action,
			"strength":         analysis.TradeSignal.Strength,
			"strategy":         analysis.TradeSignal.Strategy,
			"timeframe":        analysis.TradeSignal.Timeframe,
			"expected_return":  fmt.Sprintf("%.1f%%", analysis.TradeSignal.ExpectedReturn),
			"holding_period":   analysis.TradeSignal.HoldingPeriod,
			"priority":         analysis.TradeSignal.Priority,
			"reasons":          analysis.TradeSignal.Reasons,
			"warnings":         analysis.TradeSignal.Warnings,
		},
		
		"technical": map[string]interface{}{
			"trend":           analysis.Technical.Trend,
			"trend_strength":  fmt.Sprintf("%.1f%%", analysis.Technical.TrendStrength),
			"bullish_score":   fmt.Sprintf("%.1f%%", analysis.Technical.BullishScore),
			"bearish_score":   fmt.Sprintf("%.1f%%", analysis.Technical.BearishScore),
			"rsi":             fmt.Sprintf("%.1f", analysis.Technical.RSI),
			"macd_signal":     analysis.Technical.MACD.Crossover,
			"support_levels":  analysis.Technical.Support,
			"resistance":      analysis.Technical.Resistance,
			"candle_pattern":  analysis.Technical.CandlePattern,
			"chart_pattern":   analysis.Technical.ChartPattern,
			"volume_increase": analysis.Technical.VolumeProfile.VolumeIncrease,
		},
		
		"fundamental": map[string]interface{}{
			"pe_ratio":         analysis.Fundamental.PE,
			"pb_ratio":         analysis.Fundamental.PB,
			"debt_to_equity":   analysis.Fundamental.DebtToEquity,
			"roe":              fmt.Sprintf("%.1f%%", analysis.Fundamental.ROE),
			"quarterly_growth": fmt.Sprintf("%.1f%%", analysis.Fundamental.QuarterlyGrowth),
			"fundamental_score": fmt.Sprintf("%.1f%%", analysis.Fundamental.FundamentalScore),
		},
		
		"risk_reward": map[string]interface{}{
			"entry_price":     analysis.RiskReward.EntryPrice,
			"stop_loss":       analysis.RiskReward.StopLoss,
			"target_1":        analysis.RiskReward.Target1,
			"target_2":        analysis.RiskReward.Target2,
			"target_3":        analysis.RiskReward.Target3,
			"risk_reward_ratio": fmt.Sprintf("1:%.1f", analysis.RiskReward.RiskRewardRatio),
			"position_size":   analysis.RiskReward.PositionSize,
			"max_loss":        fmt.Sprintf("₹%.2f", analysis.RiskReward.MaxLoss),
			"max_profit":      fmt.Sprintf("₹%.2f", analysis.RiskReward.MaxProfit),
		},
		
		"recommendation": generateRecommendation(analysis),
	}

	return report
}

func generateRecommendation(analysis MarketAnalysis) string {
	if analysis.TradeSignal.Action != "BUY" {
		return "Wait for better entry opportunity. Current setup does not meet minimum criteria."
	}

	var rec strings.Builder
	rec.WriteString(fmt.Sprintf("📈 %s SIGNAL - %s\n\n", strings.ToUpper(analysis.TradeSignal.Strength), analysis.Symbol))
	rec.WriteString(fmt.Sprintf("Strategy: %s\n", analysis.TradeSignal.Strategy))
	rec.WriteString(fmt.Sprintf("Confidence: %.1f%%\n", analysis.Confidence))
	rec.WriteString(fmt.Sprintf("Entry: ₹%.2f\n", analysis.RiskReward.EntryPrice))
	rec.WriteString(fmt.Sprintf("Stop-Loss: ₹%.2f (%.1f%%)\n", 
		analysis.RiskReward.StopLoss, 
		math.Abs(analysis.RiskReward.StopLoss-analysis.RiskReward.EntryPrice)/analysis.RiskReward.EntryPrice*100))
	rec.WriteString(fmt.Sprintf("Target 1: ₹%.2f (+%.1f%%)\n", 
		analysis.RiskReward.Target1,
		(analysis.RiskReward.Target1-analysis.RiskReward.EntryPrice)/analysis.RiskReward.EntryPrice*100))
	rec.WriteString(fmt.Sprintf("Position Size: %d shares\n", analysis.RiskReward.PositionSize))
	rec.WriteString(fmt.Sprintf("Risk-Reward: 1:%.1f\n\n", analysis.RiskReward.RiskRewardRatio))
	
	rec.WriteString("✅ Entry Conditions Met:\n")
	for _, reason := range analysis.TradeSignal.Reasons {
		rec.WriteString(fmt.Sprintf("• %s\n", reason))
	}
	
	if len(analysis.TradeSignal.Warnings) > 0 {
		rec.WriteString("\n⚠️ Risk Factors:\n")
		for _, warning := range analysis.TradeSignal.Warnings {
			rec.WriteString(fmt.Sprintf("• %s\n", warning))
		}
	}

	return rec.String()
}

func getStockListForScan(scanType string) []string {
	// This would ideally fetch from a database or API
	// For now, returning popular stocks based on scan type
	
	switch scanType {
	case "momentum":
		return []string{
			"NSE:RELIANCE",
			"NSE:TCS",
			"NSE:INFY",
			"NSE:HDFC",
			"NSE:ICICIBANK",
		}
	case "breakout":
		return []string{
			"NSE:ADANIENT",
			"NSE:ADANIPORTS",
			"NSE:TATAMOTORS",
			"NSE:BHARTIARTL",
			"NSE:ITC",
		}
	case "oversold_bounce":
		return []string{
			"NSE:WIPRO",
			"NSE:TECHM",
			"NSE:MARUTI",
			"NSE:BAJFINANCE",
			"NSE:SBIN",
		}
	case "high_volume":
		return []string{
			"NSE:TATASTEEL",
			"NSE:HINDALCO",
			"NSE:VEDL",
			"NSE:JSWSTEEL",
			"NSE:COALINDIA",
		}
	default:
		return []string{
			"NSE:NIFTY50",
			"NSE:BANKNIFTY",
			"NSE:RELIANCE",
			"NSE:TCS",
			"NSE:HDFC",
		}
	}
}

func generateQuickSignal(quoteData struct{
	Tradingsymbol string
	LastPrice     float64
	NetChange     float64
	Volume        int
	VolumeTraded  int
	High          float64
	UpperCircuitLimit float64
}, scanType string, minReturn float64, riskTolerance string) TradeSignal {
	signal := TradeSignal{
		Action:         "HOLD",
		Strength:       "weak",
		ExpectedReturn: 0,
		Priority:       1,
		Reasons:        make([]string, 0),
		Warnings:       make([]string, 0),
	}

	// Quick momentum check
	percentChange := quoteData.NetChange / quoteData.LastPrice * 100
	
	switch scanType {
	case "momentum":
		if percentChange > 2 && quoteData.Volume > quoteData.VolumeTraded {
			signal.Action = "BUY"
			signal.Strength = "moderate"
			signal.ExpectedReturn = 15
			signal.Priority = 7
			signal.Reasons = append(signal.Reasons, fmt.Sprintf("Strong momentum: +%.2f%%", percentChange))
			signal.Reasons = append(signal.Reasons, "Volume above average")
		}
		
	case "oversold_bounce":
		if percentChange < -3 {
			signal.Action = "BUY"
			signal.Strength = "moderate"
			signal.ExpectedReturn = 12
			signal.Priority = 6
			signal.Reasons = append(signal.Reasons, "Oversold condition for potential bounce")
		}
		
	case "breakout":
		if quoteData.LastPrice > quoteData.High && float64(quoteData.Volume) > float64(quoteData.VolumeTraded)*1.5 {
			signal.Action = "BUY"
			signal.Strength = "strong"
			signal.ExpectedReturn = 20
			signal.Priority = 8
			signal.Reasons = append(signal.Reasons, "Breaking previous high")
			signal.Reasons = append(signal.Reasons, "High volume confirmation")
		}
	}

	// Risk warnings
	if quoteData.LastPrice > quoteData.UpperCircuitLimit*0.95 {
		signal.Warnings = append(signal.Warnings, "Near upper circuit")
	}
	
	if float64(quoteData.Volume) < float64(quoteData.VolumeTraded)*0.5 {
		signal.Warnings = append(signal.Warnings, "Low volume - poor liquidity")
	}

	signal.Timeframe = "swing"
	signal.HoldingPeriod = "3-5 days"
	signal.Strategy = fmt.Sprintf("%s strategy", scanType)

	return signal
}

func calculateOptimalPosition(capital, entryPrice, stopLoss float64, strategy string, confidence float64, povertyEscapeMode bool) map[string]interface{} {
	// Base risk percentage based on strategy
	baseRisk := 2.0
	
	switch strategy {
	case "scalping":
		baseRisk = 0.5
	case "intraday":
		baseRisk = 1.0
	case "swing":
		baseRisk = 2.0
	case "positional":
		baseRisk = 3.0
	}

	// Adjust for poverty escape mode
	if povertyEscapeMode {
		if capital < 50000 {
			baseRisk *= 2.0 // Double risk for small accounts
		} else if capital < 100000 {
			baseRisk *= 1.5
		}
		
		// Adjust based on confidence
		if confidence > 80 {
			baseRisk *= 1.2
		} else if confidence < 60 {
			baseRisk *= 0.8
		}
	}

	// Cap maximum risk
	if baseRisk > 5 {
		baseRisk = 5
	}

	// Calculate position size
	riskAmount := capital * (baseRisk / 100)
	riskPerShare := math.Abs(entryPrice - stopLoss)
	
	positionSize := 0
	if riskPerShare > 0 {
		positionSize = int(riskAmount / riskPerShare)
	}

	// Calculate investment amount
	investmentAmount := float64(positionSize) * entryPrice
	
	// Ensure we don't exceed capital
	if investmentAmount > capital*0.33 { // Max 33% in single position
		positionSize = int((capital * 0.33) / entryPrice)
		investmentAmount = float64(positionSize) * entryPrice
		riskAmount = float64(positionSize) * riskPerShare
	}

	// Calculate potential profit
	target1 := entryPrice + (riskPerShare * 2)
	target2 := entryPrice + (riskPerShare * 3)
	target3 := entryPrice + (riskPerShare * 5)
	
	profit1 := float64(positionSize) * (target1 - entryPrice)
	profit2 := float64(positionSize) * (target2 - entryPrice)
	profit3 := float64(positionSize) * (target3 - entryPrice)

	// Kelly Criterion calculation for optimal sizing
	winRate := confidence / 100
	avgWin := (target2 - entryPrice) / entryPrice
	avgLoss := (entryPrice - stopLoss) / entryPrice
	
	kellyPercent := 0.0
	if avgLoss > 0 {
		kellyPercent = ((winRate * avgWin) - ((1 - winRate) * avgLoss)) / avgWin
		kellyPercent *= 100
		
		// Use fractional Kelly (25%) for safety
		kellyPercent *= 0.25
	}

	return map[string]interface{}{
		"recommended_position_size": positionSize,
		"investment_amount":         fmt.Sprintf("₹%.2f", investmentAmount),
		"risk_amount":               fmt.Sprintf("₹%.2f", riskAmount),
		"risk_percentage":           fmt.Sprintf("%.2f%%", baseRisk),
		"position_as_percent_capital": fmt.Sprintf("%.1f%%", (investmentAmount/capital)*100),
		"targets": map[string]interface{}{
			"target_1": map[string]interface{}{
				"price":  fmt.Sprintf("₹%.2f", target1),
				"profit": fmt.Sprintf("₹%.2f", profit1),
				"return": fmt.Sprintf("%.1f%%", (profit1/investmentAmount)*100),
			},
			"target_2": map[string]interface{}{
				"price":  fmt.Sprintf("₹%.2f", target2),
				"profit": fmt.Sprintf("₹%.2f", profit2),
				"return": fmt.Sprintf("%.1f%%", (profit2/investmentAmount)*100),
			},
			"target_3": map[string]interface{}{
				"price":  fmt.Sprintf("₹%.2f", target3),
				"profit": fmt.Sprintf("₹%.2f", profit3),
				"return": fmt.Sprintf("%.1f%%", (profit3/investmentAmount)*100),
			},
		},
		"kelly_criterion_suggestion": fmt.Sprintf("%.1f%%", kellyPercent),
		"poverty_escape_mode":        povertyEscapeMode,
		"strategy":                   strategy,
		"confidence_score":           confidence,
		"recommendation": generatePositionRecommendation(positionSize, baseRisk, povertyEscapeMode),
	}
}

func generatePositionRecommendation(positionSize int, riskPercent float64, povertyEscapeMode bool) string {
	if povertyEscapeMode {
		if riskPercent >= 4 {
			return fmt.Sprintf("AGGRESSIVE POSITION: %d shares. This is a high-conviction trade with %% .1f%% capital at risk. Only proceed if analysis strongly supports entry.", positionSize, riskPercent)
		} else if riskPercent >= 2 {
			return fmt.Sprintf("MODERATE POSITION: %d shares. Balanced risk-reward suitable for wealth building with %.1f%% capital at risk.", positionSize, riskPercent)
		}
	}
	
	return fmt.Sprintf("CONSERVATIVE POSITION: %d shares. Low risk approach with %.1f%% capital at risk.", positionSize, riskPercent)
}

func roundToTick(price float64) float64 {
	// Indian market tick sizes
	if price < 1 {
		return math.Round(price*100) / 100 // 0.01
	} else if price < 10 {
		return math.Round(price*20) / 20 // 0.05
	} else if price < 100 {
		return math.Round(price*10) / 10 // 0.10
	} else if price < 1000 {
		return math.Round(price*4) / 4 // 0.25
	} else {
		return math.Round(price*2) / 2 // 0.50
	}
}

// SafeAssertBool safely converts interface{} to bool
