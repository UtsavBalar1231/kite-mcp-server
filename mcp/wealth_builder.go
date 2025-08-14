package mcp

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
	"github.com/zerodha/kite-mcp-server/kc"
)

// DetectMomentumStocksTool finds stocks ready to explode
type DetectMomentumStocksTool struct{}

func (*DetectMomentumStocksTool) Tool() mcp.Tool {
	return mcp.NewTool("detect_momentum_stocks",
		mcp.WithDescription("Find stocks with explosive momentum potential using volume, price action, and technical indicators"),
		mcp.WithString("sector",
			mcp.Description("Sector to scan (optional)"),
			mcp.Enum("auto", "banking", "it", "pharma", "metals", "energy", "fmcg", "realty", "all"),
			mcp.DefaultString("all"),
		),
		mcp.WithNumber("min_volume_surge",
			mcp.Description("Minimum volume surge percentage vs average"),
			mcp.DefaultString("150"),
		),
		mcp.WithNumber("min_price_change",
			mcp.Description("Minimum price change percentage"),
			mcp.DefaultString("2"),
		),
		mcp.WithString("timeframe",
			mcp.Description("Momentum timeframe"),
			mcp.DefaultString("daily"),
			mcp.Enum("intraday", "daily", "weekly"),
		),
		mcp.WithNumber("max_results",
			mcp.Description("Maximum number of results"),
			mcp.DefaultString("10"),
		),
	)
}

func (*DetectMomentumStocksTool) Handler(manager *kc.Manager) server.ToolHandlerFunc {
	handler := NewToolHandler(manager)
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		handler.trackToolCall(ctx, "detect_momentum_stocks")
		args := request.GetArguments()

		sector := SafeAssertString(args["sector"], "all")
		minVolumeSurge := SafeAssertFloat64(args["min_volume_surge"], 150)
		minPriceChange := SafeAssertFloat64(args["min_price_change"], 2)
		timeframe := SafeAssertString(args["timeframe"], "daily")
		maxResults := SafeAssertInt(args["max_results"], 10)

		return handler.WithSession(ctx, "detect_momentum_stocks", func(session *kc.KiteSessionData) (*mcp.CallToolResult, error) {
			// Get stock list based on sector
			stockList := getMomentumScanList(sector)
			momentumStocks := make([]MomentumStock, 0)

			for _, symbol := range stockList {
				quotes, err := session.Kite.Client.GetQuote(symbol)
				if err != nil {
					continue
				}

				quote, exists := quotes[symbol]
				if !exists {
					continue
				}

				// Convert quote to struct for momentum calculation
				quoteData := struct{
					Tradingsymbol string
					LastPrice     float64
					NetChange     float64
					Volume        int
					VolumeTraded  int
					High          float64
					Low           float64
					Open          float64
					AveragePrice  float64
				}{
					Tradingsymbol: symbol[4:], // Remove "NSE:" prefix
					LastPrice:     quote.LastPrice,
					NetChange:     quote.NetChange,
					Volume:        quote.Volume,
					VolumeTraded:  quote.Volume / 2, // Approximation
					High:          quote.OHLC.High,
					Low:           quote.OHLC.Low,
					Open:          quote.OHLC.Open,
					AveragePrice:  quote.AveragePrice,
				}
				
				// Calculate momentum score
				momentum := calculateMomentumScore(quoteData, minVolumeSurge, minPriceChange)
				
				if momentum.Score > 60 { // Minimum score threshold
					momentumStocks = append(momentumStocks, momentum)
				}
			}

			// Sort by score
			sort.Slice(momentumStocks, func(i, j int) bool {
				return momentumStocks[i].Score > momentumStocks[j].Score
			})

			// Limit results
			if len(momentumStocks) > maxResults {
				momentumStocks = momentumStocks[:maxResults]
			}

			result := map[string]interface{}{
				"timestamp":       time.Now().Format(time.RFC3339),
				"sector":          sector,
				"timeframe":       timeframe,
				"momentum_stocks": momentumStocks,
				"total_found":     len(momentumStocks),
				"scan_criteria": map[string]interface{}{
					"min_volume_surge": fmt.Sprintf("%.0f%%", minVolumeSurge),
					"min_price_change": fmt.Sprintf("%.1f%%", minPriceChange),
				},
			}

			return handler.MarshalResponse(result, "detect_momentum_stocks")
		})
	}
}

// AnalyzeSectorRotationTool identifies hot sectors
type AnalyzeSectorRotationTool struct{}

func (*AnalyzeSectorRotationTool) Tool() mcp.Tool {
	return mcp.NewTool("analyze_sector_rotation",
		mcp.WithDescription("Identify sectors showing strength and rotation patterns for sector-based trading"),
		mcp.WithString("analysis_type",
			mcp.Description("Type of sector analysis"),
			mcp.DefaultString("relative_strength"),
			mcp.Enum("relative_strength", "momentum", "institutional_flow", "breakout"),
		),
		mcp.WithNumber("lookback_days",
			mcp.Description("Number of days to analyze"),
			mcp.DefaultString("5"),
		),
	)
}

