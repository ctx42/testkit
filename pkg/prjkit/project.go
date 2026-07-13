// Package prjkit provides a Project helper for writing integration tests
// that operate on a temporary Go project on disk. A Project manages source
// files, a Go module, a git history, and optional Docker configuration.
//
// Instances follow an open/close lifecycle: configuration methods require
// the project to be open; query and execution methods require it to be
// closed. [New] registers a test cleanup that fails the test if [Project.Close]
// is not called before the test ends.
package prjkit

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/tester"
	"github.com/ctx42/xdef/pkg/xdef"

	"github.com/ctx42/testkit/pkg/dkrkit"
	"github.com/ctx42/testkit/pkg/exekit"
	"github.com/ctx42/testkit/pkg/oskit"
	"github.com/ctx42/testkit/pkg/pathkit"
)

// dockerfile contains the content of the example Dockerfile with three targets.
//
//go:embed data/Dockerfile
var dockerfile []byte

// dockerfileNEP contains the content of the example Dockerfile with three
// targets without entrypoint.
//
//go:embed data/Dockerfile-nep
var dockerfileNEP []byte

// Default test project values.
const (
	// ProjDir represents the default project directory name. This directory
	// will be created inside the project's temporary directory.
	ProjDir = "project"

	// GoModNameStem represents default Go module name stem.
	GoModNameStem = "example.com/comp/"

	// GoModName represents the default Go module name used with "go mod init".
	GoModName = GoModNameStem + ProjDir

	// GitOrigin represents git origin without the "ssh://" prefix.
	GitOrigin = "git@example.com:comp/" + ProjDir + ".git"

	// GitSSHOrigin represents git origin with the "ssh://" prefix.
	GitSSHOrigin = "ssh://" + GitOrigin

	// DkrHost represents the default Docker private repository host.
	DkrHost = "my.nexus.dev"

	// DkrRepo represents default repository name in private Docker repository.
	DkrRepo = "repo"
)

// Sentinel errors.
var (
	// ErrOpen indicates the test project instance was expected to be closed.
	ErrOpen = errors.New("expected test project instance to be closed")

	// ErrClosed indicates the test project instance was expected to be open.
	ErrClosed = errors.New("expected test project instance to be open")
)

// WithProjectCreate is an option for [New] that creates the project root
// directory. It fails the test if the root directory already exists.
func WithProjectCreate(prj *Project) {
	checkRoot(prj.t, prj.root)
	if assert.NoDirExist(prj.t, prj.root) {
		root := oskit.MkdirAll(prj.t, prj.root)
		prj.root = pathkit.EvalSymlinks(prj.t, root)
	}
}

// WithProjectEnv is option for [New] which sets the environment to use.
func WithProjectEnv(env []string) func(*Project) {
	return func(prj *Project) { prj.env = env }
}

// Project represents a test Go project.
type Project struct {
	// Absolute path to the test project root directory.
	root string

	// Name of the project directory (default: root directory name).
	//
	// If the `root` dir is `/root/project` than `dirName` is `project`.
	dirName string

	// Go module name for the project (default: [GoModName] + dirName).
	//
	// If the `root` dir is `/root/project` than `dirName` is `project` and
	// `modName` is `example.com/comp/project`.
	modName string

	// Docker private repository.
	dkrRepo string

	// Random Docker image name.
	imgName string

	// Random Docker tag.
	imgTag string

	// Function signature for removing images.
	imgRem func(string)

	// Image refs the imgRem was called with.
	removed []string

	// Absolute path to the current working directory of a test. It's set when
	// the instance is constructed and relays on the fact that nobody changed
	// the test execution directory before [New] constructor was called.
	testDir string

	// Absolute path to the temporary directory outside the project root.
	tmpDir string

	// Tracks the number of git commits in the project.
	commits int

	// Environment when executing commands (default: os.Environ).
	env []string

	gitID  bool     // True after git user identity has been configured.
	closed bool     // Instance closed for edits.
	misuse bool     // True when t.Error* or t.Fatal were called.
	t      tester.T // Test manager.
}

