package service

import "time"

// ptrFloat 返回 float64 指针，供订阅相关测试使用。
func ptrFloat(v float64) *float64 { return &v }

// ptrTime 返回 time.Time 指针，供订阅相关测试使用。
func ptrTime(v time.Time) *time.Time { return &v }
