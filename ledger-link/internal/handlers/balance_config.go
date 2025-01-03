package handlers

import "time"

type BalanceHandlerConfig struct {
	DefaultHistoryLimit int           `json:"default_history_limit" yaml:"default_history_limit"`
	MaxHistoryLimit     int           `json:"max_history_limit" yaml:"max_history_limit"`
	TimeFormat          string        `json:"time_format" yaml:"time_format"`
	CacheDuration       time.Duration `json:"cache_duration" yaml:"cache_duration"`
}

func DefaultBalanceHandlerConfig() *BalanceHandlerConfig {
	return &BalanceHandlerConfig{
		DefaultHistoryLimit: 100,
		MaxHistoryLimit:     1000,
		TimeFormat:          time.RFC3339,
		CacheDuration:       time.Hour,
	}
}