// New returns a new Project for the given root directory. It validates that
// root is a non-empty absolute path and registers a test cleanup that fails
// the test if [Project.Close] was not called. Returns nil if t has already
// failed when New returns.
func New(t tester.T, root string, opts ...func(*Project)) *Project {
	t.Helper()
	prj := &Project{
		root:    root,
		testDir: oskit.Getwd(t),
		t:       t,
	}
	for _, opt := range opts {
		opt(prj)
	}
	prj.dirName = filepath.Base(prj.root)
	prj.modName = GoModNameStem + prj.dirName
	prj.imgRem = func(ref string) {
		dkrkit.NewT(t).ImgRm(ref)
		prj.removed = append(prj.removed, ref)
	}
	if t.Failed() {
		return nil
	}
	checkRoot(t, prj.root)
	t.Cleanup(func() {
		t.Helper()
		if !prj.misuse && !prj.closed {
			prj.misuse = true
			t.Errorf("expected instance to be closed at the test end")
		}
	})
	return prj
}

// checkRoot checks root directory path requirements. Will panic using Fatal on
// error.
func checkRoot(t tester.T, root string) {
	t.Helper()
	if root == "" {
		t.Fatal("test project root directory must be set")
	}
	if !filepath.IsAbs(root) {
		t.Fatal("test project root directory must be the absolute path")
	}
}

// Root returns the absolute path to the project root directory.
func (prj *Project) Root() string { return prj.root }

// Path constructs path rooted at the test project root directory.
func (prj *Project) Path(elems ...string) string {
	elems = append([]string{prj.root}, elems...)
	return filepath.Join(elems...)
}

// Chdir changes directory to the project root. Returns the absolute path to
// the directory which was the current working directory when [New] constructor
// was called.
func (prj *Project) Chdir() string {
	prj.t.Helper()
	prj.CheckClosed()

	oskit.Chdir(prj.t, prj.root)
	return prj.testDir
}

// ChdirBack changes working directory back to the test working directory.
func (prj *Project) ChdirBack() {
	prj.t.Helper()
	prj.CheckClosed()

	if prj.testDir != oskit.Getwd(prj.t) {
		if err := os.Chdir(prj.testDir); err != nil {
			prj.t.Fatal(err)
		}
	}
}

// ReadFileStr reads the file at the path built from the project root and elems.
func (prj *Project) ReadFileStr(pth string, elems ...string) string {
	prj.t.Helper()
	prj.CheckClosed()

	elems = append([]string{pth}, elems...)
	return oskit.ReadFileStr(prj.t, prj.root, elems...)
}

// Exe executes command in the test project context and returns stdout and
// stderr respectively.
func (prj *Project) Exe(cmd string, args ...string) (string, string) {
	prj.t.Helper()

	return prj.exe().Exe(cmd, args...)
}

// ExeStdout executes cmd in the test project context and returns stdout.
func (prj *Project) ExeStdout(cmd string, args ...string) string {
	prj.t.Helper()

	return prj.exe().ExeStdout(cmd, args...)
}

// ExeStderr executes cmd in the test project context and returns stderr.
func (prj *Project) ExeStderr(cmd string, args ...string) string {
	prj.t.Helper()

	return prj.exe().ExeStderr(cmd, args...)
}

// exe returns executor configured with project context: working directory,
// environment.
func (prj *Project) exe() *exekit.Exe {
	opts := []func(*exekit.Exe){
		exekit.WithWd(prj.root),
		exekit.WithDetCov(os.Args),
	}
	var env []string
	if prj.env == nil {
		env = os.Environ()
	} else {
		env = prj.env
	}
	opts = append(opts, exekit.WithEnv(env))
	return exekit.New(prj.t, opts...)
}

// TempDir creates a temporary directory and returns the absolute path to it.
// Multiple calls to this method return the same value.
func (prj *Project) TempDir() string {
	if prj.tmpDir == "" {
		prj.tmpDir = prj.t.TempDir()
	}
	return prj.tmpDir
}

// CreateDir creates a directory, along with any necessary parents, rooted at
// the test project root directory. Returns the absolute path to the created
// directory.
func (prj *Project) CreateDir(elems ...string) string {
	prj.t.Helper()
	prj.CheckOpen()

	return oskit.MkdirAll(prj.t, prj.root, elems...)
}

// CreateFile creates an empty file in the test project at the path built from
// elements.
func (prj *Project) CreateFile(pth string, elems ...string) string {
	prj.t.Helper()
	return prj.CreateFileWith("", pth, elems...)
}

