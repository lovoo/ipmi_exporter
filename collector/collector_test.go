package collector

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestCollector(t *testing.T) {
	f, err := os.Open("testdata/ipmi_output1.txt")
	if err != nil {
		t.Fatalf("cannot open test data: %v", err)
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatalf("reading test data failed: %v", err)
	}

	res, err := splitOutput(buf)
	if err != nil {
		t.Fatalf("parsing output failed: %v", err)
	}

	fmt.Println(res)
}
