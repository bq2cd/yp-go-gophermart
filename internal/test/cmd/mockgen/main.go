// Binary mockgen wraps original [mockgen](go.uber.org/mock/mockgen@latest) tool
// to facilitate proper placing of generated source code.
//
//nolint:lll,wrapcheck
package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"golang.org/x/mod/modfile"
)

/////////////////////////////////////////////////////////////////////////////////

const (
	PathEmpty        Path = ""
	PathBaseGoMod    Path = "go.mod"
	PathBaseInternal Path = "internal"
	PathBaseTest     Path = "test"
	PathBaseMocks    Path = "mocks"
)

type Path string

func (p Path) String() string {
	return string(p)
}

func (p Path) Base() Path {
	return Path(filepath.Base(p.String()))
}

func (p Path) Dir() Path {
	return Path(filepath.Dir(p.String()))
}

func (p Path) Join(parts ...Path) Path {
	args := append([]string{p.String()}, convertPathSlice[Path, string](parts)...)

	return Path(filepath.Join(args...))
}

func (p Path) JoinStr(parts ...string) Path {
	return p.Join(convertPathSlice[string, Path](parts)...)
}

func (p Path) TrimPrefix(path Path) Path {
	prefix := filepath.Clean(path.String())
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	return Path(strings.TrimPrefix(p.String(), prefix))
}

func (p Path) GetRelPath(target Path) Path {
	relPath, err := filepath.Rel(p.String(), target.String())
	if err != nil {
		abort(err)
	}

	return Path(relPath)
}

type FilePath struct {
	Path
}

func (p FilePath) ReadBytes() []byte {
	content, err := os.ReadFile(p.String())
	if err != nil {
		abort(err)
	}

	return content
}

func convertPathSlice[F, T ~string](input []F) []T {
	out := make([]T, 0, len(input))
	for _, item := range input {
		out = append(out, T(item))
	}

	return out
}

/////////////////////////////////////////////////////////////////////////////////

type ProjectRoot struct {
	Path
}

func (p ProjectRoot) GetGoPackage() GoPackage {
	goModPath := FilePath{p.Join(PathBaseGoMod)}

	goMod, err := modfile.Parse(goModPath.String(), goModPath.ReadBytes(), nil)
	if err != nil {
		abort(err)
	}

	return GoPackage{
		ModURL:  Path(goMod.Module.Mod.Path),
		Root:    p,
		RelPath: PathEmpty,
	}
}

/////////////////////////////////////////////////////////////////////////////////

type GoPackage struct {
	ModURL  Path
	Root    ProjectRoot
	RelPath Path
}

func (p GoPackage) String() string {
	return p.ModURL.Join(p.RelPath).String()
}

func (p GoPackage) Base() Path {
	return p.RelPath.Base()
}

func (p GoPackage) AbsPath() Path {
	return p.Root.Join(p.RelPath)
}

func (p GoPackage) SubPackage(parts ...Path) GoPackage {
	return GoPackage{
		ModURL:  p.ModURL,
		Root:    p.Root,
		RelPath: p.RelPath.Join(parts...),
	}
}

/////////////////////////////////////////////////////////////////////////////////

type Mockgen struct {
	SourcePackage  GoPackage
	Interfaces     []string
	OutputFilename string
}

func (m *Mockgen) Run() error {
	cmd := m.buildCmd()

	log.Printf("running: %v", cmd.Args)

	return cmd.Run()
}

func (m *Mockgen) buildCmd() *exec.Cmd {
	args := []string{
		"tool",
		"mockgen",
		"-typed",
		"-destination=" + m.getDestination(),
		"-package=" + m.getPackage(),
		m.SourcePackage.String(),
	}
	args = append(args, m.Interfaces...)

	cmd := exec.Command("go", args...) //nolint:noctx,gosec
	cmd.Dir = string(m.SourcePackage.AbsPath())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}

func (m *Mockgen) getDestination() string {
	destRoot := m.SourcePackage.Root.Join(PathBaseInternal, PathBaseTest, PathBaseMocks)

	sourceRelPath := m.SourcePackage.RelPath
	sourceAbsPath := m.SourcePackage.AbsPath()

	destAbsPath := destRoot.Join(sourceRelPath.TrimPrefix(PathBaseInternal), Path(m.OutputFilename))
	destRelPath := sourceAbsPath.GetRelPath(destAbsPath)

	return destRelPath.String()
}

func (m *Mockgen) getPackage() string {
	return PathBaseMocks.String()
}

/////////////////////////////////////////////////////////////////////////////////

//nolint:tagalign
type CLI struct {
	OutputFilename string   `help:"Destination filename (not path) to put generated output into" name:"outfile" short:"o" required:""`
	Interfaces     []string `help:"List of interfaces to generate mocks for"                                                          arg:""`
}

func (c *CLI) Run() error {
	mockgen := c.configureMockgen()

	return mockgen.Run()
}

func (c *CLI) configureMockgen() Mockgen {
	projRoot := c.getProjectRoot()

	rootPkg := projRoot.GetGoPackage()
	relPath := projRoot.GetRelPath(c.getOriginalCwd())

	return Mockgen{
		SourcePackage:  rootPkg.SubPackage(relPath),
		Interfaces:     c.Interfaces,
		OutputFilename: c.OutputFilename,
	}
}

func (c *CLI) getProjectRoot() ProjectRoot {
	return ProjectRoot{Path(os.Getenv("MISE_PROJECT_ROOT"))}
}

func (c *CLI) getOriginalCwd() Path {
	return Path(os.Getenv("MISE_ORIGINAL_CWD"))
}

/////////////////////////////////////////////////////////////////////////////////

func main() {
	var cli CLI

	kctx := kong.Parse(&cli)
	err := kctx.Run()

	kctx.FatalIfErrorf(err)
}

/////////////////////////////////////////////////////////////////////////////////

func abort(args ...any) {
	panic(args)
}
