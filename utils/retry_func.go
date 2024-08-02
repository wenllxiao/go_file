package utils

// retryTask 重试任务
func retryTask(handFunc func() error) error {
	var (
		err     error
		retries = 3
	)
	for retries > 0 {
		if err = handFunc(); err == nil {
			break
		}
		retries--
	}
	return err
}
