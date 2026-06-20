package show

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShow(t *testing.T) {
	bashTestPaths := map[string]string{
		"simpleArguments": "bashtests/testSimpleArguments.sh",
	}

	for _, testPath := range bashTestPaths {
		reviewBashTests(t, testPath)
	}
}

func reviewBashTests(t *testing.T, relativePath string) {
	_, err := os.Stat(relativePath)
	if errors.Is(err, os.ErrNotExist) {
		t.Fatalf("file do not exist: %s", relativePath)
	}

	bashSimpleArgumentsTestSyntaxis := exec.Command("bash", "-s", relativePath)
	if output, err := bashSimpleArgumentsTestSyntaxis.CombinedOutput(); err != nil {
		t.Fatalf(
			"wrong syntax of internal/show/test/%s:\n%s\nerr:\n%v",
			relativePath,
			string(output),
			err,
		)
	}

	bashSimpleArgumentsTest := exec.Command("bash", relativePath)
	output, err := bashSimpleArgumentsTest.CombinedOutput()
	assert.NoError(t, err, "\n%s", string(output))
}
