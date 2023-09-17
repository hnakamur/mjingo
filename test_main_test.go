package mjingo

import (
	"flag"
	"os"
	"testing"
)

var overwriteSnapshot bool

func TestMain(m *testing.M) {
	flag.BoolVar(&overwriteSnapshot, "overwrite-snapshot", false,
		"whether to overwrite test snapshot files")
	os.Exit(m.Run())
}
