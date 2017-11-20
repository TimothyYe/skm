package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestUsage(t *testing.T) {
	prepareTest(t)
	defer os.Remove("$GOPATH/bin/skm")
	cmd := exec.Command("skm", "-h")
	_, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal("Expected exit code 1 bot 0")
	}
}

func TestInvalidArgs(t *testing.T) {
	prepareTest(t)
	expectString := "No help topic for 'hogehoge'\n"
	defer os.Remove("$GOPATH/bin/skm")
	cmd := exec.Command("skm", "hogehoge")
	b, _ := cmd.CombinedOutput()

	if expectString != string(b) {
		t.Fatalf("Expected string is : %s", expectString)
	}
}

func prepareTest(t *testing.T) {
	runCmd(t, "go", "install")
}

func runCmd(t *testing.T, cmd string, args ...string) []byte {
	b, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		t.Fatalf("Expected %v, but %v: %v", nil, err, string(b))
	}
	return b
}
