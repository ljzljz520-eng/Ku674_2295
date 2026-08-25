package config

import "testing"

func TestDefaultSettingsValidate(t *testing.T) {
	if err := Default().Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestInvalidSettings(t *testing.T) {
	s := Default()
	s.MaxBatchSize = 0
	if err := s.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}