// CreateFileWith creates a file in the test project with content and path
// built from elements.
func (prj *Project) CreateFileWith(
	content, pth string,
	elems ...string,
) string {

	prj.t.Helper()
	prj.CheckOpen()

	full := filepath.Join(append([]string{pth}, elems...)...)
	if len(elems) > 0 {
		dirs := append([]string{pth}, elems[:len(elems)-1]...)
		prj.CreateDir(dirs...)
	}

	full = prj.Path(full)
	return oskit.Create(prj.t, content, full)
}

// Rename renames oldPath (rooted at the project) to newPath.
//
// Example:
//
//	prj.Rename("fileA.txt", "fileB.txt")
func (prj *Project) Rename(oldPath, newPath string) string {
	prj.t.Helper()
	prj.CheckOpen()

	newPath = prj.Path(newPath)
	if err := os.Rename(prj.Path(oldPath), newPath); err != nil {
		prj.t.Fatal(err)
	}
	return newPath
}

// FileFrom copies the file from src to the root of the test project.
func (prj *Project) FileFrom(src string) string {
	prj.t.Helper()
	prj.CheckOpen()

	return oskit.CopyFile(prj.t, prj.root, src)
}

// FilesFrom copies all the files (not directories) from src to the test
// project.
func (prj *Project) FilesFrom(src string) {
	prj.t.Helper()
	prj.CheckOpen()

	fn := func(src string, ent fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ent.IsDir() {
			return nil
		}
		prj.FileFrom(src)
		return nil
	}
	if err := filepath.WalkDir(src, fn); err != nil {
		prj.t.Fatal(err)
	}
}

// ProjectFrom copies project example from src directory. The dirName and
// modName fields will be set based on the directory name.
func (prj *Project) ProjectFrom(src string) {
	prj.t.Helper()
	prj.CheckOpen()

	prj.FilesFrom(src)
	prj.dirName = filepath.Base(src)
	prj.modName = filepath.Join(GoModNameStem, prj.dirName)
}

// ImpSpec returns the Go module path for the test project.
func (prj *Project) ImpSpec() string {
	prj.t.Helper()

	return prj.modName
}

// GoModInit initializes the go module in the project's root directory.
func (prj *Project) GoModInit() {
	prj.t.Helper()
	prj.CheckOpen()

	content := fmt.Sprintf("module %s\n\ngo 1.21\n", prj.modName)
	oskit.Write(prj.t, []byte(content), prj.root, "go.mod")
}

// GoModTidy runs `go mod tidy` in the project's root directory.
func (prj *Project) GoModTidy() {
	prj.t.Helper()
	prj.CheckOpen()

	exe := prj.exe()
	exe.Exe("go", "mod", "tidy")
}

// GitInitAddAll initializes git repository in project root, adds all files and
// commits them. If a tag is set, the commit will be tagged with it. Returns
// current git hash.
func (prj *Project) GitInitAddAll(tags ...string) *GitCommit {
	prj.t.Helper()
	prj.CheckOpen()

	prj.Exe("git", "init")
	var tag string
	if len(tags) == 1 {
		tag = tags[0]
	}
	return prj.GitCommit(tag, "Initial commit.")
}

// GitCommit commits all files in the project directory with an automatically
// generated commit message (different every time it's called) or one provided
// as mss[0], and if tag is not empty string, it tags it. Returns git
// repository hash after the commit.
func (prj *Project) GitCommit(tag string, mss ...string) *GitCommit {
	prj.t.Helper()
	prj.CheckOpen()

	if !prj.gitID {
		prj.Exe("git", "config", "user.email", "test@example.com")
		prj.Exe("git", "config", "user.name", "Test User")
		prj.gitID = true
	}
	prj.Exe("git", "add", "-A")
	msg := fmt.Sprintf("commit %d", prj.commits)
	if len(mss) == 1 {
		msg = mss[0]
	}
	prj.Exe("git", "commit", "-am", msg)
	prj.commits++
	if tag != "" {
		prj.Exe("git", "tag", "-a", "-m", "test tag"+tag, tag)
	}
	return prj.GitCommitLog().Latest()
}

