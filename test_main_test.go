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
	flag.BoolVar(&useReflect, "use-reflect", false,
		"whether to use reflect to run filters, tests, and functions")
	os.Exit(m.Run())
}
