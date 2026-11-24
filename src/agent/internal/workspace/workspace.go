// Package workspace isolerar agentens kodändringar från originalet.
//
// Agenten får aldrig skriva direkt i det repo som utvecklas. Istället
// arbetar den i en kopia, och ändringen diffas tillbaka så att en människa
// kan granska den innan den appliceras.
package workspace

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ErrNotPublished betyder att ändringen inte kom ut till källrepoet.
//
// Ändringen är då gjord och verifierad, men ligger bara lokalt.
var ErrNotPublished = errors.New("change was not published")

// Workspace är en isolerad kopia av ett repository.
type Workspace struct {
	// root är sökvägen till kopian.
	root string
	// source är repot som kopierades.
	source string
}

// New skapar en isolerad kopia av source i en temporär katalog.
//
// Kopian görs med git så att historik, .gitignore och binärer följer med,
// men inga förändringar sker i originalet.
func New(source string) (*Workspace, error) {
	absSource, err := filepath.Abs(source)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve source path: %w", err)
	}

	info, err := os.Stat(absSource)
	if err != nil {
		return nil, fmt.Errorf("failed to read source directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("source %s is not a directory", absSource)
	}

	root, err := os.MkdirTemp("", "agent-workspace-")
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace directory: %w", err)
	}

	ws := &Workspace{root: root, source: absSource}
	if err := ws.copySource(); err != nil {
		os.RemoveAll(root)
		return nil, err
	}

	return ws, nil
}

// copySource kopierar repot till arbetsytan.
//
// git används när källan är ett git-repository - annars kopieras filerna
// rakt av. git är bättre eftersom det respekterar .gitignore och behåller
// filrättigheter.
func (w *Workspace) copySource() error {
	if isGitRepository(w.source) {
		cmd := exec.Command("git", "clone", "--local", "--no-hardlinks", w.source, w.root)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to clone repository: %w: %s", err, strings.TrimSpace(string(output)))
		}
		if err := w.setIdentity(); err != nil {
			return err
		}
		return w.restoreRemote()
	}

	if err := copyTree(w.source, w.root); err != nil {
		return fmt.Errorf("failed to copy repository: %w", err)
	}
	return nil
}

// setIdentity sätter git-identiteten för arbetsytan.
//
// Agentens commits måste kunna attribueras. Identiteten sätts lokalt i
// arbetsytan - den globala git-konfigurationen i poden lämnas orörd.
func (w *Workspace) setIdentity() error {
	if _, err := w.Run("git", "config", "user.email", "agent@plan.local"); err != nil {
		return fmt.Errorf("failed to set git email: %w", err)
	}
	if _, err := w.Run("git", "config", "user.name", "Plan Agent"); err != nil {
		return fmt.Errorf("failed to set git name: %w", err)
	}
	return nil
}

// restoreRemote sätter tillbaka källans remote-adress.
//
// Arbetsytan klonas lokalt från källkatalogen, så annars pekar origin
// mot en lokal sökväg och pushen når aldrig det riktiga repot.
func (w *Workspace) restoreRemote() error {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = w.source
	output, err := cmd.Output()
	if err != nil {
		// Källan kan sakna remote - då behålls den lokala adressen.
		return nil
	}

	remote := strings.TrimSpace(string(output))
	if remote == "" || remote == w.source {
		return nil
	}

	if _, err := w.Run("git", "remote", "set-url", "origin", remote); err != nil {
		return fmt.Errorf("failed to set remote origin: %w", err)
	}
	return nil
}

// isGitRepository berättar om källan är ett git-repository.
func isGitRepository(path string) bool {
	if _, err := exec.LookPath("git"); err != nil {
		return false
	}

	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = path
	return cmd.Run() == nil
}

// Root returnerar sökvägen till arbetsytan.
func (w *Workspace) Root() string {
	return w.root
}

// Source returnerar sökvägen till repot som kopierades.
func (w *Workspace) Source() string {
	return w.source
}

