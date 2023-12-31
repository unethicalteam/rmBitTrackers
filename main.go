// main.go
package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"
    "github.com/anacrolix/torrent/metainfo"
    "runtime"
)

const version = "1.1.0"

var (
    verboseFlag = flag.Bool("verbose", false, "Enable verbose output")
    versionFlag = flag.Bool("version", false, "Show version information")
    helpFlag    = flag.Bool("help", false, "Show help message")
)

func main() {
    flag.Parse()

    if *helpFlag {
        printUsage()
        return
    }

    if *versionFlag {
        fmt.Println("rmBitTrackers version:", version)
        return
    }

    args := flag.Args()
    if len(args) < 1 {
        log.Println("error: torrent file path is required")
        printUsage()
        os.Exit(1)
    }

    filePath, err := filepath.Abs(args[0])
    if err != nil {
        log.Printf("error resolving file path: %v\n", err)
        return
    }

    if err := validateInputFile(filePath); err != nil {
        log.Println(err)
        return
    }

    outputDir := "." // default output directory
    if len(args) > 1 {
        outputDir, err = filepath.Abs(args[1])
        if err != nil {
            log.Printf("error resolving output directory path: %v\n", err)
            return
        }
    }

    file, err := openTorrentFile(filePath)
    if err != nil {
        log.Printf("error opening torrent: %v\n", err)
        return
    }
    defer file.Close()

    metaInfo, err := decodeTorrentFile(file)
    if err != nil {
        log.Printf("error decoding torrent: %v\n", err)
        return
    }

    modifyMetadata(metaInfo, "unethicalteam", "trackers removed with https://github.com/unethicalteam/rmBitTrackers")

    savedFilePath, err := saveModifiedFile(metaInfo, filePath, outputDir)
    if err != nil {
        log.Printf("error saving modified torrent: %v\n", err)
        return
    }

    fileName, err := extractNameFromMetaInfo(metaInfo)
    if err != nil {
        log.Printf("error extracting name from metadata: %v\n", err)
        return
    }

    totalSize, err := getTotalSize(metaInfo)
    if err != nil {
        log.Printf("error calculating total size: %v\n", err)
        return
    }

    infoHashString := getInfoHash(metaInfo)
    fmt.Println("hash:", infoHashString)

    magnetLink := generateMagnetLink(metaInfo, fileName, totalSize)
    fmt.Println("magnet link:", magnetLink)

    fmt.Println("modified torrent saved as:", savedFilePath)
}

func openTorrentFile(filePath string) (*os.File, error) {
    if *verboseFlag {
        fmt.Println("opening torrent:", filePath)
    }
    return os.Open(filePath)
}

func decodeTorrentFile(file *os.File) (*metainfo.MetaInfo, error) {
    if *verboseFlag {
        fmt.Println("decoding torrent...")
    }
    return metainfo.Load(file)
}

func saveModifiedFile(metaInfo *metainfo.MetaInfo, originalFilePath, outputPath string) (string, error) {
    _, fileName := filepath.Split(originalFilePath)
    var newFilePath string

    if strings.HasSuffix(outputPath, string(os.PathSeparator)) {
        err := os.MkdirAll(outputPath, 0755)
        if err != nil {
            return "", fmt.Errorf("failed to create directory: %v", err)
        }
        newFilePath = filepath.Join(outputPath, fileName)
    } else {
        ext := filepath.Ext(outputPath)
        fileInfo, err := os.Stat(outputPath)
        if os.IsNotExist(err) && ext == "" {
            err = os.MkdirAll(outputPath, 0755)
            if err != nil {
                return "", fmt.Errorf("failed to create directory: %v", err)
            }
            newFilePath = filepath.Join(outputPath, fileName)
        } else if err == nil && fileInfo.IsDir() {
            newFilePath = filepath.Join(outputPath, fileName)
        } else {
            newFilePath = outputPath
        }
    }

    newFile, err := os.Create(newFilePath)
    if err != nil {
        return "", fmt.Errorf("failed to create file: %v", err)
    }
    defer newFile.Close()

    err = metaInfo.Write(newFile)
    if err != nil {
        return "", fmt.Errorf("failed to write to file: %v", err)
    }

    return newFilePath, nil
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