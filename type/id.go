package types

import (
	"errors"
	"fmt"
)

type ID uint64

func (i ID) String() string {
	return fmt.Sprintf("%d", i)
}

func (i ID) IsValid() bool {
	return i > 0
}

func (i ID) Validate() error {
	if !i.IsValid() {
		return errors.New("ID must be greater than zero")
	}
	
	return nil
}
