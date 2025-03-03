package retry

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestDoRetry_SuccessOnFirstTry(t *testing.T) {
	maxTimes, err := GetRetryHelper(RetryMaxTimes, WithMaxTimesMaxTimes(3), WithMaxTimesRetryTimeout(100*time.Millisecond), WithMaxTimesMaxDelay(0), WithMaxTimesIsExponential(false))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	err = maxTimes.DoRetry(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestDoRetry_FailAndRetry(t *testing.T) {
	maxTimes, err := GetRetryHelper(RetryMaxTimes, WithMaxTimesMaxTimes(3), WithMaxTimesRetryTimeout(100*time.Millisecond), WithMaxTimesMaxDelay(0), WithMaxTimesIsExponential(false))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	err = maxTimes.DoRetry(func() error {
		return errors.New("test error")
	})
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
}

func TestDoRetry_ExponentialBackoff(t *testing.T) {
	maxTimes, err := GetRetryHelper(RetryMaxTimes, WithMaxTimesMaxTimes(3), WithMaxTimesRetryTimeout(500*time.Millisecond), WithMaxTimesMaxDelay(0), WithMaxTimesIsExponential(true))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	st := time.Now()
	err = maxTimes.DoRetry(func() error {
		return errors.New("test error")
	})

	if time.Since(st) < 1000*time.Millisecond {
		t.Errorf("Expected delay of at least %v, got %v", 500*time.Millisecond, time.Since(st))
	}

	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
}

func TestDoRetryReturn_SuccessOnFirstTry(t *testing.T) {
	maxTimes, err := GetRetryHelper(RetryMaxTimes, WithMaxTimesMaxTimes(3), WithMaxTimesRetryTimeout(100*time.Millisecond), WithMaxTimesMaxDelay(0), WithMaxTimesIsExponential(false))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	ret, err := maxTimes.DoRetryReturn(func() (interface{}, error) {
		return "success", nil
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if ret != "success" {
		t.Errorf("Expected 'success', got %v", ret)
	}
}

func TestDoRetryReturn_FailAndRetry(t *testing.T) {
	maxTimes, err := GetRetryHelper(RetryMaxTimes, WithMaxTimesMaxTimes(3), WithMaxTimesRetryTimeout(100*time.Millisecond), WithMaxTimesMaxDelay(0), WithMaxTimesIsExponential(false))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	ret, err := maxTimes.DoRetryReturn(func() (interface{}, error) {
		return nil, errors.New("test error")
	})
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
	if ret != nil {
		t.Errorf("Expected nil return, got %v", ret)
	}
}

func TestGetRetryTimeout_ExponentialBackoff(t *testing.T) {
	maxTimes := NewMaxTimes(3, 100*time.Millisecond, 500*time.Millisecond, true)
	timeout := maxTimes.getRetryTimeout(3)
	expected := 100 * time.Millisecond * time.Duration(math.Pow(2, float64(2-1))) / 2
	if timeout < expected || timeout > 500*time.Millisecond {
		t.Errorf("Expected timeout between %v and %v, got %v", expected, 500*time.Millisecond, timeout)
	}

}

func TestGetRetryTimeout_FixedDelay(t *testing.T) {
	maxTimes := NewMaxTimes(3, 100*time.Millisecond, 0, false)
	timeout := maxTimes.getRetryTimeout(2)
	if timeout != 100*time.Millisecond {
		t.Errorf("Expected fixed timeout of %v, got %v", 100*time.Millisecond, timeout)
	}
}

func TestGetRetryTimeout_FixedDelay2(t *testing.T) {
	maxTimes := NewMaxTimes(3, 100*time.Millisecond, 50*time.Millisecond, true)
	timeout := maxTimes.getRetryTimeout(3)
	if timeout != 50*time.Millisecond {
		t.Errorf("Expected fixed timeout of %v, got %v", 50*time.Millisecond, timeout)
	}
}
