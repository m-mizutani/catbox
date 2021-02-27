package interfaces

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

type FileOpenFunc func(string) (io.ReadCloser, error)
type FileCreateFunc func(string) (io.WriteCloser, error)
type ExecOutputFunc func(string, ...string) ([]byte, error)
type MkDirAllFunc func(path string, perm os.FileMode) error
type TempFileFunc func(string, string) (string, error)
type ReadFileFunc func(filename string) ([]byte, error)

// DefaultExecOutput is wrapper of exec.Comand
func DefaultExecOutput(cmd string, args ...string) ([]byte, error) {
	return exec.Command(cmd, args...).Output()
}

// DefaultOpenFunc is wrapper of os.Open
func DefaultOpenFunc(path string) (io.ReadCloser, error) { return os.Open(path) }

// DefaultCreateFunc is wrapper of os.Create
func DefaultCreateFunc(path string) (io.WriteCloser, error) { return os.Create(path) }

// DefaultTempFileFunc is wrapper of iotuil.TempFile
func DefaultTempFileFunc(dir, ptn string) (string, error) {
	tmp, err := ioutil.TempFile(dir, ptn)
	if err != nil {
		return "", err
	}
	tmp.Close()
	return tmp.Name(), nil
}
