// Package proposer låter agenten föreslå kodändringar.
//
// Agenten läser repot, föreslår en ändring och lämnar in den för granskning.
// Den tillämpar inte själv ändringen - det gör en människa efter att ha
// godkänt förslaget.
package proposer

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"plan/agent/internal/core/messages"
	"plan/agent/internal/core/provider"
	"plan/agent/internal/executor"
	"plan/agent/internal/runusage"
	"plan/agent/internal/workspace"
)

// Proposal är agentens förslag på en kodändring.
type Proposal struct {
	// Title beskriver ändringen kort.
	Title string
	// Description förklarar varför ändringen behövs.
	Description string
	// Changes är de föreslagna ändringarna.
	Changes []executor.Change
	// TestCommand är kommandot som verifierar ändringen.
	TestCommand string
}

// Proposer tar fram förslag med hjälp av en språkmodell.
type Proposer struct {
	provider provider.LLMProvider
	model    provider.Model
	// usage samlar tokenanvändningen för körningen. Den kan vara nil -
	// då rapporteras ingen användning.
	usage *runusage.Recorder
}

// New skapar en proposer.
func New(provider provider.LLMProvider, model provider.Model) *Proposer {
	return &Proposer{provider: provider, model: model}
}

// WithUsage kopplar en recorder för tokenanvändning.
func (p *Proposer) WithUsage(recorder *runusage.Recorder) {
	p.usage = recorder
}

// Propose läser repot och föreslår en ändring för ett mål.
//
// Modellen får se filerna i repot och svara med ett strukturerat förslag.
// Om svaret inte går att tolka görs ett nytt försök - modellen kan
// svara fel format ibland.
func (p *Proposer) Propose(ctx context.Context, ws *workspace.Workspace, goalTitle, goalDescription string) (*Proposal, error) {
	files := collectFiles(ws, 40)

	var lastErr error
	for attempt := 0; attempt < maxProposeAttempts; attempt++ {
		proposal, err := p.proposeOnce(ctx, ws, goalTitle, goalDescription, files)
		if err == nil {
			return proposal, nil
		}

		lastErr = err
		log.Printf("Proposal attempt %d failed: %v", attempt+1, err)
	}

	return nil, lastErr
}

// maxProposeAttempts är hur många gånger modellen får svara innan
// förslaget lämnas över.
const maxProposeAttempts = 2

// proposeOnce gör ett försök att få ett förslag.
func (p *Proposer) proposeOnce(ctx context.Context, ws *workspace.Workspace, goalTitle, goalDescription string, files []string) (*Proposal, error) {
	msgs := []messages.Message{
		&messages.SystemMessage{
			Role:    messages.RoleSystem,
			Content: []messages.ContentBlock{&messages.TextBlock{Text: systemPrompt}},
		},
		&messages.UserMessage{
			Role:    messages.RoleUser,
			Content: []messages.ContentBlock{&messages.TextBlock{Text: buildUserPrompt(goalTitle, goalDescription, files)}},
		},
	}

	stream, err := p.provider.Stream(ctx, p.model, msgs, nil, provider.StreamingOptions{Usage: true})
	if err != nil {
		return nil, fmt.Errorf("failed to stream response: %w", err)
	}

	var response strings.Builder
	for event := range stream {
		switch e := event.(type) {
		case *provider.TextDeltaEvent:
			response.WriteString(e.Delta)
		case *provider.DoneEvent:
			p.recordUsage(e.Message.Usage)
			proposal, err := parseProposal(response.String())
			if err != nil {
				return nil, err
			}
			relativize(ws.Root(), proposal)
			return proposal, nil
		case *provider.ErrorEvent:
			return nil, e.Error
		}
	}

	return nil, fmt.Errorf("stream ended without completion")
}

// recordUsage sparar tokenanvändningen om en recorder är kopplad.
func (p *Proposer) recordUsage(usage *messages.Usage) {
	if p.usage == nil {
		return
	}
	p.usage.Record(usage)
}

// Revise rättar ett förslag vars egen verifiering inte gick igenom.
//
// Modellen får se felet från testkörningen och ska ge ett nytt förslag
// där både ändringen och testkommandot stämmer.
func (p *Proposer) Revise(ctx context.Context, ws *workspace.Workspace, goalTitle, goalDescription string, proposal *Proposal, failure string) (*Proposal, error) {
	msgs := []messages.Message{
		&messages.SystemMessage{
			Role:    messages.RoleSystem,
			Content: []messages.ContentBlock{&messages.TextBlock{Text: revisionSystemPrompt}},
		},
		&messages.UserMessage{
			Role:    messages.RoleUser,
			Content: []messages.ContentBlock{&messages.TextBlock{Text: buildRevisionPrompt(goalTitle, goalDescription, proposal, failure)}},
		},
	}

	stream, err := p.provider.Stream(ctx, p.model, msgs, nil, provider.StreamingOptions{Usage: true})
	if err != nil {
		return nil, fmt.Errorf("failed to stream response: %w", err)
	}

	var response strings.Builder
	for event := range stream {
		switch e := event.(type) {
		case *provider.TextDeltaEvent:
			response.WriteString(e.Delta)
		case *provider.DoneEvent:
			p.recordUsage(e.Message.Usage)
			revised, err := parseProposal(response.String())
			if err != nil {
				return nil, err
			}
			relativize(ws.Root(), revised)
			return revised, nil
		case *provider.ErrorEvent:
			return nil, e.Error
		}
	}

	return nil, fmt.Errorf("stream ended without completion")
}