// GitSetRemote sets git remote repository for the test project. If no remote
// is provided, the default value of [GitOrigin] will be used.
func (prj *Project) GitSetRemote(remote ...string) {
	prj.t.Helper()
	prj.CheckOpen()

	pth := GitOrigin
	if len(remote) > 0 {
		pth = remote[0]
	}

	_, eout := prj.Exe("git", "remote", "add", "origin", pth)
	if eout != "" {
		prj.t.Error(eout)
	}
}

// GitHash returns git hash. Will fail if the test project is not a git project.
func (prj *Project) GitHash() string {
	prj.t.Helper()
	prj.CheckClosed()

	if cm := prj.GitCommitLog().Latest(); cm != nil {
		return cm.Hash
	}
	return ""
}

// GitCommitLog returns all git commits from the project git local repository.
func (prj *Project) GitCommitLog() GitCommits {
	prj.t.Helper()

	opts := []func(*exekit.Exe){
		exekit.WithWd(prj.root),
		exekit.WithDetCov(os.Args),
		exekit.WithEnv(append(os.Environ(), "TZ=UTC")),
	}
	exe := exekit.New(prj.t, opts...)

	lns := exe.ExeStdout("git", "log", "--pretty=%h (%D) %at %s")
	scn := bufio.NewScanner(strings.NewReader(lns))

	var gcs []GitCommit
	for scn.Scan() {
		gc, err := NewGitCommit(scn.Text())
		if err != nil {
			prj.t.Error(err)
			return nil
		}
		gcs = append(gcs, gc)
	}
	if err := scn.Err(); err != nil {
		prj.t.Error(err)
	}
	return gcs
}

// Compile compiles the test project and returns the absolute path to the
// binary.
func (prj *Project) Compile() string {
	prj.t.Helper()
	prj.CheckClosed()

	out := filepath.Join(prj.TempDir(), "a.out")
	args := []string{"build", "-o", out}

	sout := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	cmd := exec.CommandContext(context.Background(), "go", args...)
	cmd.Stdout = sout
	cmd.Stderr = eout
	cmd.Dir = prj.root
	if err := cmd.Run(); err != nil {
		prj.t.Fatal(eout.String())
	}
	return out
}

// WithConfig creates an empty configuration file in the test project. Returns
// the absolute path to the configuration file.
func (prj *Project) WithConfig() string {
	prj.t.Helper()
	prj.CheckOpen()

	return oskit.Write(prj.t, "", prj.CreateDir("configs"), "project.conf")
}

// CfgAdd adds key-value to the project configuration file. It does not change
// the value of the variable if it already exists it adds the new key-value to
// the end of the file. Returns the absolute path to the configuration file.
func (prj *Project) CfgAdd(key, value string) string {
	prj.t.Helper()
	prj.CheckOpen()

	pth := prj.WithConfig()
	oskit.Write(prj.t, key+"="+value+"\n", pth)
	return pth
}

// CfgDkrRepoDef adds a default (DkrRepo) private Docker repository to the
// project config. Every time it's called, it will add a new key-value to the
// end of the file.
func (prj *Project) CfgDkrRepoDef() {
	prj.t.Helper()
	prj.CheckOpen()

	prj.CfgDkrRepo(DkrHost, DkrRepo)
}

// CfgDkrRepo adds a given private Docker repository to the project config.
// Every time it's called, it will add a new key-value to the end of the file.
func (prj *Project) CfgDkrRepo(host, repo string) {
	prj.t.Helper()
	prj.CheckOpen()

	repo = path.Join(host, repo)
	prj.CfgAdd(xdef.EnvDkrRegHost, host)
	prj.CfgAdd(xdef.EnvDkrRepo, repo)
	prj.dkrRepo = repo
}

// DkrRepo returns private Docker repository.
func (prj *Project) DkrRepo() string {
	prj.t.Helper()
	prj.CheckClosed()

	return prj.dkrRepo
}

// CfgDkrTargets adds Docker targets to the project config. Every time it's
// called, it will add a new key-value to the end of the file.
func (prj *Project) CfgDkrTargets(targets string) {
	prj.t.Helper()
	prj.CheckOpen()

	prj.CfgAdd(xdef.EnvDkfTargets, targets)
}

