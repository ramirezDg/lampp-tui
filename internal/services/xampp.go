package services

import (
	"fmt"
	"os/exec"
)

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