// FixTest ber modellen skriva om testkommandot så att det går igenom på det
// föreslagna innehållet.
//
// Innehållet behålls - det är testet som är trasigt, inte ändringen.
func (p *Proposer) FixTest(ctx context.Context, ws *workspace.Workspace, goalTitle, goalDescription string, proposal *Proposal, failure string) (string, error) {
	msgs := []messages.Message{
		&messages.SystemMessage{
			Role:    messages.RoleSystem,
			Content: []messages.ContentBlock{&messages.TextBlock{Text: fixTestSystemPrompt}},
		},
		&messages.UserMessage{
			Role:    messages.RoleUser,
			Content: []messages.ContentBlock{&messages.TextBlock{Text: buildFixTestPrompt(goalTitle, goalDescription, proposal, failure)}},
		},
	}

	stream, err := p.provider.Stream(ctx, p.model, msgs, nil, provider.StreamingOptions{Usage: true})
	if err != nil {
		return "", fmt.Errorf("failed to stream response: %w", err)
	}

	var response strings.Builder
	for event := range stream {
		switch e := event.(type) {
		case *provider.TextDeltaEvent:
			response.WriteString(e.Delta)
		case *provider.DoneEvent:
			p.recordUsage(e.Message.Usage)
			return parseTestCommand(response.String()), nil
		case *provider.ErrorEvent:
			return "", e.Error
		}
	}

	return "", fmt.Errorf("stream ended without completion")
}

// parseTestCommand läser ett testkommando ur modellens svar.
func parseTestCommand(response string) string {
	for _, line := range strings.Split(response, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "TEST_COMMAND:") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "TEST_COMMAND:"))
		}
	}
	return ""
}

// buildFixTestPrompt beskriver innehållet och det felande testet.
func buildFixTestPrompt(goalTitle, goalDescription string, proposal *Proposal, failure string) string {
	var b strings.Builder

	b.WriteString("Goal: ")
	b.WriteString(goalTitle)
	b.WriteString("\n")
	b.WriteString(goalDescription)
	b.WriteString("\n\nThis test failed on your own proposed content:\n")
	b.WriteString(proposal.TestCommand)
	b.WriteString("\n\nThe failure was:\n")
	b.WriteString(failure)
	b.WriteString("\n\nThe content stays as it is:\n")

	for _, change := range proposal.Changes {
		b.WriteString("\nFILE: ")
		b.WriteString(change.Path)
		b.WriteString("\nCONTENT:\n")
		b.WriteString(change.Content)
		b.WriteString("\nEND_FILE\n")
	}

	b.WriteString("\nWrite a new test command that passes on exactly this content.")

	return b.String()
}

const fixTestSystemPrompt = `You are a software engineer fixing a test command.

The content of the change stays exactly as it is. Only the test command is
wrong - it checks for something the content does not contain, or it matches
text that is broken across lines.

Rules:
- The new test command must pass on the given content.
- Check every string the test looks for against the content, word by word.
- Prefer a test that checks structure - a heading, a function, a section -
  over one that checks exact sentences.
- The test must still verify the intent of the goal.
- Keep the test a single shell command that can be run from the repository
  root.

Answer with exactly one line:

TEST_COMMAND: [the shell command]
`

// buildRevisionPrompt beskriver det förslag som inte gick igenom.
func buildRevisionPrompt(goalTitle, goalDescription string, proposal *Proposal, failure string) string {
	var b strings.Builder

	b.WriteString("Goal: ")
	b.WriteString(goalTitle)
	b.WriteString("\n")
	b.WriteString(goalDescription)
	b.WriteString("\n\nYour previous proposal was:\n")
	b.WriteString("TITLE: ")
	b.WriteString(proposal.Title)
	b.WriteString("\nTEST_COMMAND: ")
	b.WriteString(proposal.TestCommand)
	b.WriteString("\n")

	for _, change := range proposal.Changes {
		b.WriteString("\nFILE: ")
		b.WriteString(change.Path)
		b.WriteString("\nCONTENT:\n")
		b.WriteString(change.Content)
		b.WriteString("\nEND_FILE\n")
	}

	b.WriteString("\nThe verification failed with:\n")
	b.WriteString(failure)
	b.WriteString("\n\nGive a corrected proposal. The test command must pass on the content you propose.")

	return b.String()
}

