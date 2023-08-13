package mjingo

import (
	"log"
	"testing"
)

func TestParse(t *testing.T) {
	stmt, err := parse("Hello {{ name }}", "foo.j2")
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("stmt=%+v", stmt)
}
