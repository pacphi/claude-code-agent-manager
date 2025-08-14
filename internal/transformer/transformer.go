package transformer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/pacphi/claude-code-agent-manager/internal/config"
	"github.com/pacphi/claude-code-agent-manager/internal/util"
)

// Transformer handles file transformations
type Transformer struct {
	settings config.Settings
}

// New creates a new transformer
func New(settings config.Settings) *Transformer {
	return &Transformer{
		settings: settings,
	}
}

// Apply applies a transformation to files
func (t *Transformer) Apply(files []string, transform config.Transformation, sourcePath, targetPath string) ([]string, error) {
	switch transform.Type {
	case "remove_numeric_prefix":
		return t.removeNumericPrefix(files, transform, sourcePath, targetPath)
	case "extract_docs":
		return t.extractDocs(files, transform, sourcePath, targetPath)
	case "rename_files":
		return t.renameFiles(files, transform, sourcePath, targetPath)
	case "replace_content":
		return t.replaceContent(files, transform, sourcePath, targetPath)
	case "custom_script":
		return t.runCustomScript(files, transform, sourcePath, targetPath)
	default:
		return files, fmt.Errorf("unknown transformation type: %s", transform.Type)
	}
}

// removeNumericPrefix removes numeric prefixes from directory names
func (t *Transformer) removeNumericPrefix(files []string, transform config.Transformation, sourcePath, targetPath string) ([]string, error) {
	_ = targetPath // Not used in this transformation, kept for interface consistency
	pattern := transform.Pattern
	if pattern == "" {
		pattern = "^[0-9]{2}-"
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern: %w", err)
	}

	result := make([]string, 0, len(files))
	processedDirs := make(map[string]string)

	for _, file := range files {
		dir := filepath.Dir(file)
		base := filepath.Base(dir)

		// Check if directory name matches pattern
		if re.MatchString(base) {
			// Check if we've already processed this directory
			if newDir, exists := processedDirs[dir]; exists {
				// Use the already transformed directory
				result = append(result, filepath.Join(newDir, filepath.Base(file)))
			} else {
				// Transform directory name
				newBase := re.ReplaceAllString(base, "")
				newDir := filepath.Join(filepath.Dir(dir), newBase)
				processedDirs[dir] = newDir

				// Update file path
				result = append(result, filepath.Join(newDir, filepath.Base(file)))

				// Actually rename the directory if it exists
				oldPath := filepath.Join(sourcePath, dir)
				newPath := filepath.Join(sourcePath, newDir)

				if _, err := os.Stat(oldPath); err == nil {
					if err := os.Rename(oldPath, newPath); err != nil {
						return nil, fmt.Errorf("failed to rename directory %s to %s: %w", oldPath, newPath, err)
					}
				}
			}
		} else {
			result = append(result, file)
		}
	}

	return result, nil
}

// extractDocs extracts documentation files to a separate directory
func (t *Transformer) extractDocs(files []string, transform config.Transformation, sourcePath, targetPath string) ([]string, error) {
	_ = targetPath // Not used in this transformation, kept for interface consistency
	sourcePattern := transform.SourcePattern
	if sourcePattern == "" {
		sourcePattern = "*/README.md"
	}

	targetDir := transform.TargetDir
	if targetDir == "" {
		targetDir = t.settings.DocsDir
	}

	// Ensure docs directory exists
	docsPath := targetDir
	if !filepath.IsAbs(docsPath) {
		pwd, _ := os.Getwd()
		docsPath = filepath.Join(pwd, docsPath)
	}

	if err := os.MkdirAll(docsPath, 0750); err != nil {
		return nil, fmt.Errorf("failed to create docs directory: %w", err)
	}

	result := []string{}

	for _, file := range files {
		matched, _ := filepath.Match(sourcePattern, file)
		if matched {
			// Extract the documentation file
			dir := filepath.Dir(file)
			categoryName := filepath.Base(dir)

			// Transform naming
			docName := t.transformDocName(categoryName, transform.Naming)
			docPath := filepath.Join(docsPath, docName+".md")

			// Copy the file
			srcFile := filepath.Join(sourcePath, file)
			if err := t.copyFile(srcFile, docPath); err != nil {
				return nil, fmt.Errorf("failed to extract doc %s: %w", file, err)
			}

			// Note: extracted doc is written directly to project root
		} else {
			result = append(result, file)
		}
	}

	// Note: extracted docs are written directly to project root and should not
	// be included in the result to avoid duplication by the normal installer process

	return result, nil
}