const revisionSystemPrompt = `You are a software engineer correcting a proposed change.

Your previous proposal was verified with its own test command and the test
failed. The most common reason is that the test looks for text that the
proposed content does not contain, or that the text is broken across lines.

Rules:
- Keep the goal and the intent of the previous proposal.
- Make the test command pass for the content you propose. Check every string
  the test looks for against the content, and be careful with line breaks.
- Prefer a test that checks structure over one that checks exact sentences.
- Keep the existing content of every file you change.

Answer in exactly this format:

TITLE: [one line]
DESCRIPTION: [why this change fulfils the goal]
TEST_COMMAND: [the shell command that verifies the change]

FILE: [path]
CONTENT:
[the full new content of the file]
END_FILE
`

// relativize gör om sökvägarna till relativa mot repots rot.
//
// Ett förslag ska vara giltigt oavsett var arbetsytan ligger, så sökvägarna
// får inte peka på en temporär katalog.
func relativize(root string, proposal *Proposal) {
	for i := range proposal.Changes {
		proposal.Changes[i].Path = relativePath(root, proposal.Changes[i].Path)
	}
}

// relativePath gör en sökväg relativ mot root om den ligger under den.
func relativePath(root, path string) string {
	if path == "" || !filepath.IsAbs(path) {
		return path
	}

	relative, err := filepath.Rel(root, path)
	if err != nil || strings.HasPrefix(relative, "..") {
		return path
	}
	return filepath.ToSlash(relative)
}

const systemPrompt = `You are a software engineer proposing a minimal code change.

You get a goal and the files of a repository. Propose the smallest change that
fulfils the goal, together with a test that covers it.

Rules:
- Change as little as possible. Do not refactor unrelated code.
- Always include a test that verifies the change.
- Only propose files that exist in the repository, or new files whose path you
  state explicitly.
- Never propose changes to secrets, CI configuration or dependency locks.
- Keep the existing content of every file you change. You propose the full
  new content of the file, so you must repeat everything that is already
  there - only add, change or remove what the goal actually requires.
- If you change a file, keep its existing structure, headings and tone.

Answer in exactly this format:

TITLE: [one line]
DESCRIPTION: [why this change fulfils the goal]
TEST_COMMAND: [the shell command that verifies the change]

FILE: [path]
CONTENT:
[the full new content of the file]
END_FILE
`

// buildUserPrompt bygger prompten med målet och filinnehållet.
func buildUserPrompt(goalTitle, goalDescription string, files []string) string {
	var sb strings.Builder

	sb.WriteString("## GOAL\n\n")
	sb.WriteString(goalTitle)
	sb.WriteString("\n\n")

	if goalDescription != "" {
		sb.WriteString("## DESCRIPTION\n\n")
		sb.WriteString(goalDescription)
		sb.WriteString("\n\n")
	}

	sb.WriteString("## REPOSITORY FILES\n\n")
	for _, file := range files {
		sb.WriteString(file)
		sb.WriteString("\n")
	}

	sb.WriteString("\n## YOUR TASK\n\n")
	sb.WriteString("Propose the change and the test command. Use the format from the instructions.")

	return sb.String()
}

// collectFiles läser de filer som är mest relevanta för ett förslag.
//
// Källkod och testfiler prioriteras - konfiguration och dokumentation
// sällan är vad ett kodmål handlar om.
func collectFiles(ws *workspace.Workspace, limit int) []string {
	changed, err := ws.ChangedFiles()
	if err == nil && len(changed) > 0 {
		return changed
	}

	files := []string{}
	walk(ws.Root(), ws.Root(), &files, limit)
	return files
}

// walk samlar filer i en katalog och prioriterar källkod.
func walk(root, dir string, files *[]string, limit int) {
	if len(*files) >= limit {
		return
	}

	entries, err := readDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if len(*files) >= limit {
			return
		}
		if strings.HasPrefix(entry.name, ".") {
			continue
		}
		if entry.isDir {
			if entry.name == "node_modules" || entry.name == "vendor" {
				continue
			}
			walk(root, entry.path, files, limit)
			continue
		}
		if !isRelevantSource(entry.name) {
			continue
		}

		content, err := readFile(entry.path)
		if err != nil {
			continue
		}
		*files = append(*files, fmt.Sprintf("### %s\n%s", relativePath(root, entry.path), content))
	}
}

