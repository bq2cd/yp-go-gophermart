package testutil

import (
	"context"
	"fmt"
	"net"
)

// GetRandomListenAddress finds a free available port on 'localhost'.
// It does so by creating a listener which chooses such a port,
// taking its address and closing the listener.
func GetRandomListenAddress(ctx context.Context) (string, error) {
	var lnCfg net.ListenConfig

	lnr, err := lnCfg.Listen(ctx, "tcp", "localhost:0")
	if err != nil {
		return "", fmt.Errorf("cannot start listener: %w", err)
	}

	addr := lnr.Addr().String()

	err = lnr.Close()
	if err != nil {
		return "", fmt.Errorf("cannot close listener: %w", err)
	}

	return addr, nil
}
