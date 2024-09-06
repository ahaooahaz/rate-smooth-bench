package interaction

import (
	"fmt"
	"os"
)

func Exit(err error) {
	fmt.Fprint(os.Stderr, err.Error())
	os.Exit(1)
}
