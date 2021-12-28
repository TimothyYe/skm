package actions

import (
	"flag"
	"fmt"
	"github.com/TimothyYe/skm/internal/models"
	"github.com/TimothyYe/skm/internal/utils"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	cli "gopkg.in/urfave/cli.v1"
)

var random *rand.Rand

func init() {
	random = rand.New(rand.NewSource(time.Now().Unix()))
}

func TestInitAction(t *testing.T) {
	testcases := []struct {
		code  string
		info  string
		err   bool
		check func(t *testing.T, env *models.Environment)
		setup func(t *testing.T, env *models.Environment)
	}{
		{
			code: "no-preexisting-key",
			info: "If the .ssh folder is empty, then no default profile should be created",
			err:  false,
			check: func(t *testing.T, env *models.Environment) {
				assertEmptyDir(t, env.StorePath, "no-preexisting-key")
			},
		},
		{
			code: "rsa-key",
			info: "If the .ssh folder contains an RSA key-pair, it should be copied into the default profile",
			setup: func(t *testing.T, env *models.Environment) {
				os.RemoveAll(env.StorePath)
				writeFile(t, filepath.Join(env.SSHPath, "id_rsa"), []byte("private"))
				writeFile(t, filepath.Join(env.SSHPath, "id_rsa.pub"), []byte("public"))
			},
			err: false,
			check: func(t *testing.T, env *models.Environment) {
				assertFileExists(t, filepath.Join(env.StorePath, "default", "id_rsa"), "rsa")
				assertFileExists(t, filepath.Join(env.StorePath, "default", "id_rsa.pub"), "rsa")
			},
		},
		{
			code: "ed25519-key",
			info: "If the .ssh folder contains an ED25519 key-pair, it should be copied into the default profile",
			setup: func(t *testing.T, env *models.Environment) {
				os.RemoveAll(env.StorePath)
				writeFile(t, filepath.Join(env.SSHPath, "id_ed25519"), []byte("private"))
				writeFile(t, filepath.Join(env.SSHPath, "id_ed25519.pub"), []byte("public"))
			},
			err: false,
			check: func(t *testing.T, env *models.Environment) {
				assertFileExists(t, filepath.Join(env.StorePath, "default", "id_ed25519"), "ed25519")
				assertFileExists(t, filepath.Join(env.StorePath, "default", "id_ed25519.pub"), "ed25519")
			},
		},
	}

	for _, testcase := range testcases {
		env := setupEnvironment(t)
		flags := flag.NewFlagSet("", flag.ContinueOnError)
		flags.String("ssh-path", env.SSHPath, "")
		flags.String("store-path", env.StorePath, "")
		if testcase.setup != nil {
			testcase.setup(t, env)
		}
		if !t.Failed() {
			c := cli.NewContext(nil, flags, nil)
			err := Initialize(c)
			if err == nil && testcase.err {
				t.Errorf("[%s] should have returned an error", testcase.code)
			}
			if err != nil && !testcase.err {
				t.Errorf("[%s] produced an unexpected error: %s", testcase.code, err.Error())
			}
			if testcase.check != nil {
				testcase.check(t, env)
			}
		}
		tearDownEnvironment(t, env)
		if t.Failed() {
			break
		}
	}
}

func setupEnvironment(t *testing.T) *models.Environment {
	rootFolder := filepath.Join(os.TempDir(), fmt.Sprintf("skm-testdir-%d", random.Int()))
	storePath := filepath.Join(rootFolder, ".skm")
	sshPath := filepath.Join(rootFolder, ".ssh")
	if err := os.MkdirAll(storePath, 0700); err != nil {
		t.Errorf("Failed to create store path: %s", err.Error())
		return nil
	}
	if err := os.MkdirAll(sshPath, 0700); err != nil {
		t.Errorf("Failed to create ssh path: %s", err.Error())
		return nil
	}
	return &models.Environment{
		StorePath: storePath,
		SSHPath:   sshPath,
	}
}

func tearDownEnvironment(t *testing.T, env *models.Environment) {
	rootFolder := filepath.Dir(env.StorePath)
	os.RemoveAll(rootFolder)
}

func assertFileExists(t *testing.T, path string, code string) {
	_, err := os.Stat(path)
	if err != nil {
		t.Fatalf("[%s] Failed to assert file existence: %s (err: %v)", code, path, err)
	}
}

func assertEmptyDir(t *testing.T, path string, code string) {
	empty, err := utils.IsEmpty(path)
	if err != nil || !empty {
		t.Fatalf("[%s] Failed to assert empty directory %s: %v", code, path, err)
	}
}

func writeFile(t *testing.T, path string, data []byte) {
	if t.Failed() {
		return
	}
	if err := ioutil.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("Failed to write %s: %s", path, err.Error())
	}
}
