package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var ignored = flag.Bool("ignored", false, "show ignored files")

type Walker struct {
	Root string

	gitRepo   map[string]bool
	skeleton  map[string]bool
	untracked map[string]bool
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Println("missing required argument (root path)")
		return
	}
	walker := &Walker{
		Root:      flag.Arg(0),
		gitRepo:   make(map[string]bool),
		skeleton:  make(map[string]bool),
		untracked: make(map[string]bool),
	}
	walker.findGitRepos()
	walker.findUntracked()
	walker.printStatus()
}

type LocalStatus int

const (
	LocalClean LocalStatus = iota
	LocalDirty
	LocalUntracked
)

func (ls LocalStatus) icon() rune {
	switch ls {
	case LocalClean:
		return ' '
	case LocalDirty:
		return '?'
	case LocalUntracked:
		return '!'
	}
	return 'X'
}

type Status struct {
	LocalStatus LocalStatus

	WorkingTreeStatus []byte
}

func (w *Walker) printStatus() {
	statuses := make(map[string]Status)
	for repo := range w.gitRepo {
		out, err := gitStatus(repo)
		if err != nil {
			log.Fatalf("git status error: %s", err)
		}

		st := Status{
			WorkingTreeStatus: out,
		}

		if len(out) == 0 {
			st.LocalStatus = LocalClean
		} else {
			st.LocalStatus = LocalDirty
		}

		statuses[repo] = st
	}

	for path := range w.untracked {
		statuses[path] = Status{
			LocalStatus: LocalUntracked,
		}
	}

	paths := make([]string, len(statuses))
	i := 0
	for path := range statuses {
		paths[i] = path
		i++
	}
	sort.Strings(paths)

	for _, path := range paths {
		st := statuses[path]
		if st.LocalStatus == LocalClean {
			continue
		}

		relpath, _ := filepath.Rel(w.Root, path)
		fmt.Printf("%c %s\n", st.LocalStatus.icon(), relpath)
		if len(st.WorkingTreeStatus) > 0 {
			s := string(st.WorkingTreeStatus)
			lines := strings.Split(s, "\n")
			for i, line := range lines {
				if line == "" {
					continue
				}
				fmt.Printf("  %s\n", line)
				if i >= 10 {
					fmt.Printf("  (%d more lines)\n", len(lines)-i)
					break
				}
			}
			fmt.Println("")
		}
	}
}

func (w *Walker) findGitRepos() error {
	return filepath.Walk(w.Root, w.walkGit)
}

// findUntracked should be called after findGitRepos.
func (w *Walker) findUntracked() error {
	return filepath.Walk(w.Root, w.walkUntracked)
}

func (w *Walker) walkGit(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() && info.Name() == ".git" {
		repo := filepath.Dir(path) // drop .git suffix
		w.gitRepo[repo] = true

		// The repo's ancestors are skeleton directories: they provide structure
		// that organizes git repos, and should not be considered untracked.
		ancestors := repoAncestors(w.Root, repo)
		for _, dir := range ancestors {
			w.skeleton[dir] = true
		}

		// Skip the .git directory.
		return filepath.SkipDir
	}
	return nil
}

func repoAncestors(root string, repoDir string) []string {
	ancestors := make([]string, 0, 3)
	curr := repoDir
	for {
		dir := filepath.Dir(curr)
		ancestors = append(ancestors, dir)
		if dir == "." || dir == root || dir == curr {
			break
		}
		curr = dir
	}
	return ancestors
}

// walkUntracked finds files and directories that are not part of a git repository.
func (w *Walker) walkUntracked(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if w.gitRepo[path] {
		return filepath.SkipDir
	}
	if w.skeleton[path] {
		// Skeleton directories contain a git repo, so we should not mark
		// the whole directory untracked. Instead, we recurse into the directory
		// to find untracked paths.
		return nil
	}

	w.untracked[path] = true

	if info.IsDir() {
		return filepath.SkipDir
	} else {
		return nil
	}
}

func gitStatus(repo string) ([]byte, error) {
	args := []string{"-C", repo, "status", "--short", "--porcelain"}
	if *ignored {
		args = append(args, "--ignored")
	}
	cmd := exec.Command("git", args...)
	return cmd.Output()
}
