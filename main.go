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
	angularFlag    bool
	thymeleafFlag  bool
	scriptFlag     bool
	fileOutputFlag bool
	outputDir      string

	totalAngularVars   int // Nuevo contador total para variables Angular
	totalThymeleafVars int // Contador total para variables Thymeleaf
	totalScriptTags    int // Contador total para etiquetas de scripts
	totalHtmlFiles     int // Nuevo contador total para archivos HTML

	// Nuevo contador total para archivos HTML sin etiquetas de script y con ellas
	totalHtmlFilesWithoutScriptTags int
	totalHtmlFilesWithScriptTags    int

	// Nuevo contador total para archivos HTML sin variables y con ellas
	totalHtmlFilesWithoutVars int
	totalHtmlFilesWithVars    int

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

			// Create the output directory if it doesn't exist
			err := os.Mkdir(outputDir, os.ModePerm)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}

			err = processFolder(folderPath, angularAttrs, thymeleafAttrs, angularVars, thymeleafVars)
			if err != nil {
				return err
			}

			// Al final del análisis, imprimir y guardar los contadores totales
			err = writeTotalCountToFile()
			if err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&angularFlag, "angular", "a", false, "Detect Angular attributes (ng-)")
	rootCmd.PersistentFlags().BoolVarP(&thymeleafFlag, "thymeleaf", "t", false, "Detect Thymeleaf attributes (th:)")
	rootCmd.PersistentFlags().BoolVarP(&scriptFlag, "scripts", "s", false, "Count <script> tags")
	rootCmd.PersistentFlags().BoolVarP(&fileOutputFlag, "fileoutput", "f", false, "Outputs the results to a file with the name of the folder and .txt extension")
	rootCmd.PersistentFlags().StringVarP(&outputDir, "outputdir", "o", "output", "Specify the directory to store output files")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	fmt.Println("🚀 Éxito increíble! Has logrado correr el script sin problemas. ¡Sigue con el gran trabajo!")
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
		totalHtmlFiles++
	default:
		return nil
	}

	fileHasScriptTags := false
	fileHasVars := false
	scriptCount := 0

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

		// Añadir las variables de Angular al contador total
		for _, count := range angularVars {
			totalAngularVars += count
		}

		// Si el archivo tiene variables Angular, marcar fileHasVars como true
		if len(angularVars) > 0 || len(angularAttrs) > 0 {
			fileHasVars = true
		}
	}

	if thymeleafFlag {
		thymeleafDetector := &attrdet.ThymeleafDetector{BaseDetector: attrdet.BaseDetector{}}
		thymeleafDetector.DetectAttributes(doc, thymeleafAttrs)
		thymeleafDetector.DetectVariables(doc, thymeleafVars)

		results = append(results, formatAttributes("Thymeleaf attributes:", thymeleafAttrs))
		results = append(results, formatVariables("Thymeleaf variables:", thymeleafVars))

		// Añadir las variables de Thymeleaf al contador total
		for _, count := range thymeleafVars {
			totalThymeleafVars += count
		}

		// Si el archivo tiene variables Thymeleaf, marcar fileHasVars como true
		if len(thymeleafVars) > 0 || len(thymeleafAttrs) > 0 {
			fileHasVars = true
		}
	}

	if scriptFlag {
		scriptDetector := attrdet.BaseDetector{}
		scriptCount = scriptDetector.DetectScriptTags(doc)

		// Añadir las etiquetas de script al contador total
		totalScriptTags += scriptCount

		results = append(results, fmt.Sprintf("Number of <script> tags: %d\n", scriptCount))

		// Si el archivo tiene etiquetas de script, marcar fileHasScriptTags como true
		if scriptCount > 0 {
			fileHasScriptTags = true
		}
	}

	// Incrementar contadores de archivos sin etiquetas de script y con ellas
	if fileHasScriptTags {
		totalHtmlFilesWithScriptTags++
	} else {
		totalHtmlFilesWithoutScriptTags++
	}

	// Incrementar contadores de archivos sin variables y con ellas
	if fileHasVars {
		totalHtmlFilesWithVars++
	} else {
		totalHtmlFilesWithoutVars++
	}

	results = append(results, "=====\n")

	// No crear archivo si no hay ninguna variable, atributo o script
	if len(angularAttrs) == 0 && len(thymeleafAttrs) == 0 && len(angularVars) == 0 && len(thymeleafVars) == 0 && scriptCount == 0 {
		return nil
	}

	if fileOutputFlag {
		err = writeResultsToFile(filePath, results)
		if err != nil {
			return err
		}
	} else {
		fmt.Println(strings.Join(results, ""))
	}

	return nil
}

func writeResultsToFile(filePath string, results []string) error {
	outputFilePath := filepath.Join(outputDir, filepath.Base(filePath)+".txt")
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	_, err = outputFile.WriteString(strings.Join(results, ""))
	if err != nil {
		return err
	}

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

func writeTotalCountToFile() error {
	outputFilePath := filepath.Join(outputDir, "Total_Counts.txt")
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	if angularFlag {
		_, err = outputFile.WriteString(fmt.Sprintf("Total Angular Variables: %d\n", totalAngularVars))
		if err != nil {
			return err
		}
	}

	if thymeleafFlag {
		_, err = outputFile.WriteString(fmt.Sprintf("Total Thymeleaf Variables: %d\n", totalThymeleafVars))
		if err != nil {
			return err
		}
	}

	if scriptFlag {
		_, err = outputFile.WriteString(fmt.Sprintf("Total Script Tags: %d\n", totalScriptTags))
		if err != nil {
			return err
		}
	}

	// Escribir el contador total de archivos HTML en el archivo
	_, err = outputFile.WriteString(fmt.Sprintf("Total HTML Files: %d\n", totalHtmlFiles))
	if err != nil {
		return err
	}

	// Escribir contadores de archivos sin etiquetas de script y con ellas en el archivo
	_, err = outputFile.WriteString(fmt.Sprintf("Total HTML Files Without Script Tags: %d\n", totalHtmlFilesWithoutScriptTags))
	if err != nil {
		return err
	}
	_, err = outputFile.WriteString(fmt.Sprintf("Total HTML Files With Script Tags: %d\n", totalHtmlFilesWithScriptTags))
	if err != nil {
		return err
	}

	// Escribir contadores de archivos sin variables y con ellas en el archivo
	_, err = outputFile.WriteString(fmt.Sprintf("Total HTML Files Without Variables: %d\n", totalHtmlFilesWithoutVars))
	if err != nil {
		return err
	}
	_, err = outputFile.WriteString(fmt.Sprintf("Total HTML Files With Variables: %d\n", totalHtmlFilesWithVars))
	if err != nil {
		return err
	}

	return nil
}
