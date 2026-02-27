package services

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

/* XAMPP Services */
type XAMPPServiceStatus struct {
	Apache bool
	MySQL  bool
	FTP    bool
}

type ServiceInfo struct {
	Name  string
	PID   string
	Port  string
	State bool
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

func GetXAMPPServiceDetails() (map[string]ServiceInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	infos := make(map[string]ServiceInfo)

	// Apache (buscar httpd y apache2)
	apachePID := ""
	portApache := "N/A"
	// Buscar PID para httpd
	cmdApache := exec.CommandContext(ctx, "pidof", "-s", "httpd")
	if out, err := cmdApache.Output(); err == nil && strings.TrimSpace(string(out)) != "" {
		apachePID = strings.TrimSpace(string(out))
	} else {
		// Buscar PID para apache2
		cmdApache2 := exec.CommandContext(ctx, "pidof", "-s", "apache2")
		if out2, err2 := cmdApache2.Output(); err2 == nil && strings.TrimSpace(string(out2)) != "" {
			apachePID = strings.TrimSpace(string(out2))
		}
	}
	if apachePID != "" {
		cmdPort := exec.CommandContext(ctx, "ss", "-lptn")
		if pout, perr := cmdPort.Output(); perr == nil {
			lines := strings.Split(string(pout), "\n")
			found := false
			for _, line := range lines {
				if strings.Contains(line, apachePID) && (strings.Contains(line, "httpd") || strings.Contains(line, "apache2")) {
					re := regexp.MustCompile(`:(\d+)`)
					if pmatch := re.FindStringSubmatch(line); len(pmatch) == 2 {
						portApache = pmatch[1]
						found = true
						break
					}
				}
			}
			// Si no se encontró por PID, buscar por nombre de proceso
			if !found {
				for _, line := range lines {
					if strings.Contains(line, "httpd") || strings.Contains(line, "apache2") {
						re := regexp.MustCompile(`:(\d+)`)
						if pmatch := re.FindStringSubmatch(line); len(pmatch) == 2 {
							portApache = pmatch[1]
							break
						}
					}
				}
			}
		}
	}

	// MySQL
	mysqlPID := ""
	portMySQL := "N/A"
	cmdMySQL := exec.CommandContext(ctx, "pidof", "-s", "mysqld")
	if out, err := cmdMySQL.Output(); err == nil && strings.TrimSpace(string(out)) != "" {
		mysqlPID = strings.TrimSpace(string(out))
		if mysqlPID != "" {
			cmdPort := exec.CommandContext(ctx, "ss", "-lptn")
			if pout, perr := cmdPort.Output(); perr == nil {
				lines := strings.Split(string(pout), "\n")
				for _, line := range lines {
					if strings.Contains(line, mysqlPID) && strings.Contains(line, "mysqld") {
						re := regexp.MustCompile(`:(\d+)`)
						if pmatch := re.FindStringSubmatch(line); len(pmatch) == 2 {
							portMySQL = pmatch[1]
							break
						}
					}
				}
			}
		}
	}

	// FTP (ProFTPD)
	ftpPID := ""
	portFTP := "N/A"
	cmdFTP := exec.CommandContext(ctx, "pidof", "-s", "proftpd")
	if out, err := cmdFTP.Output(); err == nil && strings.TrimSpace(string(out)) != "" {
		ftpPID = strings.TrimSpace(string(out))
		if ftpPID != "" {
			cmdPort := exec.CommandContext(ctx, "ss", "-lptn")
			if pout, perr := cmdPort.Output(); perr == nil {
				lines := strings.Split(string(pout), "\n")
				for _, line := range lines {
					if strings.Contains(line, ftpPID) && strings.Contains(line, "proftpd") {
						re := regexp.MustCompile(`:(\d+)`)
						if pmatch := re.FindStringSubmatch(line); len(pmatch) == 2 {
							portFTP = pmatch[1]
							break
						}
					}
				}
			}
		}
	}

	// Estado desde status
	cmdStatus := exec.CommandContext(ctx, "sudo", "/opt/lampp/lampp", "status")
	outStatus, err := cmdStatus.Output()
	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("timeout al consultar estado de XAMPP")
	}
	if err != nil {
		return nil, err
	}
	status := string(outStatus)

	infos["Apache"] = ServiceInfo{
		Name:  "Apache",
		PID:   apachePID,
		Port:  portApache,
		State: strings.Contains(status, "Apache is running"),
	}
	infos["MySQL"] = ServiceInfo{
		Name:  "MySQL",
		PID:   mysqlPID,
		Port:  portMySQL,
		State: strings.Contains(status, "MySQL is running"),
	}
	infos["FTP"] = ServiceInfo{
		Name:  "FTP",
		PID:   ftpPID,
		Port:  portFTP,
		State: strings.Contains(status, "ProFTPD is running"),
	}

	return infos, nil
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
