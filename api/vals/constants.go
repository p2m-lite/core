package vals

import "time"

const (
	// AnalysisIntervalHours (X): How often the analyzer runs
	AnalysisIntervalHours = 10 * time.Minute

	// LookbackPeriodDays (M): How many days of logs to analyze
	LookbackPeriodDays = 7

	// CooldownPeriodDays (N): How many days to wait before processing a recorder again
	CooldownPeriodDays = 4

	// Quality Thresholds
	MinPH        = 5
	MaxPH        = 9
	MaxTurbidity = 500

	// Reward Amount (in Wei)
	RewardAmount = 250_000_000_000_000 // 0.00025 BNB (₹ 20 approx)
)
