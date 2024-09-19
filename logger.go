package component

import (
	"context"
	"fmt"
)

type LoggerInterface interface {
	Debug(ctx context.Context, args ...interface{})
	Debugf(ctx context.Context, format string, args ...interface{})
	Info(ctx context.Context, args ...interface{})
	Infof(ctx context.Context, format string, args ...interface{})
	Warn(ctx context.Context, args ...interface{})
	Warnf(ctx context.Context, format string, args ...interface{})
	Error(ctx context.Context, args ...interface{})
	Errorf(ctx context.Context, format string, args ...interface{})
	Panic(ctx context.Context, args ...interface{})
	Panicf(ctx context.Context, format string, args ...interface{})
	Fatal(ctx context.Context, args ...interface{})
	Fatalf(ctx context.Context, format string, args ...interface{})
}

type DefaultLogger struct{}

func (l *DefaultLogger) Debug(ctx context.Context, args ...interface{}) {
	fmt.Print("[debug] ")
	fmt.Println(args...)
}
func (l *DefaultLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	fmt.Print("[debug] ")
	fmt.Printf(format, args...)
	fmt.Print("\n")
}

func (l *DefaultLogger) Info(ctx context.Context, args ...interface{}) {
	fmt.Print("[info] ")
	fmt.Println(args...)
}

func (l *DefaultLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	fmt.Print("[info] ")
	fmt.Printf(format, args...)
	fmt.Print("\n")
}

func (l *DefaultLogger) Warn(ctx context.Context, args ...interface{}) {
	fmt.Print("[warn] ")
	fmt.Println(args...)
}

func (l *DefaultLogger) Warnf(ctx context.Context, format string, args ...interface{}) {
	fmt.Print("[warn] ")
	fmt.Printf(format, args...)
	fmt.Print("\n")
}

func (l *DefaultLogger) Error(ctx context.Context, args ...interface{}) {
	fmt.Print("[error] ")
	fmt.Println(args...)
}

func (l *DefaultLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	fmt.Print("[error] ")
	fmt.Printf(format, args...)
	fmt.Print("\n")
}

func (l *DefaultLogger) Panic(ctx context.Context, args ...interface{}) {
	fmt.Print("[panic] ")
	fmt.Println(args...)
}

func (l *DefaultLogger) Panicf(ctx context.Context, format string, args ...interface{}) {
	fmt.Print("[panic] ")
	fmt.Printf(format, args...)
	fmt.Print("\n")
}

func (l *DefaultLogger) Fatal(ctx context.Context, args ...interface{}) {
	fmt.Print("[fatal] ")
	fmt.Println(args...)
}

func (l *DefaultLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {
	fmt.Print("[fatal] ")
	fmt.Printf(format, args...)
	fmt.Print("\n")
}

var (
	Logger LoggerInterface = &DefaultLogger{}
)

func SetLogger(l LoggerInterface) {
	Logger = l
}