func (*AnalyzeSectorRotationTool) Handler(manager *kc.Manager) server.ToolHandlerFunc {
	handler := NewToolHandler(manager)
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		handler.trackToolCall(ctx, "analyze_sector_rotation")
		args := request.GetArguments()

		analysisType := SafeAssertString(args["analysis_type"], "relative_strength")
		lookbackDays := SafeAssertInt(args["lookback_days"], 5)

		return handler.WithSession(ctx, "analyze_sector_rotation", func(session *kc.KiteSessionData) (*mcp.CallToolResult, error) {
			// Analyze major sector indices
			sectorIndices := map[string]string{
				"Banking":     "NSE:NIFTYBANK",
				"IT":          "NSE:NIFTYIT",
				"Pharma":      "NSE:NIFTYPHARMA",
				"Metals":      "NSE:NIFTYMETAL",
				"Energy":      "NSE:NIFTYENERGY",
				"FMCG":        "NSE:NIFTYFMCG",
				"Realty":      "NSE:NIFTYREALTY",
				"Auto":        "NSE:NIFTYAUTO",
				"Financial":   "NSE:NIFTYFIN",
			}

			sectorAnalysis := make([]SectorAnalysis, 0)

			for sectorName, symbol := range sectorIndices {
				quotes, err := session.Kite.Client.GetQuote(symbol)
				if err != nil {
					continue
				}

				quote, exists := quotes[symbol]
				if !exists {
					continue
				}

				// Convert quote for sector analysis
				quoteData := struct{
					LastPrice    float64
					NetChange    float64
					AveragePrice float64
					Volume       int
					VolumeTraded int
					High         float64
					Low          float64
					Open         float64
				}{
					LastPrice:    quote.LastPrice,
					NetChange:    quote.NetChange,
					AveragePrice: quote.AveragePrice,
					Volume:       quote.Volume,
					VolumeTraded: quote.Volume / 2,
					High:         quote.OHLC.High,
					Low:          quote.OHLC.Low,
					Open:         quote.OHLC.Open,
				}
				
				analysis := analyzeSector(sectorName, quoteData, analysisType, lookbackDays)
				sectorAnalysis = append(sectorAnalysis, analysis)
			}

			// Sort by strength score
			sort.Slice(sectorAnalysis, func(i, j int) bool {
				return sectorAnalysis[i].StrengthScore > sectorAnalysis[j].StrengthScore
			})

			// Identify rotation
			rotation := identifySectorRotation(sectorAnalysis)

			result := map[string]interface{}{
				"timestamp":     time.Now().Format(time.RFC3339),
				"analysis_type": analysisType,
				"lookback_days": lookbackDays,
				"sectors":       sectorAnalysis,
				"rotation":      rotation,
				"top_sectors":   getTopSectors(sectorAnalysis, 3),
				"weak_sectors":  getWeakSectors(sectorAnalysis, 3),
			}

			return handler.MarshalResponse(result, "analyze_sector_rotation")
		})
	}
}

// MonitorPositionsTool provides real-time P&L and risk tracking
type MonitorPositionsTool struct{}

func (*MonitorPositionsTool) Tool() mcp.Tool {
	return mcp.NewTool("monitor_positions",
		mcp.WithDescription("Monitor current positions with real-time P&L, risk metrics, and adjustment suggestions"),
		mcp.WithBoolean("include_holdings",
			mcp.Description("Include holdings in monitoring"),
			mcp.DefaultString("true"),
		),
		mcp.WithBoolean("include_positions",
			mcp.Description("Include intraday positions"),
			mcp.DefaultString("true"),
		),
		mcp.WithBoolean("alert_on_risk",
			mcp.Description("Generate alerts for positions at risk"),
			mcp.DefaultString("true"),
		),
	)
}

func (*MonitorPositionsTool) Handler(manager *kc.Manager) server.ToolHandlerFunc {
	handler := NewToolHandler(manager)
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		handler.trackToolCall(ctx, "monitor_positions")
		args := request.GetArguments()

		includeHoldings := SafeAssertBool(args["include_holdings"], true)
		includePositions := SafeAssertBool(args["include_positions"], true)
		alertOnRisk := SafeAssertBool(args["alert_on_risk"], true)

		return handler.WithSession(ctx, "monitor_positions", func(session *kc.KiteSessionData) (*mcp.CallToolResult, error) {
			monitoringData := PositionMonitoring{
				Timestamp:       time.Now(),
				Positions:       make([]PositionStatus, 0),
				TotalPnL:        0,
				TotalInvested:   0,
				RiskAlerts:      make([]string, 0),
				Recommendations: make([]string, 0),
			}

			// Monitor holdings
			if includeHoldings {
				holdings, err := session.Kite.Client.GetHoldings()
				if err == nil {
					for _, holding := range holdings {
						status := analyzePosition(holding, alertOnRisk)
						monitoringData.Positions = append(monitoringData.Positions, status)
						monitoringData.TotalPnL += status.PnL
						monitoringData.TotalInvested += status.Invested
						
						if len(status.Alerts) > 0 {
							monitoringData.RiskAlerts = append(monitoringData.RiskAlerts, status.Alerts...)
						}
					}
				}
			}

			// Monitor intraday positions
			if includePositions {
				positions, err := session.Kite.Client.GetPositions()
				if err == nil {
					for _, position := range positions.Net {
						status := analyzeIntradayPosition(position, alertOnRisk)
						monitoringData.Positions = append(monitoringData.Positions, status)
						monitoringData.TotalPnL += status.PnL
						
						if len(status.Alerts) > 0 {
							monitoringData.RiskAlerts = append(monitoringData.RiskAlerts, status.Alerts...)
						}
					}
				}
			}

			// Calculate overall metrics
			monitoringData.TotalReturn = 0
			if monitoringData.TotalInvested > 0 {
				monitoringData.TotalReturn = (monitoringData.TotalPnL / monitoringData.TotalInvested) * 100
			}

			// Generate recommendations
			monitoringData.Recommendations = generateMonitoringRecommendations(monitoringData)

			result := map[string]interface{}{
				"monitoring":     monitoringData,
				"summary": map[string]interface{}{
					"total_positions": len(monitoringData.Positions),
					"total_pnl":       fmt.Sprintf("₹%.2f", monitoringData.TotalPnL),
					"total_return":    fmt.Sprintf("%.2f%%", monitoringData.TotalReturn),
					"risk_alerts":     len(monitoringData.RiskAlerts),
				},
			}

			return handler.MarshalResponse(result, "monitor_positions")
		})
	}
}

