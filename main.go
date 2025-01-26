package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/81hvac1k02/FanslyStreamDownloader/fansly"
)

// loadEnv reads a file and sets the environment variables
func loadEnv(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Split the line into key and value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Skip malformed lines
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		// Set the environment variable
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func main() {
	logger := log.New(os.Stdout, "FanslyDownloader: ", log.LstdFlags)
	userData := fansly.UserData{}
	var envFile string
	// Define command line flags
	flag.StringVar(&userData.FanslyCreator, "name", "", "Creator's username to download stream from")
	flag.StringVar(&userData.FanslyToken, "token", "", "Fansly authentication token")
	flag.StringVar(&userData.UserAgent, "agent", "", "User agent string")
	flag.StringVar(&userData.BasePath, "path", "", "Base path for downloads")
	flag.StringVar(&envFile, "env", ".env", "Path to .env file")
	flag.BoolVar(&userData.Metadata, "metadata", false, "Download metadata y/N")
	flag.Parse()
	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// Handle username from positional argument
	if userData.FanslyCreator == "" && len(flag.Args()) > 0 {
		userData.FanslyCreator = flag.Args()[0]
	}

	if userData.FanslyCreator == "" {
		logger.Fatal("No username provided. Use --name flag or provide as first argument")
	}

	// Load environment variables if flags not provided
	if userData.FanslyToken == "" || userData.UserAgent == "" {
		err := loadEnv(envFile)
		if err != nil {
			logger.Fatal("Could not read env:", err)
		}

		// Load from env if not provided by flags
		if userData.FanslyToken == "" {
			userData.FanslyToken = os.Getenv("FANSLY_TOKEN")
		}
		if userData.UserAgent == "" {
			userData.UserAgent = os.Getenv("USER_AGENT")
		}
		if userData.BasePath == "" {
			userData.BasePath = os.Getenv("BASEPATH")
		}
	}
	// Set default base path if not provided
	if userData.BasePath == "" {
		logger.Println("No basepath provided. Using current working directory")
		var err error
		userData.BasePath, err = os.Getwd()
		if err != nil {
			logger.Fatal("Could not get current working directory:", err)
		}
	}

	// Download the stream

	for errCount := 0; errCount <= 6; errCount++ {

		if fansly.DownloadStream(&userData, true) != nil {
			logger.Printf("Error downloading stream. Retrying in %v seconds\n", errCount+1)
			time.Sleep(time.Duration(errCount) + 1*time.Second) //exponential backoff

			continue
		}

	}

}
