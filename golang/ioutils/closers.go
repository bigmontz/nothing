package ioutils

import (
	"fmt"
	"io"
)

func SafeClose(err error, closer io.Closer) error {
	closeErr := closer.Close()
	if closeErr == nil {
		return err
	}
	if err == nil {
		return closeErr
	}
	return fmt.Errorf("close error %v occurred after %w", closeErr, err)
}