// Write skriver en fil i arbetsytan och skapar katalager vid behov.
func (w *Workspace) Write(relativePath, content string) error {
	target, err := w.resolve(relativePath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	if err := os.WriteFile(target, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

// Read läser en fil från arbetsytan.
func (w *Workspace) Read(relativePath string) (string, error) {
	target, err := w.resolve(relativePath)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(target)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(content), nil
}

// Exists berättar om en fil finns i arbetsytan.
func (w *Workspace) Exists(relativePath string) bool {
	target, err := w.resolve(relativePath)
	if err != nil {
		return false
	}
	_, err = os.Stat(target)
	return err == nil
}

// resolve gör om en relativ sökväg till en sökväg i arbetsytan.
//
// Sökvägar som försöker lämna arbetsytan avvisas - agenten ska inte kunna
// skriva till originalet eller till systemkataloger.
func (w *Workspace) resolve(relativePath string) (string, error) {
	if relativePath == "" {
		return "", fmt.Errorf("empty path")
	}
	if filepath.IsAbs(relativePath) {
		return "", fmt.Errorf("absolute paths are not allowed: %s", relativePath)
	}

	target := filepath.Join(w.root, filepath.Clean(relativePath))
	if !strings.HasPrefix(target, w.root+string(os.PathSeparator)) && target != w.root {
		return "", fmt.Errorf("path escapes the workspace: %s", relativePath)
	}

	return target, nil
}

// Diff returnerar skillnaden mellan originalet och arbetsytan.
//
// Utan git-repository används git diff --no-index, som jämför två kataloger
// utan att kräva att någon av dem är ett repository.
func (w *Workspace) Diff() (string, error) {
	if isGitRepository(w.root) {
		cmd := exec.Command("git", "diff", "--no-color")
		cmd.Dir = w.root
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("failed to produce diff: %w: %s", err, strings.TrimSpace(string(output)))
		}
		return string(output), nil
	}

	if _, err := exec.LookPath("git"); err == nil {
		cmd := exec.Command("git", "diff", "--no-color", "--no-index", w.source, w.root)
		output, _ := cmd.CombinedOutput()
		return string(output), nil
	}

	return diffTrees(w.source, w.root)
}

// ChangedFiles returnerar de filer som ändrats i arbetsytan.
func (w *Workspace) ChangedFiles() ([]string, error) {
	if isGitRepository(w.root) {
		cmd := exec.Command("git", "status", "--porcelain")
		cmd.Dir = w.root
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("failed to list changed files: %w: %s", err, strings.TrimSpace(string(output)))
		}
		return parseGitStatus(string(output)), nil
	}

	return changedFiles(w.source, w.root)
}

// Run utför ett kommando i arbetsytan och returnerar utdata.
//
// Kommandot körs med arbetsytan som arbetskatalog, utan skalning - agenten
// ska kunna köra projektets egna tester.
func (w *Workspace) Run(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = w.root

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("command %q failed: %w: %s", command, err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

// RunShell utför ett kommando med skalning i arbetsytan.
//
// Används för kommandon som behöver skalning, som "go test ./... && go vet ./...".
// Kommandot körs i arbetsytan, inte i originalet.
func (w *Workspace) RunShell(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = w.root

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("command %q failed: %w: %s", command, err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

// CommitToBranch committer ändringarna i arbetsytan och skickar branchen
// till källrepoet.
//
// Branchen är granskbar: en människa kan jämföra den med huvudgrenen och
// merge:a den när den är rätt. Huvudgrenen rörs aldrig.
//
// Om pushen misslyckas - till exempel för att miljön saknar rättigheter att
// skriva till repot - returneras ErrNotPublished. Ändringen finns då lokalt
// och i diffen, men inte i källrepoet.
func (w *Workspace) CommitToBranch(branch, message string) error {
	if !isGitRepository(w.root) {
		return fmt.Errorf("workspace is not a git repository")
	}

	if _, err := w.Run("git", "checkout", "-b", branch); err != nil {
		return fmt.Errorf("failed to create branch %s: %w", branch, err)
	}
	if _, err := w.Run("git", "add", "-A"); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}
	if _, err := w.Run("git", "commit", "-m", message); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}
	if err := w.pushBranch(branch); err != nil {
		return fmt.Errorf("%w: %v", ErrNotPublished, err)
	}

	return nil
}

// pushBranch skickar branchen till källrepoet.
//
// Autentisering sker med GIT_AUTH_HEADER om den finns - en färdig
// Authorization-header. Ingen hemlighet skrivs då i .git/config.
func (w *Workspace) pushBranch(branch string) error {
	// git -c måste komma före subkommandot, annars tolkas det som ett
	// alternativ till push.
	args := []string{}
	if header := os.Getenv("GIT_AUTH_HEADER"); header != "" {
		args = append(args, "-c", "http.extraheader="+header)
	}
	args = append(args, "push", "origin", branch)

	if _, err := w.Run("git", args...); err != nil {
		return err
	}
	return nil
}

// Cleanup tar bort arbetsytan.
func (w *Workspace) Cleanup() error {
	if w.root == "" {
		return nil
	}
	return os.RemoveAll(w.root)
}

// copyTree kopierar en katalog rekursivt.
func copyTree(source, target string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}

		destination := filepath.Join(target, relative)
		if info.IsDir() {
			return os.MkdirAll(destination, info.Mode())
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
			return err
		}
		return os.WriteFile(destination, content, info.Mode())
	})
}

// parseGitStatus tolkar git status --porcelain och ger filnamnen.
//
// Formatet är "XY <sökväg>" - de två första tecknen är status. Radens
// inledande blanksteg hör till statusen och ska inte tas bort före slicing,
// då blir sökvägen förkortad.
func parseGitStatus(output string) []string {
	files := []string{}
	for _, line := range strings.Split(output, "\n") {
		if len(line) < 4 {
			continue
		}
		files = append(files, strings.TrimSpace(line[3:]))
	}
	return files
}

// diffTrees jämför två kataloger och returnerar en enkel diff.
func diffTrees(source, target string) (string, error) {
	changed, err := changedFiles(source, target)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, file := range changed {
		sb.WriteString("--- a/" + file + "\n")
		sb.WriteString("+++ b/" + file + "\n")

		original, _ := os.ReadFile(filepath.Join(source, file))
		updated, _ := os.ReadFile(filepath.Join(target, file))
		sb.WriteString(fmt.Sprintf("@@ original %d bytes, updated %d bytes @@\n", len(original), len(updated)))
	}

	return sb.String(), nil
}

// changedFiles jämför filerna i två kataloger.
func changedFiles(source, target string) ([]string, error) {
	changed := []string{}

	err := filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relative, err := filepath.Rel(target, path)
		if err != nil {
			return err
		}

		original, err := os.ReadFile(filepath.Join(source, relative))
		if err != nil {
			changed = append(changed, relative)
			return nil
		}

		updated, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if string(original) != string(updated) {
			changed = append(changed, relative)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return changed, nil
}