// WithDockerfile adds an example Dockerfile to the project. The Dockerfile
// defines three targets with entrypoint to simplify testing. It also generates
// random Docker image and tag values and adds removal of the images with those
// references (and "latest" tag) to test cleanup. If the private repo value is
// required, [Project.CfgDkrRepo] or [Project.CfgDkrRepoDef] method must be
// called before this one. Returns the absolute path to the Dockerfile.
func (prj *Project) WithDockerfile() string {
	prj.t.Helper()
	prj.CheckOpen()

	pth := filepath.Join(prj.Root(), "Dockerfile")
	oskit.Write(prj.t, dockerfile, pth)
	prj.DkrImgNameTagSet(dkrkit.RandName(), dkrkit.RandTag())
	return pth
}

// WithDockerfileNEP adds an example Dockerfile to the project (NEP: no
// entrypoint). The Dockerfile defines three targets without an ENTRYPOINT.
// It also generates random Docker image and tag values and adds removal of
// those images (and "latest" tag) to test cleanup. If the private repo value
// is required, [Project.CfgDkrRepo] or [Project.CfgDkrRepoDef] must be
// called before this one. Returns the absolute path to the Dockerfile.
func (prj *Project) WithDockerfileNEP() string {
	prj.t.Helper()
	prj.CheckOpen()

	pth := filepath.Join(prj.Root(), "Dockerfile")
	oskit.Write(prj.t, dockerfileNEP, pth)
	prj.DkrImgNameTagSet(dkrkit.RandName(), dkrkit.RandTag())
	return pth
}

// DkrImgNameTagSet sets the Docker image name and tag. It overrides the
// random values assigned to them by default.
func (prj *Project) DkrImgNameTagSet(name, tag string) {
	prj.t.Helper()
	prj.CheckOpen()
	prj.imgName = name
	prj.imgTag = tag
	prj.t.Cleanup(func() {
		prj.imgRem(prj.dkrImgRef())
		prj.imgRem(prj.dkrImgRefLatest())
	})
}

// DkrImgName returns a random Docker image name. Multiple calls to this method
// return the same value.
func (prj *Project) DkrImgName() string {
	prj.t.Helper()
	prj.CheckClosed()

	return prj.imgName
}

// DkrImgTag returns a random Docker tag name. Multiple calls to this method
// return the same value.
func (prj *Project) DkrImgTag() string {
	prj.t.Helper()
	prj.CheckClosed()

	return prj.imgTag
}

// DkrImgRef returns Docker image reference. Multiple calls to this method
// return the same value.
func (prj *Project) DkrImgRef() string {
	prj.t.Helper()
	prj.CheckClosed()

	return prj.dkrImgRef()
}

// dkrImgRef works the same as [Project.DkrImgRef] but doesn't check
// the [Project] is closed.
func (prj *Project) dkrImgRef() string {
	ref := dkrkit.Ref(prj.dkrRepo, prj.imgName, prj.imgTag)
	return ref
}

// DkrImgRefLatest returns Docker image reference with the "latest" tag.
// Multiple calls to this method return the same value.
func (prj *Project) DkrImgRefLatest() string {
	prj.t.Helper()
	prj.CheckClosed()

	return prj.dkrImgRefLatest()
}

// dkrImgRefLatest works the same as [Project.DkrImgRefLatest] but doesn't
// check the [Project] is closed.
func (prj *Project) dkrImgRefLatest() string {
	ref := dkrkit.Ref(prj.dkrRepo, prj.imgName, "latest")
	return ref
}

// DkrTgtName returns Docker image name for a Dockerfile target.
func (prj *Project) DkrTgtName(target string) string {
	prj.t.Helper()
	prj.CheckClosed()

	name := prj.imgName + "-" + target
	return dkrkit.Ref(prj.dkrRepo, name, "")
}

// DkrTgtRef returns Docker image reference for a Dockerfile target.
func (prj *Project) DkrTgtRef(target string) string {
	prj.t.Helper()
	prj.CheckClosed()

	name := prj.imgName + "-" + target
	ref := dkrkit.Ref(prj.dkrRepo, name, prj.imgTag)
	refLatest := dkrkit.Ref(prj.dkrRepo, name, "latest")
	prj.t.Cleanup(func() {
		prj.imgRem(ref)
		prj.imgRem(refLatest)
	})
	return ref
}

