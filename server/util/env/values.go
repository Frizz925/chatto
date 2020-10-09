package env

import "os"

func IsIntegrationTest() bool {
	return os.Getenv("INTEGRATION_TEST") == "1"
}

func AppMode() string {
	return os.Getenv("APP_MODE")
}

func IsAppMode(mode string) bool {
	return AppMode() == mode
}
