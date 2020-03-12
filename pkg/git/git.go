package git

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path"

	"github.com/google/go-github/v29/github"
	"github.com/juju/errors"
	"github.com/you06/releaser/config"
	"github.com/you06/releaser/pkg/types"
	"github.com/you06/releaser/pkg/utils"
)

// Git ...
type Git struct {
	Github      *github.Client
	User        *github.User
	BaseDir     string
	Dir         string
	BaseRepo    types.Repo
	HeadRepo    types.Repo
	GithubToken string
}

// Config ...
type Config struct {
	Github *github.Client
	User   *github.User
	Base   types.Repo
	Head   types.Repo
	Dir    string
}

// New creates Git instance
func New(cfg *config.Config, gitCfg *Config) *Git {
	return &Git{
		Github:      gitCfg.Github,
		User:        gitCfg.User,
		BaseDir:     cfg.GitDir,
		Dir:         gitCfg.Dir,
		BaseRepo:    gitCfg.Base,
		HeadRepo:    gitCfg.Head,
		GithubToken: cfg.GithubToken,
	}
}

// Clone repo
func (g *Git) Clone() error {
	baseHTTPSaddr := g.BaseRepo.ComposeHTTPSWithCredential(g.User.GetLogin(), g.GithubToken)
	dir := path.Join(g.BaseDir, g.Dir)
	_, err := do(g.BaseDir, "git", "clone", baseHTTPSaddr, dir)
	return errors.Trace(err)
}

// Checkout a branch
func (g *Git) Checkout(branch string) error {
	dir := path.Join(g.BaseDir, g.Dir)
	_, err := do(dir, "git", "checkout", branch)
	return errors.Trace(err)
}

// CheckoutNew a new branch
func (g *Git) CheckoutNew(branch string) error {
	dir := path.Join(g.BaseDir, g.Dir)
	_, err := do(dir, "git", "checkout", "-b", branch)
	return errors.Trace(err)
}

// ReadFileContent read file by a relative path
func (g *Git) ReadFileContent(relative string) (string, error) {
	realpath := path.Join(g.BaseDir, g.Dir, relative)
	dat, err := ioutil.ReadFile(realpath)
	if err != nil {
		return "", errors.Trace(err)
	}
	return string(dat), nil
}

// WriteFileContent read file by a relative path
func (g *Git) WriteFileContent(relative, content string) error {
	realpath := path.Join(g.BaseDir, g.Dir, relative)
	return errors.Trace(ioutil.WriteFile(realpath, []byte(content), 0644))
}

// Commit do git add & git commit with sign up
func (g *Git) Commit(message string) error {
	dir := path.Join(g.BaseDir, g.Dir)
	_, _ = do(dir, "git", "add", "*")
	_, err := do(dir, "git", "commit", "-s", "-m", message)
	return errors.Trace(err)
}

// Push to head repo
func (g *Git) Push(branch string) error {
	var (
		dir      = path.Join(g.BaseDir, g.Dir)
		baseRepo = fmt.Sprintf("https://%s:%s@github.com/%s/%s.git",
			g.User.GetLogin(), g.GithubToken, g.HeadRepo.Owner, g.HeadRepo.Repo)
	)
	_, err := do(dir, "git", "push", baseRepo, branch, "--force")
	return errors.Trace(err)
}

// Clear delete cloned repo
func (g *Git) Clear() error {
	dir := path.Join(g.BaseDir, g.Dir)
	_, err := do(g.BaseDir, "rm", "-rf", dir)
	return errors.Trace(err)
}

// CreatePull creates pull request
func (g *Git) CreatePull(title, branch string) (*github.PullRequest, error) {
	newPull := github.NewPullRequest{
		Title:               github.String(title),
		Head:                github.String(fmt.Sprintf("%s:%s", g.HeadRepo.Owner, branch)),
		Base:                github.String("master"),
		Body:                github.String(title),
		MaintainerCanModify: github.Bool(true),
		Draft:               github.Bool(false),
	}
	ctx, _ := utils.NewTimeoutContext()
	pull, _, err := g.Github.PullRequests.Create(ctx,
		g.BaseRepo.Owner, g.BaseRepo.Repo, &newPull)
	return pull, errors.Trace(err)
}

func do(dir string, c string, args ...string) (string, error) {
	cmd := exec.Command(c, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}