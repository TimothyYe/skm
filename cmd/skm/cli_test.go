package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	uuid "github.com/satori/go.uuid"
)

func TestUsage(t *testing.T) {
	tmp := prepareTest(t)
	defer os.RemoveAll(tmp)
	cmd := exec.Command("skm", "-h")
	_, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal("Expected exit code 1 bot 0")
	}
}

func TestInvalidArgs(t *testing.T) {
	tmp := prepareTest(t)
	expectString := "No help topic for 'hogehoge'\n"
	defer os.RemoveAll(tmp)
	cmd := exec.Command("skm", "hogehoge")
	b, _ := cmd.CombinedOutput()

	if expectString != string(b) {
		t.Fatalf("Expected string is : %s", expectString)
	}
}

func prepareTest(t *testing.T) (tmpPath string) {
	tmp := os.TempDir()
	tmp = filepath.Join(tmp, uuid.NewV4().String())
	runCmd(t, "go", "build", "-o", filepath.Join(tmp, "bin", "skm"), "github.com/TimothyYe/skm")
	os.Setenv("PATH", filepath.Join(tmp, "bin")+string(filepath.ListSeparator)+os.Getenv("PATH"))
	os.MkdirAll(filepath.Join(tmp, "src"), 0755)
	return tmp
}

func runCmd(t *testing.T, cmd string, args ...string) []byte {
	b, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		t.Fatalf("Expected %v, but %v: %v", nil, err, string(b))
	}
	return b
}
