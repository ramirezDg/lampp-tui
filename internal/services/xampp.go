package services

import (
	"fmt"
	"os/exec"
)

/* XAMPP Services */
type XAMPPServiceStatus struct {
	Apache bool
	MySQL  bool
	FTP    bool
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (contains(s[1:], substr) || contains(s[:len(s)-1], substr)))) || (len(s) >= len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}

func GetXAMPPServiceStatus() (XAMPPServiceStatus, error) {
	cmd := exec.Command("sudo", "/opt/lampp/lampp", "status")
	out, err := cmd.Output()
	if err != nil {
		return XAMPPServiceStatus{}, fmt.Errorf("error al obtener estado de XAMPP: %v", err)
	}
	status := string(out)
	return XAMPPServiceStatus{
		Apache: contains(status, "Apache is running"),
		MySQL:  contains(status, "MySQL is running"),
		FTP:    contains(status, "ProFTPD is running"),
	}, nil
}

func ControlXAMPPService(service, action string) error {
	var cmd *exec.Cmd
	switch service {
	case "apache":
		cmd = exec.Command("/opt/lampp/lampp", action+"apache")
	case "mysql":
		cmd = exec.Command("/opt/lampp/lampp", action+"mysql")
	case "ftp":
		cmd = exec.Command("/opt/lampp/lampp", action+"ftp")
	case "all":
		cmd = exec.Command("/opt/lampp/lampp", action)
	default:
		return fmt.Errorf("servicio no soportado: %s", service)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error al ejecutar acción %s en %s: %v", action, service, err)
	}
	return nil
}
