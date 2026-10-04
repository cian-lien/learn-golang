// Package book 是一个极简"图书"业务，用来演示 OTel 在分层项目里的埋点位置：
//
//	handler（Gin 路由）  —— 由 otelgin 自动埋点，不用手写
//	service（业务逻辑）  —— 手动埋业务 span、打业务指标
//	store（模拟数据库）  —— 手动埋 IO span（SpanKind=Client）
package book

import (
	"context"
	"time"
)

// must 用于构造 OTel 仪表（instrument）：只有名字、单位不合法才会报错，
// 属于启动期就该炸的编程错误。
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// sleep 模拟耗时 IO，同时尊重 ctx 取消——真实代码里换成 DB / HTTP 调用即可。
func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