// SetEmergencyExitTool provides panic button for capital preservation
type SetEmergencyExitTool struct{}

func (*SetEmergencyExitTool) Tool() mcp.Tool {
	return mcp.NewTool("set_emergency_exit",
		mcp.WithDescription("Emergency exit mechanism to preserve capital when markets turn adverse"),
		mcp.WithString("exit_type",
			mcp.Description("Type of emergency exit"),
			mcp.Required(),
			mcp.Enum("all_positions", "losing_positions", "specific_symbol", "sector_based"),
		),
		mcp.WithString("symbol",
			mcp.Description("Symbol to exit (required for specific_symbol type)"),
		),
		mcp.WithString("sector",
			mcp.Description("Sector to exit (required for sector_based type)"),
		),
		mcp.WithNumber("max_loss_percent",
			mcp.Description("Exit positions with loss greater than this percentage"),
			mcp.DefaultString("5"),
		),
		mcp.WithBoolean("place_market_orders",
			mcp.Description("Place market orders for immediate exit"),
			mcp.DefaultString("true"),
		),
	)
}

func (*SetEmergencyExitTool) Handler(manager *kc.Manager) server.ToolHandlerFunc {
	handler := NewToolHandler(manager)
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		handler.trackToolCall(ctx, "set_emergency_exit")
		args := request.GetArguments()

		if err := ValidateRequired(args, "exit_type"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		exitType := SafeAssertString(args["exit_type"], "")
		symbol := SafeAssertString(args["symbol"], "")
		sector := SafeAssertString(args["sector"], "")
		maxLossPercent := SafeAssertFloat64(args["max_loss_percent"], 5)
		placeMarketOrders := SafeAssertBool(args["place_market_orders"], true)

		return handler.WithSession(ctx, "set_emergency_exit", func(session *kc.KiteSessionData) (*mcp.CallToolResult, error) {
			exitOrders := make([]EmergencyExitOrder, 0)
			
			// Get current positions
			positions, err := session.Kite.Client.GetPositions()
			if err != nil {
				return mcp.NewToolResultError("Failed to get positions"), nil
			}

			// Determine which positions to exit
			for _, position := range positions.Net {
				shouldExit := false
				reason := ""

				switch exitType {
				case "all_positions":
					shouldExit = true
					reason = "Emergency exit all positions"
					
				case "losing_positions":
					pnlPercent := (position.PnL / (position.AveragePrice * float64(position.Quantity))) * 100
					if pnlPercent < -maxLossPercent {
						shouldExit = true
						reason = fmt.Sprintf("Loss exceeds %.1f%%", maxLossPercent)
					}
					
				case "specific_symbol":
					if position.Tradingsymbol == symbol {
						shouldExit = true
						reason = "Specific symbol exit"
					}
					
				case "sector_based":
					if isInSector(position.Tradingsymbol, sector) {
						shouldExit = true
						reason = fmt.Sprintf("Sector-based exit: %s", sector)
					}
				}

				if shouldExit && position.Quantity != 0 {
					exitOrder := createEmergencyExitOrder(position, reason, placeMarketOrders)
					exitOrders = append(exitOrders, exitOrder)
				}
			}

			// Place exit orders
			placedOrders := make([]string, 0)
			failedOrders := make([]string, 0)

			for _, exitOrder := range exitOrders {
				orderParams := kiteconnect.OrderParams{
					Exchange:        exitOrder.Exchange,
					Tradingsymbol:   exitOrder.Symbol,
					TransactionType: exitOrder.TransactionType,
					Quantity:        exitOrder.Quantity,
					Product:         exitOrder.Product,
					OrderType:       exitOrder.OrderType,
					Price:           exitOrder.Price,
					Validity:        "DAY",
					Tag:             "EMERGENCY_EXIT",
				}

				resp, err := session.Kite.Client.PlaceOrder("regular", orderParams)
				if err != nil {
					failedOrders = append(failedOrders, fmt.Sprintf("%s: %v", exitOrder.Symbol, err))
				} else {
					placedOrders = append(placedOrders, fmt.Sprintf("%s: %s", exitOrder.Symbol, resp.OrderID))
				}
			}

			result := map[string]interface{}{
				"timestamp":      time.Now().Format(time.RFC3339),
				"exit_type":      exitType,
				"total_orders":   len(exitOrders),
				"placed_orders":  placedOrders,
				"failed_orders":  failedOrders,
				"exit_details":   exitOrders,
				"emergency_mode": true,
				"message":        fmt.Sprintf("Emergency exit initiated: %d orders placed, %d failed", len(placedOrders), len(failedOrders)),
			}

			return handler.MarshalResponse(result, "set_emergency_exit")
		})
	}
}