// transformDocName transforms a category name according to naming strategy
func (t *Transformer) transformDocName(name, naming string) string {
	switch naming {
	case "UPPERCASE_UNDERSCORE":
		// Convert to uppercase with underscores
		name = strings.ReplaceAll(name, "-", "_")
		name = strings.ToUpper(name)
		return name

	case "lowercase_dash":
		// Already in this format typically
		return strings.ToLower(name)

	case "CamelCase":
		// Convert dash-separated to CamelCase
		parts := strings.Split(name, "-")
		for i, part := range parts {
			parts[i] = cases.Title(language.English).String(part)
		}
		return strings.Join(parts, "")

	default:
		// Default to uppercase underscore
		name = strings.ReplaceAll(name, "-", "_")
		name = strings.ToUpper(name)
		return name
	}
}

// renameFiles performs batch file renaming
func (t *Transformer) renameFiles(files []string, transform config.Transformation, sourcePath, targetPath string) ([]string, error) {
	// This would implement batch renaming logic
	// For now, return files unchanged
	return files, nil
}

// replaceContent performs content replacement in files
func (t *Transformer) replaceContent(files []string, transform config.Transformation, sourcePath, targetPath string) ([]string, error) {
	// This would implement content replacement logic
	// For now, return files unchanged
	return files, nil
}

// validateTransformArg validates transformation script arguments for security
func validateTransformArg(arg string) error {
	// Check for null bytes
	if strings.Contains(arg, "\x00") {
		return fmt.Errorf("null byte detected in argument: %s", arg)
	}

	// Check for command injection patterns
	injectionPatterns := []string{
		";", "|", "&", "$(", "`", "&&", "||", ">>", "<<",
	}

	for _, pattern := range injectionPatterns {
		if strings.Contains(arg, pattern) {
			return fmt.Errorf("potential command injection in argument: %s", arg)
		}
	}

	return nil
}

// runCustomScript runs a custom transformation script
func (t *Transformer) runCustomScript(files []string, transform config.Transformation, sourcePath, targetPath string) ([]string, error) {
	if transform.Script == "" {
		return nil, fmt.Errorf("script path is required for custom_script transformation")
	}

	// Validate script path for security
	if err := util.ValidatePath(transform.Script); err != nil {
		return nil, fmt.Errorf("invalid script path: %w", err)
	}

	// Validate paths for security
	if err := util.ValidatePath(sourcePath); err != nil {
		return nil, fmt.Errorf("invalid source path: %w", err)
	}
	if err := util.ValidatePath(targetPath); err != nil {
		return nil, fmt.Errorf("invalid target path: %w", err)
	}

	// Prepare arguments
	args := append([]string{}, transform.Args...)

	// Validate all arguments for security
	for i, arg := range args {
		if err := validateTransformArg(arg); err != nil {
			return nil, fmt.Errorf("invalid argument %d: %w", i, err)
		}
	}

	args = append(args, sourcePath, targetPath)

	// Run the script
	cmd := exec.Command(transform.Script, args...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("SOURCE_PATH=%s", sourcePath),
		fmt.Sprintf("TARGET_PATH=%s", targetPath),
		fmt.Sprintf("FILES_COUNT=%d", len(files)),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("custom script failed: %s\nOutput: %s", err, output)
	}

	// For now, return the original files
	// A more sophisticated implementation would parse script output
	return files, nil
}

// copyFile copies a file from src to dst
func (t *Transformer) copyFile(src, dst string) error {
	fm := util.NewFileManager()
	return fm.Copy(src, dst)
}
