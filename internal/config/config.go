package config

import "time"

type Settings struct {
	ListenAddress      string
	DatabasePath       string
	MaxBatchSize       int
	MoistureMinimum    float64
	MoistureMaximum    float64
	TemperatureMinimum float64
	TemperatureMaximum float64
	LightMaximum       float64
	BatteryMinimum     float64
	RetryLimit         int
	RetryWindow        time.Duration
	StaleAfter         time.Duration
}

func Default() Settings {
	return Settings{
		ListenAddress: ":8080", DatabasePath: "farm-sensors.db", MaxBatchSize: 500,
		MoistureMinimum: 0, MoistureMaximum: 100, TemperatureMinimum: -40, TemperatureMaximum: 80,
		LightMaximum: 150000, BatteryMinimum: 0, RetryLimit: 3, RetryWindow: 10 * time.Minute,
		StaleAfter: 30 * time.Minute,
	}
}

func (s Settings) Validate() error {
	if s.MaxBatchSize < 1 {
		return ErrInvalidSettings("max batch size must be positive")
	}
	if s.MoistureMinimum >= s.MoistureMaximum || s.TemperatureMinimum >= s.TemperatureMaximum {
		return ErrInvalidSettings("sensor ranges must be ordered")
	}
	if s.RetryLimit < 0 || s.RetryWindow <= 0 || s.StaleAfter <= 0 {
		return ErrInvalidSettings("retry and stale durations must be positive")
	}
	return nil
}

type settingsError string

func (e settingsError) Error() string { return string(e) }

func ErrInvalidSettings(message string) error { return settingsError(message) }