// DkrTgtRefLatest returns Docker image reference with the "latest" tag for a
// Dockerfile target.
func (prj *Project) DkrTgtRefLatest(target string) string {
	prj.t.Helper()
	prj.CheckClosed()

	name := prj.DkrImgName() + "-" + target
	tag := prj.DkrImgTag()
	ref := dkrkit.Ref(prj.dkrRepo, name, tag)
	refLatest := dkrkit.Ref(prj.dkrRepo, name, "latest")
	prj.t.Cleanup(func() {
		prj.imgRem(ref)
		prj.imgRem(refLatest)
	})
	return refLatest
}

// Close marks the project as closed and ready for use. It must be called
// once the project is fully configured. Calling Close a second time marks
// the test as failed.
func (prj *Project) Close() {
	prj.t.Helper()
	if prj.closed {
		prj.misuse = true
		prj.t.Error("instance already closed")
	}
	prj.closed = true
}

// CheckClosed returns true if the project is in the closed (ready-to-use)
// state. Otherwise it marks the test as failed and returns false.
func (prj *Project) CheckClosed() bool {
	if prj.closed {
		return true
	}
	prj.t.Helper()
	prj.misuse = true
	prj.t.Fatal(ErrOpen)
	return false
}

// CheckOpen returns true if the project is in the open (configuration)
// state. Otherwise it marks the test as failed and returns false.
func (prj *Project) CheckOpen() bool {
	if !prj.closed {
		return true
	}
	prj.t.Helper()
	prj.misuse = true
	prj.t.Fatal(ErrClosed)
	return false
}

// GitCommits is an ordered slice of git commits. Elements are ordered
// most-recent first, matching the output order of "git log --oneline".
type GitCommits []GitCommit

// Find finds commit by hash in the collection. Returns nil if the hash is not
// in the collection.
func (gcs GitCommits) Find(hash string) *GitCommit {
	for _, gc := range gcs {
		if gc.Hash == hash {
			return &gc
		}
	}
	return nil
}

// Latest returns the latest commit from the collection. Returns nil if the
// collection is empty.
//
// It's assumed that the latest commit is the first (index 0) commit in
// the collection.
func (gcs GitCommits) Latest() *GitCommit {
	if len(gcs) > 0 {
		return &gcs[0]
	}
	return nil
}

// First returns the first commit from the collection. Returns nil if the
// collection is empty.
//
// It's assumed that the last commit is the last commit in the collection.
func (gcs GitCommits) First() *GitCommit {
	if len(gcs) > 0 {
		return &gcs[len(gcs)-1]
	}
	return nil
}

// N returns nth commit. Returns nil if the collection is empty or the nth
// element is not available.
func (gcs GitCommits) N(i int) *GitCommit {
	if i < 0 || i >= len(gcs) {
		return nil
	}
	return &gcs[i]
}

// GitCommit represents single git commit.
type GitCommit struct {
	Hash    string    // Commit hash.
	Rev     string    // Commit revision (tag).
	Date    time.Time // Author date in UTC.
	Summary string    // Commit summary.
}

// ErrInvGitLogLine is returned by NewGitCommit when a git log line has an
// invalid format.
var ErrInvGitLogLine = errors.New("invalid git log line")

// Regular expressions to pars git log line.
var (
	gitLogLineRX = regexp.MustCompile(`([0-9a-f]+) \((.*)\) (\d+) (.*)`)
	gitLogTagRX  = regexp.MustCompile(`tag: (.*),?`)
)

// NewGitCommit parses git commit line. The expected line formats:
//
//	16c806f () 946782245 commit 1
//	6907f52 (HEAD -> master, tag: v0.5.11, origin/master) 946782245 Bump.
//	2728491 (HEAD -> master, tag: v1.2.3) 946782245 commit 2
//
// Returns error for line in different format.
func NewGitCommit(lin string) (GitCommit, error) {
	ms := gitLogLineRX.FindStringSubmatch(lin)
	if len(ms) != 5 {
		return GitCommit{}, ErrInvGitLogLine
	}

	var tag string
	if ms[2] != "" {
		tm := gitLogTagRX.FindStringSubmatch(ms[2])
		if len(tm) == 2 {
			tag = tm[1]
		}
	}

	ts, err := strconv.ParseInt(ms[3], 10, 64)
	if err != nil {
		return GitCommit{}, ErrInvGitLogLine
	}
	gc := GitCommit{
		Hash:    ms[1],
		Rev:     tag,
		Date:    time.Unix(ts, 0).UTC(),
		Summary: ms[4],
	}
	return gc, nil
}
