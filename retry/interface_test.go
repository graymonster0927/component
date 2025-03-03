package retry

import (
	"testing"
)

func TestGetRetryHelper_RegisteredStrategy_ReturnsInstance(t *testing.T) {

	// 测试另一个已注册的策略
	strategy := RetryMaxTimes
	retryHelper, err := GetRetryHelper(strategy)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if retryHelper == nil {
		t.Errorf("Expected a non-nil retry helper")
	}

}

func TestGetRetryHelper_UnregisteredStrategy_ReturnsError(t *testing.T) {
	// 测试一个未注册的策略
	strategy := Strategy(999) // 假设999未注册
	retryHelper, err := GetRetryHelper(strategy)
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
	if retryHelper != nil {
		t.Errorf("Expected a nil retry helper")
	}
	if err.Error() != "strategy have not register" {
		t.Errorf("Expected error message 'strategy have not register', got %v", err.Error())
	}
}
