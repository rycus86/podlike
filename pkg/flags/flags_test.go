package flags

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func TestReadmeIsUpToDate(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet("/podlike", flag.ContinueOnError)
	setupVariables()

	output := strOutput{}

	flag.CommandLine.SetOutput(&output)
	flag.CommandLine.Usage()

	readmeData, err := ioutil.ReadFile("../../README.md")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(readmeData), output.Text) {
		t.Error("The command like usage is not found in the README")
		fmt.Println(output.Text)
	}
}

type strOutput struct {
	Text string
}

func (s *strOutput) Write(p []byte) (n int, err error) {
	s.Text += string(p)
	return len(p), nil
}
