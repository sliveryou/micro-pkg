package retry

import (
	"time"

	"github.com/zeromicro/go-zero/core/mathx"
)

// 偏差阈值：使实际持续时间偏差在 [0.95, 1.05] 中
const deviation = 0.05

// unstable 不稳定器
var unstable = mathx.NewUnstable(deviation)

// Strategy 尝试策略函数，在每次尝试进行前调用
type Strategy func(attempt uint) bool

// Limit 限制尝试策略，限制整个执行过程的尝试次数
func Limit(attemptLimit uint) Strategy {
	return func(attempt uint) bool {
		return attempt < attemptLimit
	}
}

// Delay 延迟尝试策略，第一次尝试将在等待给定的持续时间（存在细微偏差）后进行
func Delay(duration time.Duration) Strategy {
	return func(attempt uint) bool {
		if attempt == 0 {
			time.Sleep(unstable.AroundDuration(duration))
		}

		return true
	}
}

// Wait 等待尝试策略，在每次尝试后等待给定的持续时间（存在细微偏差），
// 如果尝试次数大于持续时间列表的长度，则使用最后的持续时间
func Wait(durations ...time.Duration) Strategy {
	return func(attempt uint) bool {
		if attempt > 0 && len(durations) > 0 {
			durationIndex := int(attempt - 1)

			if len(durations) <= durationIndex {
				durationIndex = len(durations) - 1
			}

			time.Sleep(unstable.AroundDuration(durations[durationIndex]))
		}

		return true
	}
}

// FailLimit 失败尝试策略，达到一定尝试次数执行预先指定的失败方法并退出
func FailLimit(attemptLimit uint, failAction Action) Strategy {
	return func(attempt uint) bool {
		if attempt < attemptLimit {
			return true
		}

		if failAction != nil {
			_ = failAction(attempt)
		}

		return false
	}
}

// Fail 失败策略，每达到一定尝试次数执行
// 目前这个预先的策略会导致前置策略判断失败直接退出循环无法执行该策略
// 故此策略需要在Limit策略之前加入
func Fail(attemptLimit uint, failAction Action) Strategy {
	return func(attempt uint) bool {
		if attempt%attemptLimit == 0 && failAction != nil {
			_ = failAction(attempt)
		}

		return true
	}
}
