package store

import (
	"errors"
	"fmt"
)

var ErrRuntimeInvalidStage = errors.New("runtime invalid stage")

func invalidStagef(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrRuntimeInvalidStage, fmt.Sprintf(format, args...))
}

func IsRuntimeInvalidStage(err error) bool {
	return errors.Is(err, ErrRuntimeInvalidStage)
}
