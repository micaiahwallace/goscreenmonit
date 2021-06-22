package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/micaiahwallace/gowatchprog"
)

// Create the function definition and retreive the current executable directory path
func makeCurrentProgram(args []string) (*gowatchprog.Program, string) {

	// Get current exe file
	_, exePath, exeerr := currentExePath()
	if exeerr != nil {
		log.Fatalf("unable to get current exe path. %v\n", exeerr)
	}

	// Get parts to install
	exeFile := filepath.Base(exePath)

	// Create watchprog program definition
	prog := &gowatchprog.Program{
		Name:               PROG_NAME,
		ExeFile:            exeFile,
		Args:               args,
		InstallContext:     gowatchprog.AllUsers,
		StartupContext:     gowatchprog.AllUsers,
		WatchRetries:       -1,
		WatchRetryWait:     time.Second * 5,
		WatchRetryIncrease: 1,
	}

	return prog, exePath
}

// Create a file path with arguments appended
func createPathWithArgs(path string, args []string) string {
	return fmt.Sprintf(`"%s" %s`, path, strings.Join(args, " "))
}

// Get path to current executable and its directory
func currentExePath() (string, string, error) {
	exePath, pErr := os.Executable()
	if pErr != nil {
		return "", "", fmt.Errorf("exe path error: %v", pErr)
	}
	dirPath := filepath.Dir(exePath)

	return dirPath, exePath, nil
}

// Set logging output to filename in the same directory as the called executable
func logToFile(filename string) {

	var logFilePath string

	// Test if we are writing to a specific absolute file path
	if filepath.IsAbs(filename) {
		logFilePath = filename
	} else {

		// Get exe path dir
		dirPath, _, dperr := currentExePath()
		if dperr != nil {
			log.Println(dperr)
			return
		}

		// Get path to log file
		logFilePath = filepath.Join(dirPath, filename)
	}

	// Set log file path
	logfile, lferr := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if lferr != nil {
		log.Println(lferr)
		return
	}

	// Update global log path
	log.SetOutput(logfile)
}

// Set log file based on program data directory
func logToDataDir(p *gowatchprog.Program, fileName string) {
	dir, err := p.DataDirectory(true)
	if err != nil {
		log.Printf("unable to get or create data directory. %v", err)
	}
	logToFile(filepath.Join(dir, fileName))
}
