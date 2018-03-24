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

func TestCollectRawOutput(t *testing.T) {
	sampleResults := [][]string{
		{"PSU2", "80", "W"}, // Hex value that is >= 0x80
	}
	res, err := convertRawOutput(sampleResults)

	if err != nil {
		t.Error(err)
	}

	expected := float64(128)
	if res[0].value != expected {
		t.Fatalf("Expexted %f got %f", expected, res[0].value)
	}
}
