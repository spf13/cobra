// Copyright 2013-2023 The Cobra Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package doc

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// SkillsConfig holds configuration for generating an agentskills.io-compatible
// SKILL.md file.
type SkillsConfig struct {
	// Name is the skill name. Must be 1-64 lowercase alphanumeric characters
	// and hyphens. Must not start or end with a hyphen or contain consecutive
	// hyphens. If empty, it is derived from the command name.
	Name string

	// Description describes what the skill does and when to use it.
	// Max 1024 characters. If empty, it is derived from the command's
	// Short and Long descriptions.
	Description string

	// License is an optional license name or reference.
	License string

	// Compatibility optionally indicates environment requirements.
	// Max 500 characters.
	Compatibility string

	// Metadata holds arbitrary key-value pairs for additional metadata.
	Metadata map[string]string

	// AllowedTools is an optional space-delimited list of pre-approved tools.
	AllowedTools string

	// DisableModelInvocation prevents agents from automatically loading
	// this skill. Set to true for workflows triggered manually.
	DisableModelInvocation bool

	// Notes are global notes rendered in SKILL.md body as a "Notes" section.
	// Each entry becomes a bullet point. Useful for cross-cutting information
	// that applies to multiple commands (e.g., "Most list commands support -o json").
	Notes []string
}

// toSkillName converts a command name to a valid skill name.
func toSkillName(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	if len(s) > 64 {
		s = s[:64]
		s = strings.TrimRight(s, "-")
	}
	return s
}

// GenSkills generates an agentskills.io-compatible SKILL.md document
// for the command tree. The output is concise and suitable for use as
// the main SKILL.md. For large command trees, use GenSkillsDir which
// also generates a references/REFERENCE.md with detailed documentation.
func GenSkills(cmd *cobra.Command, w io.Writer, config SkillsConfig) error {
	return genSkillsInternal(cmd, w, config, false)
}

func genSkillsInternal(cmd *cobra.Command, w io.Writer, config SkillsConfig, hasReference bool) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)

	name := config.Name
	if name == "" {
		name = toSkillName(cmd.Name())
	}

	description := config.Description
	if description == "" {
		description = cmd.Short
		if len(cmd.Long) > 0 {
			description = cmd.Long
		}
	}
	if len(description) > 1024 {
		description = description[:1024]
	}

	genFrontmatter(buf, name, description, config)
	genSkillsBody(buf, cmd, hasReference, config)

	_, err := buf.WriteTo(w)
	return err
}