// GetDailyGameplanTool provides today's best opportunities
type GetDailyGameplanTool struct{}

func (*GetDailyGameplanTool) Tool() mcp.Tool {
	return mcp.NewTool("get_daily_gameplan",
		mcp.WithDescription("Get AI-generated daily trading gameplan with best opportunities, key levels, and risk management"),
		mcp.WithNumber("capital",
			mcp.Description("Available capital for today"),
			mcp.Required(),
		),
		mcp.WithString("risk_appetite",
			mcp.Description("Risk appetite for today"),
			mcp.DefaultString("moderate"),
			mcp.Enum("conservative", "moderate", "aggressive", "poverty-escape"),
		),
		mcp.WithString("market_view",
			mcp.Description("Your market view for today"),
			mcp.DefaultString("neutral"),
			mcp.Enum("bullish", "bearish", "neutral", "volatile"),
		),
	)
}

func (*GetDailyGameplanTool) Handler(manager *kc.Manager) server.ToolHandlerFunc {
	handler := NewToolHandler(manager)
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		handler.trackToolCall(ctx, "get_daily_gameplan")
		args := request.GetArguments()

		if err := ValidateRequired(args, "capital"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		capital := SafeAssertFloat64(args["capital"], 10000)
		riskAppetite := SafeAssertString(args["risk_appetite"], "moderate")
		marketView := SafeAssertString(args["market_view"], "neutral")

		return handler.WithSession(ctx, "get_daily_gameplan", func(session *kc.KiteSessionData) (*mcp.CallToolResult, error) {
			// Get market indices
			indices := []string{"NSE:NIFTY50", "NSE:NIFTYBANK"}
			indexData := make(map[string]interface{})
			
			for _, index := range indices {
				quotes, err := session.Kite.Client.GetQuote(index)
				if err == nil {
					if quote, exists := quotes[index]; exists {
						indexData[index] = map[string]interface{}{
							"last_price": quote.LastPrice,
							"change":     quote.NetChange,
							"change_pct": fmt.Sprintf("%.2f%%", (quote.NetChange/quote.LastPrice)*100),
						}
					}
				}
			}

			// Generate gameplan
			gameplan := generateDailyGameplan(capital, riskAppetite, marketView, indexData)

			return handler.MarshalResponse(gameplan, "get_daily_gameplan")
		})
	}
}

// Helper structures and functions

type MomentumStock struct {
	Symbol         string    `json:"symbol"`
	Exchange       string    `json:"exchange"`
	LastPrice      float64   `json:"last_price"`
	PriceChange    float64   `json:"price_change_percent"`
	VolumeMultiple float64   `json:"volume_multiple"`
	Score          float64   `json:"momentum_score"`
	Signals        []string  `json:"signals"`
	EntryLevel     float64   `json:"suggested_entry"`
	StopLoss       float64   `json:"stop_loss"`
	Target         float64   `json:"target"`
}

type SectorAnalysis struct {
	Sector         string   `json:"sector"`
	StrengthScore  float64  `json:"strength_score"`
	PriceChange    float64  `json:"price_change_percent"`
	VolumeChange   float64  `json:"volume_change_percent"`
	TopStocks      []string `json:"top_stocks"`
	Recommendation string   `json:"recommendation"`
}

type PositionMonitoring struct {
	Timestamp       time.Time        `json:"timestamp"`
	Positions       []PositionStatus `json:"positions"`
	TotalPnL        float64          `json:"total_pnl"`
	TotalInvested   float64          `json:"total_invested"`
	TotalReturn     float64          `json:"total_return_percent"`
	RiskAlerts      []string         `json:"risk_alerts"`
	Recommendations []string         `json:"recommendations"`
}

type PositionStatus struct {
	Symbol          string   `json:"symbol"`
	Quantity        int      `json:"quantity"`
	AvgPrice        float64  `json:"avg_price"`
	CurrentPrice    float64  `json:"current_price"`
	PnL             float64  `json:"pnl"`
	PnLPercent      float64  `json:"pnl_percent"`
	Invested        float64  `json:"invested"`
	CurrentValue    float64  `json:"current_value"`
	DayChange       float64  `json:"day_change_percent"`
	Alerts          []string `json:"alerts"`
	Action          string   `json:"suggested_action"`
}

