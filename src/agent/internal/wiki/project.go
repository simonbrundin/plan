package wiki

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ProjectTagPrefix markerar vilket projekt en anteckning gäller.
const ProjectTagPrefix = "projekt:"

// DetectProject returnerar namnet på projektet (repot) som katalogen tillhör.
// Det görs genom att fråga git om repo-roten; faller det tillbaka på
// katalognamnet.
func DetectProject(dir string) string {
	if dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return ""
		}
		dir = cwd
	}

	if name := gitRepoName(dir); name != "" {
		return name
	}
	return filepath.Base(dir)
}

// gitRepoName frågar git vilket repo katalogen tillhör.
func gitRepoName(dir string) string {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return filepath.Base(strings.TrimSpace(string(out)))
}

// ProjectTag bygger taggen som markerar ett projekt.
func ProjectTag(project string) string {
	if project == "" {
		return ""
	}
	return ProjectTagPrefix + project
}

// ProjectOf returnerar projektnamnet en anteckning gäller, om något.
func (n *Note) ProjectOf() string {
	for _, tag := range n.Tags {
		if strings.HasPrefix(tag, ProjectTagPrefix) {
			return strings.TrimPrefix(tag, ProjectTagPrefix)
		}
	}
	return ""
}

// HasProject svarar på om anteckningen gäller ett visst projekt.
func (n *Note) HasProject(project string) bool {
	if project == "" {
		return true
	}
	return strings.EqualFold(n.ProjectOf(), project)
}
