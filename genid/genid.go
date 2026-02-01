package genid

import (
	"fmt"

	"github.com/segmentio/ksuid"
)

func New(e string) string {
	return fmt.Sprintf("%s_%s", e, ksuid.New().String())
}