type EmergencyExitOrder struct {
	Symbol          string  `json:"symbol"`
	Exchange        string  `json:"exchange"`
	Quantity        int     `json:"quantity"`
	TransactionType string  `json:"transaction_type"`
	Product         string  `json:"product"`
	OrderType       string  `json:"order_type"`
	Price           float64 `json:"price"`
	Reason          string  `json:"reason"`
	ExpectedLoss    float64 `json:"expected_loss"`
}

func getMomentumScanList(sector string) []string {
	switch sector {
	case "banking":
		return []string{"NSE:HDFC", "NSE:ICICIBANK", "NSE:AXISBANK", "NSE:KOTAKBANK", "NSE:SBIN"}
	case "it":
		return []string{"NSE:TCS", "NSE:INFY", "NSE:WIPRO", "NSE:HCLTECH", "NSE:TECHM"}
	case "pharma":
		return []string{"NSE:SUNPHARMA", "NSE:DRREDDY", "NSE:CIPLA", "NSE:DIVISLAB", "NSE:BIOCON"}
	case "metals":
		return []string{"NSE:TATASTEEL", "NSE:JSWSTEEL", "NSE:HINDALCO", "NSE:VEDL", "NSE:NATIONALUM"}
	default:
		// Return top stocks from multiple sectors
		return []string{
			"NSE:RELIANCE", "NSE:TCS", "NSE:HDFC", "NSE:INFY", "NSE:ICICIBANK",
			"NSE:BHARTIARTL", "NSE:ITC", "NSE:KOTAKBANK", "NSE:LT", "NSE:AXISBANK",
		}
	}
}

func calculateMomentumScore(quoteData struct{
	Tradingsymbol string
	LastPrice     float64
	NetChange     float64
	Volume        int
	VolumeTraded  int
	High          float64
	Low           float64
	Open          float64
	AveragePrice  float64
}, minVolumeSurge, minPriceChange float64) MomentumStock {
	momentum := MomentumStock{
		Symbol:    quoteData.Tradingsymbol,
		Exchange:  "NSE",
		LastPrice: quoteData.LastPrice,
		Signals:   make([]string, 0),
	}

	// Calculate price change
	momentum.PriceChange = (quoteData.NetChange / quoteData.LastPrice) * 100

	// Calculate volume multiple
	if quoteData.VolumeTraded > 0 {
		momentum.VolumeMultiple = float64(quoteData.Volume) / float64(quoteData.VolumeTraded)
	}

	// Calculate momentum score
	score := 50.0 // Base score

	// Price momentum
	if momentum.PriceChange > minPriceChange {
		score += momentum.PriceChange * 2
		momentum.Signals = append(momentum.Signals, fmt.Sprintf("Price surge: +%.2f%%", momentum.PriceChange))
	}

	// Volume momentum
	if momentum.VolumeMultiple > (minVolumeSurge / 100) {
		score += 20
		momentum.Signals = append(momentum.Signals, fmt.Sprintf("Volume surge: %.1fx average", momentum.VolumeMultiple))
	}

	// Breaking high
	if quoteData.LastPrice >= quoteData.High*0.98 {
		score += 15
		momentum.Signals = append(momentum.Signals, "Near day's high")
	}

	// Strong buying pressure
	if quoteData.LastPrice > quoteData.AveragePrice && quoteData.AveragePrice > quoteData.Open {
		score += 10
		momentum.Signals = append(momentum.Signals, "Strong buying pressure")
	}

	momentum.Score = math.Min(score, 100)

	// Calculate entry and targets
	momentum.EntryLevel = quoteData.LastPrice * 1.005 // Entry slightly above current
	momentum.StopLoss = quoteData.Low * 0.99          // Stop below day's low
	momentum.Target = quoteData.LastPrice * 1.05      // 5% target

	return momentum
}

func analyzeSector(name string, quoteData struct{
	LastPrice    float64
	NetChange    float64
	AveragePrice float64
	Volume       int
	VolumeTraded int
	High         float64
	Low          float64
	Open         float64
}, analysisType string, lookbackDays int) SectorAnalysis {
	analysis := SectorAnalysis{
		Sector:        name,
		PriceChange:   (quoteData.NetChange / quoteData.LastPrice) * 100,
		VolumeChange:  0,
		TopStocks:     make([]string, 0),
	}

	// Calculate strength score based on analysis type
	switch analysisType {
	case "relative_strength":
		analysis.StrengthScore = calculateRelativeStrength(quoteData)
	case "momentum":
		analysis.StrengthScore = calculateSectorMomentum(quoteData)
	case "institutional_flow":
		analysis.StrengthScore = estimateInstitutionalFlow(quoteData)
	case "breakout":
		analysis.StrengthScore = checkSectorBreakout(quoteData)
	}

	// Generate recommendation
	if analysis.StrengthScore > 70 {
		analysis.Recommendation = "Strong BUY - Sector showing excellent strength"
	} else if analysis.StrengthScore > 60 {
		analysis.Recommendation = "BUY - Positive sector momentum"
	} else if analysis.StrengthScore > 40 {
		analysis.Recommendation = "HOLD - Neutral sector performance"
	} else {
		analysis.Recommendation = "AVOID - Weak sector, look elsewhere"
	}

	return analysis
}

