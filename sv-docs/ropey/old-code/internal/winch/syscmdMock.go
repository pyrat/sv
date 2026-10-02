package winch

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type SysCmdMock struct {
	mock.Mock
}

func (m *SysCmdMock) Reboot(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
