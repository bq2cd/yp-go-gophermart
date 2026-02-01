package spechelpers

import (
	"fmt"
	"os"
	"path/filepath"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
)

// TempDir defines a wrapper struct that holds a function
// to get a sub-path (as a string) inside the temporary directory
// created by [CreateLocalTempDir], as well as a function
// to clean up (remove) the temporary directory.
type TempDir struct {
	PathTo  func(string) string
	Cleanup func()
}

// CreateLocalTempDir creates a local temporary directory (`tmp-N`, where `N` is parallel process number), and
// returns [TempDir] instance with two functions:
// first (`PathTo`) is to get path inside the temporary directory,
// second (`Cleanup`) is to remove the temporary directory on cleanup.
// If `--fail-fast` option is used and a spec that uses [CreateTempDir] fails,
// the temporary directory will be preserved.
// Otherwise, it is the spec responsibility to call returned [TempDirCleanupFunc] function.
func CreateLocalTempDir() *TempDir {
	g.GinkgoHelper()

	dir := fmt.Sprintf("./tmp-%d", g.GinkgoParallelProcess())

	o.Expect(os.RemoveAll(dir)).To(o.Succeed())
	o.Expect(os.MkdirAll(dir, 0750)).To(o.Succeed()) //nolint:mnd

	pathTo := func(path string) string {
		return filepath.Join(dir, path)
	}

	cleanup := func() {
		suiteConfig, _ := g.GinkgoConfiguration()
		if g.CurrentSpecReport().Failed() && suiteConfig.FailFast {
			g.GinkgoWriter.Printf("Preserving artifacts in %s\n", dir)

			return
		}

		o.Expect(os.RemoveAll(dir)).To(o.Succeed())
	}

	return &TempDir{
		PathTo:  pathTo,
		Cleanup: cleanup,
	}
}