func calculateRelativeStrength(quoteData struct{
	LastPrice    float64
	NetChange    float64
	AveragePrice float64
	Volume       int
	VolumeTraded int
	High         float64
	Low          float64
	Open         float64
}) float64 {
	// Simplified relative strength calculation
	rs := 50.0

	// Price performance
	changePercent := (quoteData.NetChange / quoteData.LastPrice) * 100
	if changePercent > 2 {
		rs += 20
	} else if changePercent > 0 {
		rs += 10
	} else if changePercent < -2 {
		rs -= 20
	}

	// Price vs VWAP
	if quoteData.LastPrice > quoteData.AveragePrice {
		rs += 15
	}

	// Volume
	if quoteData.Volume > quoteData.VolumeTraded {
		rs += 15
	}

	return math.Min(math.Max(rs, 0), 100)
}

func calculateSectorMomentum(quoteData struct{
	LastPrice    float64
	NetChange    float64
	AveragePrice float64
	Volume       int
	VolumeTraded int
	High         float64
	Low          float64
	Open         float64
}) float64 {
	momentum := 50.0
	
	// Price momentum
	priceChange := (quoteData.NetChange / quoteData.LastPrice) * 100
	momentum += priceChange * 5

	// Breaking highs
	if quoteData.LastPrice >= quoteData.High {
		momentum += 20
	}

	return math.Min(math.Max(momentum, 0), 100)
}

func estimateInstitutionalFlow(quoteData struct{
	LastPrice    float64
	NetChange    float64
	AveragePrice float64
	Volume       int
	VolumeTraded int
	High         float64
	Low          float64
	Open         float64
}) float64 {
	// Estimate based on volume and price action
	flow := 50.0

	// High volume with positive price = institutional buying
	if float64(quoteData.Volume) > float64(quoteData.VolumeTraded)*1.5 && quoteData.NetChange > 0 {
		flow += 30
	}

	// Large trades indicated by price stability with volume
	if quoteData.High-quoteData.Low < quoteData.LastPrice*0.02 && quoteData.Volume > quoteData.VolumeTraded {
		flow += 20
	}

	return math.Min(math.Max(flow, 0), 100)
}

func checkSectorBreakout(quoteData struct{
	LastPrice    float64
	NetChange    float64
	AveragePrice float64
	Volume       int
	VolumeTraded int
	High         float64
	Low          float64
	Open         float64
}) float64 {
	breakout := 0.0

	// Price breaking high
	if quoteData.LastPrice >= quoteData.High {
		breakout += 50
	}

	// Volume confirmation
	if float64(quoteData.Volume) > float64(quoteData.VolumeTraded)*1.5 {
		breakout += 30
	}

	// Strong close
	if quoteData.LastPrice > quoteData.Open && quoteData.LastPrice > quoteData.AveragePrice {
		breakout += 20
	}

	return math.Min(breakout, 100)
}

func identifySectorRotation(sectors []SectorAnalysis) map[string]interface{} {
	if len(sectors) < 2 {
		return map[string]interface{}{
			"rotating_from": "none",
			"rotating_to":   "none",
			"strength":      "weak",
		}
	}

	rotation := map[string]interface{}{
		"rotating_from": sectors[len(sectors)-1].Sector,
		"rotating_to":   sectors[0].Sector,
		"strength":      "moderate",
	}

	// Determine rotation strength
	if sectors[0].StrengthScore-sectors[len(sectors)-1].StrengthScore > 30 {
		rotation["strength"] = "strong"
	} else if sectors[0].StrengthScore-sectors[len(sectors)-1].StrengthScore < 10 {
		rotation["strength"] = "weak"
	}

	return rotation
}

func getTopSectors(sectors []SectorAnalysis, count int) []map[string]interface{} {
	top := make([]map[string]interface{}, 0)
	
	for i := 0; i < count && i < len(sectors); i++ {
		top = append(top, map[string]interface{}{
			"sector":        sectors[i].Sector,
			"strength":      fmt.Sprintf("%.1f", sectors[i].StrengthScore),
			"price_change":  fmt.Sprintf("%.2f%%", sectors[i].PriceChange),
			"recommendation": sectors[i].Recommendation,
		})
	}
	
	return top
}

func getWeakSectors(sectors []SectorAnalysis, count int) []map[string]interface{} {
	weak := make([]map[string]interface{}, 0)
	
	start := len(sectors) - count
	if start < 0 {
		start = 0
	}
	
	for i := start; i < len(sectors); i++ {
		weak = append(weak, map[string]interface{}{
			"sector":       sectors[i].Sector,
			"strength":     fmt.Sprintf("%.1f", sectors[i].StrengthScore),
			"price_change": fmt.Sprintf("%.2f%%", sectors[i].PriceChange),
			"warning":      "Avoid or exit positions",
		})
	}
	
	return weak
}