// GenSkillsDir generates a skill directory following the agentskills.io
// progressive disclosure convention:
//
//	<dir>/<skill-name>/SKILL.md
//	<dir>/<skill-name>/references/<command_path>.md  (one per command)
//
// SKILL.md contains the frontmatter and a concise command overview.
// Each reference file contains detailed documentation for a single
// command including usage, examples, and flags. Agents load only the
// reference file they need, keeping context usage minimal.
func GenSkillsDir(cmd *cobra.Command, dir string, config SkillsConfig) error {
	name := config.Name
	if name == "" {
		name = toSkillName(cmd.Name())
	}

	skillDir := filepath.Join(dir, name)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		return err
	}

	skillFile, err := os.Create(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		return err
	}
	defer skillFile.Close()

	if err := genSkillsInternal(cmd, skillFile, config, true); err != nil {
		return err
	}

	refDir := filepath.Join(skillDir, "references")
	if err := os.MkdirAll(refDir, 0o755); err != nil {
		return err
	}

	commands := collectCommands(cmd)
	for _, c := range commands {
		basename := cmdRefFilename(c)
		f, err := os.Create(filepath.Join(refDir, basename))
		if err != nil {
			return err
		}
		err = genRefFile(c, f)
		f.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// genFrontmatter writes the YAML frontmatter block.
func genFrontmatter(buf *bytes.Buffer, name, description string, config SkillsConfig) {
	buf.WriteString("---\n")
	fmt.Fprintf(buf, "name: %s\n", name)
	fmt.Fprintf(buf, "description: %s\n", yamlEscapeString(description))
	if config.License != "" {
		fmt.Fprintf(buf, "license: %s\n", config.License)
	}
	if config.Compatibility != "" {
		fmt.Fprintf(buf, "compatibility: %s\n", yamlEscapeString(config.Compatibility))
	}
	if config.DisableModelInvocation {
		buf.WriteString("disable-model-invocation: true\n")
	}
	if config.AllowedTools != "" {
		fmt.Fprintf(buf, "allowed-tools: %s\n", config.AllowedTools)
	}
	if len(config.Metadata) > 0 {
		buf.WriteString("metadata:\n")
		keys := make([]string, 0, len(config.Metadata))
		for k := range config.Metadata {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(buf, "  %s: %s\n", k, yamlEscapeString(config.Metadata[k]))
		}
	}
	buf.WriteString("---\n\n")
}

// genSkillsBody writes the concise SKILL.md body with a command summary.
func genSkillsBody(buf *bytes.Buffer, cmd *cobra.Command, hasReference bool, config SkillsConfig) {
	commands := collectCommands(cmd)

	buf.WriteString("# " + cmd.Name() + "\n\n")
	if len(cmd.Long) > 0 {
		buf.WriteString(cmd.Long + "\n\n")
	} else {
		buf.WriteString(cmd.Short + "\n\n")
	}

	if len(config.Notes) > 0 {
		buf.WriteString("## Notes\n\n")
		for _, note := range config.Notes {
			fmt.Fprintf(buf, "- %s\n", note)
		}
		buf.WriteString("\n")
	}

	if cmd.Runnable() {
		fmt.Fprintf(buf, "```\n%s\n```\n\n", cmd.UseLine())
	}

	if len(commands) > 1 {
		buf.WriteString("## Available Commands\n\n")
		for _, c := range commands[1:] {
			if hasReference {
				fmt.Fprintf(buf, "- [`%s`](references/%s) - %s\n", c.CommandPath(), cmdRefFilename(c), c.Short)
			} else {
				fmt.Fprintf(buf, "- `%s` - %s\n", c.CommandPath(), c.Short)
			}
		}
		buf.WriteString("\n")
	}

	if hasReference {
		fmt.Fprintf(buf, "See [references/%s](references/%s) for root command flags.\n\n", cmdRefFilename(cmd), cmdRefFilename(cmd))
	}

	buf.WriteString("Run `" + cmd.Name() + " --help` or `" + cmd.Name() + " <command> --help` for full usage details.\n")
}

// cmdRefFilename returns the reference filename for a command,
// e.g. "root_echo_times.md".
func cmdRefFilename(cmd *cobra.Command) string {
	return strings.ReplaceAll(cmd.CommandPath(), " ", "_") + markdownExtension
}

// genRefFile writes a detailed reference file for a single command.
func genRefFile(cmd *cobra.Command, w io.Writer) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)
	name := cmd.CommandPath()

	buf.WriteString("# " + name + "\n\n")
	buf.WriteString(cmd.Short + "\n\n")

	if len(cmd.Long) > 0 && cmd.Long != cmd.Short {
		buf.WriteString(cmd.Long + "\n\n")
	}

	if cmd.Runnable() {
		fmt.Fprintf(buf, "```\n%s\n```\n\n", cmd.UseLine())
	}

	if len(cmd.Example) > 0 {
		buf.WriteString("## Examples\n\n")
		fmt.Fprintf(buf, "```\n%s\n```\n\n", cmd.Example)
	}

	if err := printOptions(buf, cmd, name); err != nil {
		return err
	}

	_, err := buf.WriteTo(w)
	return err
}

// collectCommands returns cmd and all available descendant commands
// in depth-first order.
func collectCommands(cmd *cobra.Command) []*cobra.Command {
	var result []*cobra.Command
	result = append(result, cmd)

	children := cmd.Commands()
	sort.Sort(byName(children))

	for _, child := range children {
		if !child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
			continue
		}
		result = append(result, collectCommands(child)...)
	}
	return result
}

// yamlEscapeString wraps a string in quotes if it contains special
// YAML characters.
func yamlEscapeString(s string) string {
	if strings.ContainsAny(s, ":#{}[]|>&*!%@`,\n\"") {
		escaped := strings.ReplaceAll(s, `"`, `\"`)
		return `"` + escaped + `"`
	}
	return s
}
