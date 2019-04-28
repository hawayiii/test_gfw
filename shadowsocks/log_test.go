package shadowsocks

import "testing"

func Test_log(t *testing.T) {
	LOG_INFO_F("1. hello world.")
	LOG_INFO("2. hello world.")
}