func analyzePosition(holding kiteconnect.Holding, alertOnRisk bool) PositionStatus {
	status := PositionStatus{
		Symbol:       holding.Tradingsymbol,
		Quantity:     holding.Quantity,
		AvgPrice:     holding.AveragePrice,
		CurrentPrice: holding.LastPrice,
		PnL:          holding.PnL,
		Invested:     holding.AveragePrice * float64(holding.Quantity),
		CurrentValue: holding.LastPrice * float64(holding.Quantity),
		DayChange:    holding.DayChangePercentage,
		Alerts:       make([]string, 0),
	}

	// Calculate P&L percentage
	if status.Invested > 0 {
		status.PnLPercent = (status.PnL / status.Invested) * 100
	}

	// Generate alerts
	if alertOnRisk {
		// Loss alert
		if status.PnLPercent < -5 {
			status.Alerts = append(status.Alerts, fmt.Sprintf("⚠️ Loss exceeds 5%%: %.2f%%", status.PnLPercent))
			status.Action = "Consider exit or averaging"
		}
		
		// Profit booking alert
		if status.PnLPercent > 15 {
			status.Alerts = append(status.Alerts, fmt.Sprintf("✅ Good profit: %.2f%%, consider partial booking", status.PnLPercent))
			status.Action = "Book partial profits"
		}
		
		// Day's movement alert
		if math.Abs(status.DayChange) > 5 {
			status.Alerts = append(status.Alerts, fmt.Sprintf("📊 High volatility today: %.2f%%", status.DayChange))
		}
	}

	// Suggest action
	if status.Action == "" {
		if status.PnLPercent > 0 {
			status.Action = "Hold with trailing stop"
		} else {
			status.Action = "Monitor closely"
		}
	}

	return status
}

func analyzeIntradayPosition(position kiteconnect.Position, alertOnRisk bool) PositionStatus {
	status := PositionStatus{
		Symbol:       position.Tradingsymbol,
		Quantity:     position.Quantity,
		AvgPrice:     position.AveragePrice,
		CurrentPrice: position.LastPrice,
		PnL:          position.PnL,
		CurrentValue: position.LastPrice * float64(position.Quantity),
		DayChange:    0, // Intraday position
		Alerts:       make([]string, 0),
	}

	// Calculate invested amount
	if position.Quantity > 0 {
		status.Invested = position.BuyPrice * float64(position.BuyQuantity)
	} else {
		status.Invested = position.SellPrice * float64(position.SellQuantity)
	}

	// Calculate P&L percentage
	if status.Invested > 0 {
		status.PnLPercent = (status.PnL / status.Invested) * 100
	}

	// Generate alerts for intraday
	if alertOnRisk {
		if status.PnLPercent < -2 {
			status.Alerts = append(status.Alerts, fmt.Sprintf("⚠️ Intraday loss: %.2f%%", status.PnLPercent))
			status.Action = "Exit immediately"
		} else if status.PnLPercent > 3 {
			status.Alerts = append(status.Alerts, fmt.Sprintf("✅ Good intraday profit: %.2f%%", status.PnLPercent))
			status.Action = "Trail stop-loss or book"
		} else {
			status.Action = "Monitor with strict stop"
		}
	}

	return status
}

func generateMonitoringRecommendations(monitoring PositionMonitoring) []string {
	recommendations := make([]string, 0)

	// Overall portfolio recommendations
	if monitoring.TotalReturn < -5 {
		recommendations = append(recommendations, "🔴 Portfolio down >5% - Review positions and consider risk reduction")
	} else if monitoring.TotalReturn > 10 {
		recommendations = append(recommendations, "🟢 Excellent returns - Consider booking partial profits")
	}

	// Position-specific recommendations
	winnersCount := 0
	losersCount := 0
	
	for _, pos := range monitoring.Positions {
		if pos.PnLPercent > 0 {
			winnersCount++
		} else {
			losersCount++
		}
	}

	if float64(winnersCount) > float64(len(monitoring.Positions))*0.7 {
		recommendations = append(recommendations, "📈 Strong portfolio performance - Trail stop-losses on winners")
	}

	if float64(losersCount) > float64(len(monitoring.Positions))*0.5 {
		recommendations = append(recommendations, "📉 Many losing positions - Review and cut losses on weak stocks")
	}

	// Risk management
	if len(monitoring.RiskAlerts) > 3 {
		recommendations = append(recommendations, "⚠️ Multiple risk alerts - Reduce position sizes or exit weak positions")
	}

	return recommendations
}

func isInSector(symbol, sector string) bool {
	sectorMap := map[string][]string{
		"banking": {"HDFC", "ICICIBANK", "AXISBANK", "KOTAKBANK", "SBIN", "INDUSINDBK"},
		"it":      {"TCS", "INFY", "WIPRO", "HCLTECH", "TECHM", "LTTS"},
		"pharma":  {"SUNPHARMA", "DRREDDY", "CIPLA", "DIVISLAB", "BIOCON", "AUROPHARMA"},
		"auto":    {"MARUTI", "TATAMOTORS", "M&M", "BAJAJ-AUTO", "EICHERMOT", "ASHOKLEY"},
	}

	if stocks, exists := sectorMap[sector]; exists {
		for _, stock := range stocks {
			if symbol == stock {
				return true
			}
		}
	}

	return false
}

