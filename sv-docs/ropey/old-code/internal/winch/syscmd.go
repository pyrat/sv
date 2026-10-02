package winch

import (
	"context"
	"os/exec"

	"github.com/pkg/errors"
)

type SysCmd interface {
	Reboot(ctx context.Context) error
}

type sysCmd struct{}

func (s *sysCmd) Reboot(ctx context.Context) error {
	cmdReboot := exec.CommandContext(ctx, "sh", "-c", "/usr/bin/nsenter --target 1 --pid --uts --ipc --net --mount reboot")

	err := cmdReboot.Start()
	if err != nil {
		return errors.Wrap(err, "cmd nsenter reboot")
	}

	return nil
}
