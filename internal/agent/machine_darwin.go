package agent

import (
	"context"
	"errors"
	"os/exec"
	"regexp"
	"time"
)

func systemMachineIdentity() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	b, err := exec.CommandContext(ctx, "/usr/sbin/ioreg", "-rd1", "-c", "IOPlatformExpertDevice").Output()
	if err != nil {
		return "", err
	}
	m := regexp.MustCompile(`"IOPlatformUUID"\s*=\s*"([^"]+)"`).FindSubmatch(b)
	if len(m) != 2 {
		return "", errors.New("machine UUID unavailable")
	}
	return string(m[1]), nil
}