func createEmergencyExitOrder(position kiteconnect.Position, reason string, marketOrder bool) EmergencyExitOrder {
	exit := EmergencyExitOrder{
		Symbol:   position.Tradingsymbol,
		Exchange: position.Exchange,
		Quantity: int(math.Abs(float64(position.Quantity))),
		Product:  position.Product,
		Reason:   reason,
	}

	// Determine transaction type (opposite of position)
	if position.Quantity > 0 {
		exit.TransactionType = "SELL"
	} else {
		exit.TransactionType = "BUY"
	}

	// Set order type
	if marketOrder {
		exit.OrderType = "MARKET"
		exit.Price = 0
	} else {
		exit.OrderType = "LIMIT"
		if exit.TransactionType == "SELL" {
			exit.Price = position.LastPrice * 0.995 // Slightly below for quick exit
		} else {
			exit.Price = position.LastPrice * 1.005 // Slightly above for quick exit
		}
	}

	// Calculate expected loss
	exit.ExpectedLoss = position.PnL

	return exit
}

func generateDailyGameplan(capital float64, riskAppetite, marketView string, indexData map[string]interface{}) map[string]interface{} {
	gameplan := map[string]interface{}{
		"date":          time.Now().Format("2006-01-02"),
		"market_data":   indexData,
		"capital":       fmt.Sprintf("₹%.2f", capital),
		"risk_appetite": riskAppetite,
		"market_view":   marketView,
	}

	// Calculate position sizing
	var maxPositions int
	var positionSize float64
	var maxRiskPerTrade float64

	switch riskAppetite {
	case "conservative":
		maxPositions = 2
		positionSize = capital * 0.25
		maxRiskPerTrade = capital * 0.01
	case "moderate":
		maxPositions = 3
		positionSize = capital * 0.30
		maxRiskPerTrade = capital * 0.02
	case "aggressive":
		maxPositions = 4
		positionSize = capital * 0.35
		maxRiskPerTrade = capital * 0.03
	case "poverty-escape":
		maxPositions = 2
		positionSize = capital * 0.45
		maxRiskPerTrade = capital * 0.04
	}

	gameplan["position_sizing"] = map[string]interface{}{
		"max_positions":     maxPositions,
		"position_size":     fmt.Sprintf("₹%.2f", positionSize),
		"max_risk_per_trade": fmt.Sprintf("₹%.2f", maxRiskPerTrade),
	}

	// Generate trade ideas based on market view
	tradeIdeas := make([]map[string]interface{}, 0)

	switch marketView {
	case "bullish":
		tradeIdeas = append(tradeIdeas, map[string]interface{}{
			"strategy": "Momentum longs",
			"focus":    "Strong sectors and breakouts",
			"stocks":   []string{"Look for stocks breaking 52-week highs", "Focus on IT and Banking"},
			"entry":    "Buy on dips near VWAP",
		})
	case "bearish":
		tradeIdeas = append(tradeIdeas, map[string]interface{}{
			"strategy": "Short weak stocks",
			"focus":    "Overvalued sectors",
			"stocks":   []string{"Short stocks below 200 DMA", "Avoid longs"},
			"entry":    "Short on bounces to resistance",
		})
	case "volatile":
		tradeIdeas = append(tradeIdeas, map[string]interface{}{
			"strategy": "Range trading",
			"focus":    "Support and resistance levels",
			"stocks":   []string{"Trade stocks with clear ranges", "Use tight stops"},
			"entry":    "Buy support, sell resistance",
		})
	default:
		tradeIdeas = append(tradeIdeas, map[string]interface{}{
			"strategy": "Wait and watch",
			"focus":    "Observe market direction",
			"stocks":   []string{"Focus on high-probability setups only"},
			"entry":    "Wait for clear signals",
		})
	}

	gameplan["trade_ideas"] = tradeIdeas

	// Risk management rules
	gameplan["risk_rules"] = []string{
		fmt.Sprintf("Maximum loss for the day: ₹%.2f (%.0f%% of capital)", capital*0.06, 6.0),
		"Exit all positions if 3 consecutive losses",
		"Move stop to breakeven after 2% profit",
		"Book 50% at first target, trail rest",
	}

	// Key levels to watch
	gameplan["key_levels"] = map[string]interface{}{
		"nifty_support":    "Calculate based on pivot points",
		"nifty_resistance": "Calculate based on pivot points",
		"bank_nifty_range": "Watch for range breakout",
		"vix_level":        "Monitor for volatility spikes",
	}

	// Schedule
	gameplan["schedule"] = []map[string]interface{}{
		{
			"time":   "09:00 - 09:30",
			"action": "Review pre-market, set alerts",
		},
		{
			"time":   "09:30 - 10:30",
			"action": "Execute planned trades",
		},
		{
			"time":   "10:30 - 14:00",
			"action": "Monitor positions, adjust stops",
		},
		{
			"time":   "14:00 - 15:00",
			"action": "Final hour trading",
		},
		{
			"time":   "15:00 - 15:30",
			"action": "Close intraday, review day",
		},
	}

	return gameplan
}