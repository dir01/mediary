package tests

import (
	"os"
	"os/exec"
	"path"
	"testing"
)

const ytdlRevision = "4549522"

func GetYtdlDir(t *testing.T) string {
	temdDir := os.TempDir()
	ytdlDir := path.Join(temdDir, "ytdl-for-tests")

	if os.Getenv("UPDATE_YTDL") != "" {
		if out, err := exec.Command("rm", "-rf", ytdlDir).CombinedOutput(); err != nil {
			t.Fatalf("failed to remove youtube-dl repo (%v):\n%s", err, string(out))
		}
	}

	// if ytdlDir/.git exists then we assume it's a valid repo
	if _, err := os.Stat(path.Join(ytdlDir, ".git")); err == nil {
		return ytdlDir
	}
	// otherwise we clone the repo
	if err := exec.Command(
		"git", "clone", "https://github.com/ytdl-org/youtube-dl.git", ytdlDir,
	).Run(); err != nil {
		t.Fatalf("failed to clone youtube-dl repo: %v", err)
	}

	checkoutCmd := exec.Command("git", "checkout", ytdlRevision)
	checkoutCmd.Dir = ytdlDir
	if output, err := checkoutCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to checkout youtube-dl commit %s (%v): %s", ytdlRevision, err, string(output))
	}

	return ytdlDir
}
