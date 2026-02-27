package services

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sudo", "/opt/lampp/lampp", "status")
	out, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		return XAMPPServiceStatus{}, fmt.Errorf("timeout al consultar estado de XAMPP")
	}
	if err != nil {
		return XAMPPServiceStatus{}, err
	}
	status := string(out)
	return XAMPPServiceStatus{
		Apache: strings.Contains(status, "Apache is running"),
		MySQL:  strings.Contains(status, "MySQL is running"),
		FTP:    strings.Contains(status, "ProFTPD is running"),
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
