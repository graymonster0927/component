package restful_formater

import (
	"fmt"
	"strings"
	"testing"
)

func TestFormatter(t *testing.T) {
	tests := []struct {
		name     string
		apis     []string
		expected string
	}{
		{
			name: "simple path",
			apis: []string{
				"/user/profile",
				"/user/settings",
			},
			expected: "API Tree:\n" +
				"user\n" +
				"├── profile\n" +
				"└── settings\n",
		},
		{
			name: "nested paths",
			apis: []string{
				"/api/v1/posts/list",
				"/api/v1/posts/create",
				"/api/v1/users/profile",
				"/api/v1/users/settings",
			},
			expected: "API Tree:\n" +
				"api\n" +
				"└── v1\n" +
				"    ├── posts\n" +
				"    │   ├── list\n" +
				"    │   └── create\n" +
				"    └── users\n" +
				"        ├── profile\n" +
				"        └── settings\n",
		},
		{
			name: "empty and special paths",
			apis: []string{
				"",
				"/",
				"//",
				"/user///profile",
			},
			expected: "API Tree:\n" +
				"user\n" +
				"└── profile\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := GetFormatter()

			// Record all APIs
			for _, api := range tt.apis {
				err := formatter.RecordAPI(api)
				if err != nil {
					t.Errorf("RecordAPI() error = %v", err)
					return
				}
			}

			// Get the string representation
			result := formatter.String()

			// Compare with expected output
			if !compareStrings(result, tt.expected) {
				t.Errorf("String() got:\n%v\nwant:\n%v", result, tt.expected)
			}
			formatter.Clear()
		})
	}
}

// compareStrings compares two strings ignoring differences in line endings
func compareStrings(got, want string) bool {
	// Normalize line endings
	got = strings.ReplaceAll(got, "\r\n", "\n")
	want = strings.ReplaceAll(want, "\r\n", "\n")
	return got == want
}

