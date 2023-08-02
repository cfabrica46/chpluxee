package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cfabrica46/attrdet"
	"github.com/spf13/cobra"
	"golang.org/x/net/html"
)

var (
	angularFlag   bool
	thymeleafFlag bool
	scriptFlag    bool

	rootCmd = &cobra.Command{
		Use:   "main",
		Short: "Attribute detector",
		Long:  `This program detects certain attributes in HTML files.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("missing folder argument")
			}

			if (angularFlag && thymeleafFlag) || (!angularFlag && !thymeleafFlag && !scriptFlag) {
				return fmt.Errorf("please provide at least one flag: --angular, --thymeleaf or --scripts")
			}

			folderPath := args[0]

			angularAttrs := make(map[string]int)
			thymeleafAttrs := make(map[string]int)
			angularVars := make(map[string]int)
			thymeleafVars := make(map[string]int)

			return processFolder(folderPath, angularAttrs, thymeleafAttrs, angularVars, thymeleafVars)
		},
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&angularFlag, "angular", "a", false, "Detect Angular attributes (ng-)")
	rootCmd.PersistentFlags().BoolVarP(&thymeleafFlag, "thymeleaf", "t", false, "Detect Thymeleaf attributes (th:)")
	rootCmd.PersistentFlags().BoolVarP(&scriptFlag, "scripts", "s", false, "Count <script> tags")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func processFolder(folderPath string, angularAttrs, thymeleafAttrs, angularVars, thymeleafVars map[string]int) error {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return fmt.Errorf("failed to read folder: %w", err)
	}

	for _, file := range files {
		path := filepath.Join(folderPath, file.Name())

		if file.IsDir() {
			if err = processFolder(path, angularAttrs, thymeleafAttrs, angularVars, thymeleafVars); err != nil {
				return err
			}
		} else {
			if err = processFile(path, angularAttrs, thymeleafAttrs, angularVars, thymeleafVars); err != nil {
				return err
			}
		}
	}

	return nil
}

func processFile(filePath string, angularAttrs, thymeleafAttrs, angularVars, thymeleafVars map[string]int) error {
	switch filepath.Ext(filePath) {
	case ".html", ".mst":
		// continue
	default:
		return nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	doc, err := html.Parse(bufio.NewReader(file))
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %w", err)
	}

	var results []string
	results = append(results, fmt.Sprintf("File: %s\n", filePath))

	if angularFlag {
		angularDetector := &attrdet.AngularDetector{BaseDetector: attrdet.BaseDetector{}}
		angularDetector.DetectAttributes(doc, angularAttrs)
		angularDetector.DetectVariables(doc, angularVars)

		results = append(results, formatAttributes("Angular (ng) attributes:", angularAttrs))
		results = append(results, formatVariables("Angular variables:", angularVars))
	}

	if thymeleafFlag {
		thymeleafDetector := &attrdet.ThymeleafDetector{BaseDetector: attrdet.BaseDetector{}}
		thymeleafDetector.DetectAttributes(doc, thymeleafAttrs)
		thymeleafDetector.DetectVariables(doc, thymeleafVars)

		results = append(results, formatAttributes("Thymeleaf attributes:", thymeleafAttrs))
		results = append(results, formatVariables("Thymeleaf variables:", thymeleafVars))
	}

	if scriptFlag {
		scriptDetector := attrdet.BaseDetector{}
		scriptCount := scriptDetector.DetectScriptTags(doc)

		results = append(results, fmt.Sprintf("Number of <script> tags: %d\n", scriptCount))
	}

	results = append(results, "=====\n")
	fmt.Println(strings.Join(results, ""))

	return nil
}

func formatAttributes(title string, attrs map[string]int) string {
	if len(attrs) == 0 {
		return ""
	}

	var result string
	result += fmt.Sprintf("%s\n", title)

	for attr, count := range attrs {
		result += fmt.Sprintf("  %s: %d occurrences\n", attr, count)
	}

	result += "\n"

	return result
}

func formatVariables(title string, variables map[string]int) string {
	if len(variables) == 0 {
		return ""
	}

	var result string
	result += fmt.Sprintf("%s\n", title)

	for variable, count := range variables {
		result += fmt.Sprintf("  %s: %d occurrences\n", variable, count)
	}

	result += "\n"

	return result
}

