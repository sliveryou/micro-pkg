package retry

// Action 具体的行为函数
type Action func(attempt uint) error

// Retry 根据尝试策略执行具体的行为函数，
// 当执行成功时，会终止尝试，或因为尝试策略结果影响提早终止尝试
func Retry(action Action, strategies ...Strategy) error {
	var err error

	for attempt := uint(0); (attempt == 0 || err != nil) &&
		shouldAttempt(attempt, strategies...); attempt++ {
		err = action(attempt)
	}

	return err
}

// MustRetry 根据尝试策略执行具体的行为函数，
// 当执行成功时，会终止尝试，或因为尝试策略结果影响提早终止尝试
func MustRetry(action Action, strategies ...Strategy) {
	err := Retry(action, strategies...)
	if err != nil {
		panic(err)
	}
}

// shouldAttempt 判断当前尝试在给定策略下能否进行
func shouldAttempt(attempt uint, strategies ...Strategy) bool {
	shouldAttempt := true

	for i := 0; shouldAttempt && i < len(strategies); i++ {
		shouldAttempt = shouldAttempt && strategies[i](attempt)
	}

	return shouldAttempt
}
