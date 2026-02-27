package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func InstalarXAMPP(version string) error {
	carpetaDestino := "./downloads"
	if _, err := os.Stat(carpetaDestino); os.IsNotExist(err) {
		if err := os.MkdirAll(carpetaDestino, 0755); err != nil {
			LogToFile(fmt.Sprintf("Error creating destination folder: %v", err))
			return fmt.Errorf("error creating destination folder: %v", err)
		}
		LogToFile(fmt.Sprintf("Created destination folder: %s", carpetaDestino))
	}
	logMsg := fmt.Sprintf("Installing XAMPP version: %s", version)
	fmt.Println(logMsg)
	LogToFile(logMsg)

	downloadURL := "https://sourceforge.net/projects/xampp/files/XAMPP%20Linux/" + version + "/xampp-linux-x64-" + version + "-0-installer.run/download"
	logMsg = fmt.Sprintf("Downloading from: %s", downloadURL)
	fmt.Println(logMsg)
	LogToFile(logMsg)

	logMsg = fmt.Sprintf("The download will be saved in the folder: %s", carpetaDestino)
	fmt.Println(logMsg)
	LogToFile(logMsg)

	cmdDescarga := exec.Command("wget", "--content-disposition", "-P", carpetaDestino, downloadURL)
	if err := cmdDescarga.Run(); err != nil {
		LogToFile(fmt.Sprintf("Error downloading XAMPP: %v", err))
		return fmt.Errorf("error downloading XAMPP: %v", err)
	}
	LogToFile("Download command executed successfully.")

	files, err := os.ReadDir(carpetaDestino)
	if err != nil {
		LogToFile(fmt.Sprintf("Error reading destination folder: %v", err))
		return fmt.Errorf("error reading destination folder: %v", err)
	}
	found := false
	for _, file := range files {
		if !file.IsDir() && len(file.Name()) > 4 && file.Name()[len(file.Name())-4:] == ".run" {
			logMsg = fmt.Sprintf("Downloaded file: %s", file.Name())
			fmt.Println(logMsg)
			LogToFile(logMsg)
			found = true
		}
	}
	if !found {
		logMsg = "Warning: No .run file found in the downloads folder after download."
		fmt.Println(logMsg)
		LogToFile(logMsg)
	} else {
		logMsg = "Download completed."
		fmt.Println(logMsg)
		LogToFile(logMsg)
	}
	return nil
}


func LogToFile(msg string) {
	logDir := "../logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		os.Mkdir(logDir, 0755)
	}
	logPath := filepath.Join(logDir, "app.log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer logFile.Close()
	logFile.WriteString(msg + "\n")
}