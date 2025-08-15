package types

import "fmt"

type ID uint64

func (i ID) String() string {
	return fmt.Sprintf("%d", i)
}