func TestFormatterConcurrency(t *testing.T) {
	formatter := GetFormatter()
	done := make(chan bool)

	// 并发写入测试
	for i := 0; i < 10; i++ {
		go func(i int) {
			apis := []string{
				"/api/v1/users/profile",
				"/api/v1/posts/list",
				"/api/v2/auth/login",
			}

			for _, api := range apis {
				err := formatter.RecordAPI(api)
				if err != nil {
					t.Errorf("Concurrent RecordAPI() error = %v", err)
				}
			}
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证最终结果包含预期的节点
	result := formatter.String()
	expectedNodes := []string{
		"api",
		"v1",
		"v2",
		"users",
		"posts",
		"auth",
		"profile",
		"list",
		"login",
	}

	for _, node := range expectedNodes {
		if !strings.Contains(result, node) {
			t.Errorf("Expected node %s not found in result:\n%s", node, result)
		}
	}
}

func TestWaitingQueue(t *testing.T) {
	formatter := GetFormatter()
	formatter.Clear()

	// 设置较小的阈值以便测试
	formatter = WithThreshold(3)

	// 创建大量并发请求
	concurrentCount := 100
	done := make(chan bool, concurrentCount)

	// 模拟大量并发请求
	for i := 0; i < concurrentCount; i++ {
		go func(i int) {
			api := "/api/v1/users/" + string(rune('A'+i%26)) + "/profile"
			err := formatter.RecordAPI(api)
			if err != nil && err.Error() != "too many requests" {
				t.Errorf("Unexpected error: %v", err)
			}
			done <- true
		}(i)
	}

	// 等待所有请求完成
	for i := 0; i < concurrentCount; i++ {
		<-done
	}

	// 验证结果
	result := formatter.String()
	if !strings.Contains(result, "users") {
		t.Error("Expected to find 'users' in the result")
	}
}

func TestThresholdSetting(t *testing.T) {
	tests := []struct {
		name      string
		threshold int
		apis      []string
		expectErr bool
	}{
		{
			name:      "negative threshold",
			threshold: -1,
			apis:      []string{"/api/test"},
			expectErr: false,
		},
		{
			name:      "zero threshold",
			threshold: 0,
			apis:      []string{"/api/test"},
			expectErr: false,
		},
		{
			name:      "normal threshold",
			threshold: 5,
			apis:      []string{"/api/test"},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := GetFormatter()
			formatter.Clear()
			formatter = WithThreshold(tt.threshold)

			for _, api := range tt.apis {
				err := formatter.RecordAPI(api)
				if (err != nil) != tt.expectErr {
					t.Errorf("RecordAPI() error = %v, expectErr %v", err, tt.expectErr)
				}
			}
		})
	}
}

func TestHighConcurrencyWithThreshold(t *testing.T) {
	formatter := GetFormatter()
	formatter.Clear()
	formatter = WithThreshold(5)

	// 创建大量并发请求，模拟高并发场景
	concurrentCount := 200
	requestsPerGoroutine := 5
	done := make(chan bool, concurrentCount)

	start := make(chan bool) // 用于同步所有goroutine的开始时间

	// 启动并发goroutine
	for i := 0; i < concurrentCount; i++ {
		go func(i int) {
			<-start // 等待开始信号
			for j := 0; j < requestsPerGoroutine; j++ {
				api := "/api/v1/users/" + string(rune('A'+i%26)) + "/profile/" + string(rune('0'+j))
				err := formatter.RecordAPI(api)
				if err != nil && err.Error() != "too many requests" {
					t.Errorf("Unexpected error: %v", err)
				}
				api = "/api/v1/users/G/profile/0/G/profile/0"
				err = formatter.RecordAPI(api)
			}
			done <- true
		}(i)
	}

	// 发送开始信号
	close(start)

	// 等待所有请求完成
	for i := 0; i < concurrentCount; i++ {
		<-done
	}
	// 验证结果
	result := formatter.String()
	if !strings.Contains(result, "users") {
		t.Error("Expected to find 'users' in the result")
	}
	// 尝试扫描RESTful模式
	patterns, err := formatter.ScanRestfulPattern()
	if err != nil {
		t.Errorf("ScanRestfulPattern failed: %v", err)
	}

	// 验证是否找到了RESTful模式
	if len(patterns) == 0 {
		t.Error("Expected to find some RESTful patterns")
	}
}

func TestScanRestfulPattern(t *testing.T) {
	tests := []struct {
		name         string
		threshold    int
		apis         []string
		wantPatterns []string
		wantErr      bool
	}{
		{
			name:      "basic restful pattern",
			threshold: 3,
			apis: []string{
				"/api/v1/users/1",
				"/api/v1/users/2",
				"/api/v1/users/3",
				"/api/v1/users/4",
				"/api/v1/posts/1",
				"/api/v1/posts/2",
				"/api/v1/posts/3",
			},
			wantPatterns: []string{
				"api/v1/users/*",
				"api/v1/posts/*",
			},
			wantErr: false,
		},
		{
			name:      "nested restful pattern",
			threshold: 3,
			apis: []string{
				"/api/v1/users/1/posts/1",
				"/api/v1/users/1/posts/2",
				"/api/v1/users/1/posts/3",
				"/api/v1/users/2/posts/1",
				"/api/v1/users/2/posts/2",
				"/api/v1/users/2/posts/3",
			},
			wantPatterns: []string{
				"api/v1/users/1/posts/*",
				"api/v1/users/2/posts/*",
			},
			wantErr: false,
		},
		{
			name:      "mixed patterns with different depths",
			threshold: 3,
			apis: []string{
				"/api/v1/users/1/profile",
				"/api/v1/users/2/profile",
				"/api/v1/users/3/profile",
				"/api/v1/users/1/posts/1/comments",
				"/api/v1/users/1/posts/2/comments",
				"/api/v1/users/1/posts/3/comments",
			},
			wantPatterns: []string{
				"api/v1/users/*/profile",
				"api/v1/users/1/posts/*/comments",
			},
			wantErr: false,
		},
		{
			name:      "below threshold",
			threshold: 5,
			apis: []string{
				"/api/v1/users/1",
				"/api/v1/users/2",
				"/api/v1/users/3", // 只有3个，低于阈值5
			},
			wantPatterns: []string{},
			wantErr:      false,
		},
		{
			name:         "empty input",
			threshold:    3,
			apis:         []string{},
			wantPatterns: []string{},
			wantErr:      false,
		},
		{
			name:      "single path",
			threshold: 3,
			apis: []string{
				"/api/v1/users/1",
			},
			wantPatterns: []string{},
			wantErr:      false,
		},
		{
			name:      "with special characters",
			threshold: 3,
			apis: []string{
				"/api/v1/users/1!@#/profile",
				"/api/v1/users/2$%^/profile",
				"/api/v1/users/3&*()/profile",
				"/api/v1/users/4-_+/profile",
			},
			wantPatterns: []string{
				"api/v1/users/*/profile",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := GetFormatter()
			formatter.Clear()
			formatter = WithThreshold(tt.threshold)

			// 记录所有API路径
			for _, api := range tt.apis {
				err := formatter.RecordAPI(api)
				if err != nil {
					t.Errorf("RecordAPI() error = %v", err)
					return
				}
			}

			// 执行扫描
			gotPatterns, err := formatter.ScanRestfulPattern()
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanRestfulPattern() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 验证结果
			if !compareRestfulPatterns(gotPatterns, tt.wantPatterns) {
				t.Errorf("ScanRestfulPattern() got = %v, want %v", gotPatterns, tt.wantPatterns)
			}
		})
	}
}

func TestConcurrentScanRestfulPattern(t *testing.T) {
	formatter := GetFormatter()
	formatter.Clear()
	formatter = WithThreshold(3)
	formatter = WithWaitingList(3)
	// 并发写入和扫描
	concurrentCount := 100
	done := make(chan bool, concurrentCount)

	// 启动写入goroutine
	for i := 0; i < concurrentCount; i++ {
		go func(i int) {
			// 每个goroutine写入一组相关的API
			base := i * 5 // 每组5个API
			for j := 0; j < 5; j++ {
				api := fmt.Sprintf("/api/v1/users/%d/posts/%d", i, base+j)
				err := formatter.RecordAPI(api)
				if err != nil && err.Error() != "too many requests" {
					t.Errorf("RecordAPI() error = %v", err)
				}
			}
			done <- true
		}(i)
	}

	// 同时启动扫描goroutine
	go func() {
		_, err := formatter.ScanRestfulPattern()
		if err != nil {
			t.Errorf("ScanRestfulPattern() error = %v", err)
		}
		done <- true
	}()

	// 等待所有操作完成
	for i := 0; i < concurrentCount+1; i++ { // +1 是因为有一个扫描goroutine
		<-done
	}

	// 最后验证结果
	patterns, err := formatter.ScanRestfulPattern()
	if err != nil {
		t.Errorf("Final ScanRestfulPattern() error = %v", err)
	}

	// 验证是否找到了RESTful模式
	if len(patterns) == 0 {
		t.Error("Expected to find some RESTful patterns")
	}
}

// compareRestfulPatterns 比较两个RESTful模式切片是否相等（忽略顺序）
func compareRestfulPatterns(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}

	gotMap := make(map[string]bool)
	for _, pattern := range got {
		gotMap[pattern] = true
	}

	for _, pattern := range want {
		if !gotMap[pattern] {
			return false
		}
	}

	return true
}
