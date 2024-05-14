// main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/anacrolix/torrent/metainfo"
)

const version = "1.1.0"

var (
	verboseFlag = flag.Bool("verbose", false, "Enable verbose output")
	versionFlag = flag.Bool("version", false, "Show version information")
	helpFlag    = flag.Bool("help", false, "Show help message")
)

func main() {
	flag.Parse()

	switch {
	case *helpFlag:
		printUsage()
		return
	case *versionFlag:
		fmt.Println("rmBitTrackers version:", version)
		return
	}

	args := flag.Args()
	if len(args) < 1 {
		logAndExit("Error: Torrent file path is required", 1)
	}

	filePath, err := filepath.Abs(args[0])
	if err != nil {
		logAndExit(fmt.Sprintf("Error resolving file path: %v", err), 1)
	}

	if err := validateInputFile(filePath); err != nil {
		logAndExit(fmt.Sprintf("Error validating file: %v", err), 1)
	}

	outputDir := getOutputDir(args)

	metaInfo, err := loadMetaInfo(filePath)
	if err != nil {
		logAndExit(fmt.Sprintf("Error loading meta info: %v", err), 1)
	}

	modifyMetadata(metaInfo, "unethicalteam", "trackers removed with https://github.com/unethicalteam/rmBitTrackers")

	savedFilePath, err := saveModifiedFile(metaInfo, filePath, outputDir)
	if err != nil {
		logAndExit(fmt.Sprintf("Error saving modified torrent: %v", err), 1)
	}

	processMetaInfo(metaInfo)

	fmt.Printf("Modified torrent saved as: %s\n", savedFilePath)
}

func loadMetaInfo(filePath string) (*metainfo.MetaInfo, error) {
	logVerbose("Loading torrent:", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening torrent: %v", err)
	}
	defer file.Close()

	metaInfo, err := metainfo.Load(file)
	if err != nil {
		return nil, fmt.Errorf("error decoding torrent: %v", err)
	}

	return metaInfo, nil
}

func saveModifiedFile(metaInfo *metainfo.MetaInfo, originalFilePath, outputPath string) (string, error) {
	_, fileName := filepath.Split(originalFilePath)
	newFilePath := getNewFilePath(outputPath, fileName)

	if err := os.MkdirAll(filepath.Dir(newFilePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	newFile, err := os.Create(newFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer newFile.Close()

	if err := metaInfo.Write(newFile); err != nil {
		return "", fmt.Errorf("failed to write to file: %v", err)
	}

	return newFilePath, nil
}

func processMetaInfo(metaInfo *metainfo.MetaInfo) {
	var wg sync.WaitGroup

	wg.Add(3)

	go func() {
		defer wg.Done()
		fileName, err := extractNameFromMetaInfo(metaInfo)
		if err != nil {
			log.Printf("Error extracting name from metadata: %v\n", err)
			return
		}
		fmt.Println("File name:", fileName)
	}()

	go func() {
		defer wg.Done()
		totalSize, err := getTotalSize(metaInfo)
		if err != nil {
			log.Printf("Error calculating total size: %v\n", err)
			return
		}
		fmt.Println("Total size:", totalSize)
	}()

	go func() {
		defer wg.Done()
		infoHashString := getInfoHash(metaInfo)
		fmt.Println("Info hash:", infoHashString)

		magnetLink := generateMagnetLink(metaInfo, infoHashString, 0) // Placeholder for fileName and totalSize
		fmt.Println("Magnet link:", magnetLink)
	}()

	wg.Wait()
}

func getNewFilePath(outputPath, fileName string) string {
	if strings.HasSuffix(outputPath, string(os.PathSeparator)) || !strings.Contains(filepath.Ext(outputPath), ".") {
		return filepath.Join(outputPath, fileName)
	}
	return outputPath
}

func validateInputFile(filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}
	if fileInfo.IsDir() {
		return fmt.Errorf("expected a file but got a directory: %s", filePath)
	}
	return nil
}

func getOutputDir(args []string) string {
	if len(args) > 1 {
		outputDir, err := filepath.Abs(args[1])
		if err != nil {
			logAndExit(fmt.Sprintf("Error resolving output directory path: %v", err), 1)
		}
		return outputDir
	}
	return "."
}

func printUsage() {
	exeName := "rmBitTrackers"
	if runtime.GOOS == "windows" {
		exeName += ".exe"
	}

	fmt.Printf("Usage: %s [options] <torrent-file> [output-path]\n", exeName)
	fmt.Println("Options:")
	fmt.Println("  --verbose          Enable verbose output")
	fmt.Println("  --version          Show version information")
	fmt.Println("  --help             Show this help message")
	fmt.Println("\nExamples:")
	fmt.Printf("  %s --verbose example.torrent\n", exeName)
	fmt.Printf("  %s example.torrent ./modified/example.torrent\n", exeName)
	fmt.Printf("  %s --help\n", exeName)
}

func logVerbose(v ...interface{}) {
	if *verboseFlag {
		fmt.Println(v...)
	}
}

func logAndExit(message string, code int) {
	log.Println(message)
	os.Exit(code)
}