// parseProposal läser modellens svar och gör om det till ett förslag.
func parseProposal(response string) (*Proposal, error) {
	proposal := &Proposal{}

	lines := strings.Split(response, "\n")
	var currentPath string
	var currentContent strings.Builder
	var testCommand strings.Builder
	inFile := false
	inTestCommand := false
	heredocEnd := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Ett testkommando kan bestå av flera rader - exempelvis ett
		// heredoc-skript. Det läses tills heredoc-markören kommer.
		if inTestCommand {
			testCommand.WriteString(line)
			testCommand.WriteString("\n")
			if trimmed == heredocEnd {
				inTestCommand = false
				heredocEnd = ""
				proposal.TestCommand = proposal.TestCommand + "\n" + strings.TrimRight(testCommand.String(), "\n")
				testCommand.Reset()
			}
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "TITLE:"):
			proposal.Title = strings.TrimSpace(strings.TrimPrefix(trimmed, "TITLE:"))
			inFile = false
		case strings.HasPrefix(trimmed, "DESCRIPTION:"):
			proposal.Description = strings.TrimSpace(strings.TrimPrefix(trimmed, "DESCRIPTION:"))
			inFile = false
		case strings.HasPrefix(trimmed, "TEST_COMMAND:"):
			proposal.TestCommand = strings.TrimSpace(strings.TrimPrefix(trimmed, "TEST_COMMAND:"))
			inFile = false
			if end, ok := heredocTerminator(proposal.TestCommand); ok {
				inTestCommand = true
				heredocEnd = end
			}
		case strings.HasPrefix(trimmed, "FILE:"):
			if inFile && currentPath != "" {
				proposal.Changes = append(proposal.Changes, executor.Change{
					Path:    currentPath,
					Content: strings.TrimRight(currentContent.String(), "\n"),
				})
			}
			currentPath = strings.TrimSpace(strings.TrimPrefix(trimmed, "FILE:"))
			currentContent.Reset()
			inFile = true
		case trimmed == "END_FILE":
			if inFile && currentPath != "" {
				proposal.Changes = append(proposal.Changes, executor.Change{
					Path:    currentPath,
					Content: strings.TrimRight(currentContent.String(), "\n"),
				})
			}
			currentPath = ""
			currentContent.Reset()
			inFile = false
		case inFile && currentPath != "":
			// "CONTENT:" är en markör, inte en del av filen.
			if trimmed == "CONTENT:" {
				continue
			}
			currentContent.WriteString(line)
			currentContent.WriteString("\n")
		}
	}

	if inFile && currentPath != "" {
		proposal.Changes = append(proposal.Changes, executor.Change{
			Path:    currentPath,
			Content: strings.TrimRight(currentContent.String(), "\n"),
		})
	}

	if inTestCommand {
		// heredoc som aldrig avslutades - behåll det vi läste.
		proposal.TestCommand = strings.TrimRight(testCommand.String(), "\n")
	}

	if len(proposal.Changes) == 0 {
		return nil, fmt.Errorf("failed to parse any file change from the response")
	}
	if proposal.Title == "" {
		proposal.Title = "Föreslagen ändring"
	}

	return proposal, nil
}

// heredocTerminator läser avslutningen av ett heredoc.
//
// Ett testkommando som börjar med "python3 - <<'PY'" fortsätter tills raden
// "PY" kommer. Returnerar false om kommandot inte är ett heredoc.
func heredocTerminator(command string) (string, bool) {
	index := strings.Index(command, "<<")
	if index < 0 {
		return "", false
	}

	rest := strings.TrimSpace(command[index+2:])
	rest = strings.Trim(rest, `'"`)
	if rest == "" {
		return "", false
	}

	return strings.Fields(rest)[0], true
}

// isRelevantSource berättar om en fil är relevant för ett kodmål.
//
// Dokumentation och konfiguration räknas också - ett mål kan gälla en
// README eller en deployment, och då måste agenten se den filen för att
// kunna ändra den utan att förstöra resten.
func isRelevantSource(name string) bool {
	switch {
	case strings.HasSuffix(name, ".go"):
		return true
	case strings.HasSuffix(name, ".ts"), strings.HasSuffix(name, ".vue"):
		return true
	case strings.HasSuffix(name, ".md"):
		return true
	case strings.HasSuffix(name, ".yaml"), strings.HasSuffix(name, ".yml"):
		return true
	case strings.HasSuffix(name, ".sh"):
		return true
	case strings.HasSuffix(name, "_test.go"):
		return true
	default:
		return false
	}
}

// dirEntry är en post i en katalog.
type dirEntry struct {
	name  string
	path  string
	isDir bool
}

// readDir läser en katalog.
func readDir(path string) ([]dirEntry, error) {
	items, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	entries := make([]dirEntry, 0, len(items))
	for _, item := range items {
		entries = append(entries, dirEntry{
			name:  item.Name(),
			path:  filepath.Join(path, item.Name()),
			isDir: item.IsDir(),
		})
	}
	return entries, nil
}

// readFile läser en fil.
func readFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
