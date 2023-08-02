package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cfabrica46/attrdet"
	"golang.org/x/net/html"
)

func main() {
	angularFlag := flag.Bool("angular", false, "Detect Angular attributes (ng-)")
	thymeleafFlag := flag.Bool("thymeleaf", false, "Detect Thymeleaf attributes (th:)")
	scriptFlag := flag.Bool("scripts", false, "Count <script> tags")

	flag.Parse()

	if flag.NFlag() != 1 || (!*angularFlag && !*thymeleafFlag && !*scriptFlag) {
		fmt.Println("Usage: go run main.go -angular|-thymeleaf|-scripts <folder>")

		return
	}

	folderPath := flag.Arg(0)

	var angularAttrs, thymeleafAttrs, angularVars, thymeleafVars map[string]int

	if *angularFlag {
		angularAttrs = make(map[string]int)
		angularVars = make(map[string]int)
	}

	if *thymeleafFlag {
		thymeleafAttrs = make(map[string]int)
		thymeleafVars = make(map[string]int)
	}

	processFolder(folderPath, angularFlag, thymeleafFlag, scriptFlag, angularAttrs, thymeleafAttrs, angularVars, thymeleafVars)
}

func processFolder(folderPath string, angularFlag, thymeleafFlag, scriptFlag *bool, angularAttrs, thymeleafAttrs, angularVars, thymeleafVars map[string]int) {
	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		fmt.Println("Error reading folder:", err)

		return
	}

	for _, file := range files {
		if file.IsDir() {
			processFolder(filepath.Join(folderPath, file.Name()), angularFlag, thymeleafFlag, scriptFlag, angularAttrs, thymeleafAttrs, angularVars, thymeleafVars)
		} else {
			processFile(filepath.Join(folderPath, file.Name()), angularFlag, thymeleafFlag, scriptFlag, angularAttrs, thymeleafAttrs, angularVars, thymeleafVars)
		}
	}
}

func processFile(filePath string, angularFlag, thymeleafFlag, scriptFlag *bool, angularAttrs, thymeleafAttrs, angularVars, thymeleafVars map[string]int) {
	// fmt.Printf("\n\n%s\n\n", filePath)

	// Check if the file is a HTML file
	if filepath.Ext(filePath) != ".html" && filepath.Ext(filePath) != ".mst" {
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)

		return
	}
	defer file.Close()

	doc, err := html.Parse(bufio.NewReader(file))
	if err != nil {
		fmt.Println("Error parsing HTML:", err)

		return
	}

	fileResult := fmt.Sprintf("File: %s\n", filePath)

	if *angularFlag {
		angularDetector := &attrdet.AngularDetector{BaseDetector: attrdet.BaseDetector{}}
		angularDetector.DetectAttributes(doc, angularAttrs)
		angularDetector.DetectVariables(doc, angularVars)

		fileResult += formatAttributes("Angular (ng) attributes:", angularAttrs)
		fileResult += formatVariables("Angular variables:", angularVars)
	}

	if *thymeleafFlag {
		thymeleafDetector := &attrdet.ThymeleafDetector{BaseDetector: attrdet.BaseDetector{}}
		thymeleafDetector.DetectAttributes(doc, thymeleafAttrs)
		thymeleafDetector.DetectVariables(doc, thymeleafVars)

		fileResult += formatAttributes("Thymeleaf attributes:", thymeleafAttrs)
		fileResult += formatVariables("Thymeleaf variables:", thymeleafVars)
	}

	if *scriptFlag {
		scriptDetector := &attrdet.BaseDetector{}
		scriptCount := scriptDetector.DetectScriptTags(doc)

		fileResult += fmt.Sprintf("Number of <script> tags: %d\n", scriptCount)
	}

	fileResult += "=====\n"

	fmt.Println(fileResult)
}

func formatAttributes(title string, attrs map[string]int) string {
	var result string
	if len(attrs) > 0 {
		result += fmt.Sprintf("%s\n", title)
		for attr, count := range attrs {
			result += fmt.Sprintf("  %s: %d occurrences\n", attr, count)
		}

		result += "\n"
	}

	return result
}

func formatVariables(title string, variables map[string]int) string {
	var result string
	if len(variables) > 0 {
		result += fmt.Sprintf("%s\n", title)
		for variable, count := range variables {
			result += fmt.Sprintf("  %s: %d occurrences\n", variable, count)
		}

		result += "\n"
	}

	return result
}
