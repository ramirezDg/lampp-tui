package services

import (
	"fmt"
	"os"
	"os/exec"
)

func InstalarXAMPP(version string, url string) error {
	carpetaDestino := "./downloads"
	if _, err := os.Stat(carpetaDestino); os.IsNotExist(err) {
		if err := os.MkdirAll(carpetaDestino, 0755); err != nil {
			return fmt.Errorf("error creating destination folder: %v", err)
		}
	}
	fmt.Printf("Installing XAMPP version: %s\n", version)
	downloadURL := "https://sourceforge.net/projects/xampp/files/XAMPP%20Linux/" + version + "/xampp-linux-x64-" + version + "-installer.run/download"
	fmt.Printf("Downloading from: %s\n", downloadURL)
	fmt.Printf("The download will be saved in the folder: %s\n", carpetaDestino)
	rutaArchivo := carpetaDestino + "/xampp-linux-x64-" + version + "-installer.run"
	cmdDescarga := exec.Command("wget", downloadURL, "-O", rutaArchivo)
	if err := cmdDescarga.Run(); err != nil {
		return fmt.Errorf("error downloading XAMPP: %v", err)
	}
	cmdPermisos := exec.Command("chmod", "+x", rutaArchivo)
	if err := cmdPermisos.Run(); err != nil {
		return fmt.Errorf("error setting permissions: %v", err)
	}
	var confirm string
	fmt.Print("Do you want to run the installer now? (y/n): ")
	fmt.Scanln(&confirm)
	if confirm == "y" || confirm == "Y" {
		cmdInstalar := exec.Command("sudo", rutaArchivo)
		if err := cmdInstalar.Run(); err != nil {
			return fmt.Errorf("error installing XAMPP: %v", err)
		}
	} else {
		fmt.Println("Installer downloaded but not executed.")
	}
	return nil
}
