package services

import (
	"fmt"
	"os/exec"
)

type SFResponse struct {
	Children []struct {
		Name string `json:"name"`
	} `json:"children"`
}

func ObtenerVersiones() ([]string, error) {
	bashScript := `
		curl -s https://sourceforge.net/projects/xampp/files/XAMPP%20Linux/ | \
		gawk '
			BEGIN { ver=""; count=""; link=""; }
			/<tr title=/ { ver=""; count=""; link=""; }
			/<a href="\/projects\/xampp\/files\/XAMPP%20Linux\/[0-9.]+\// {
				match($0, /<a href="(\/projects\/xampp\/files\/XAMPP%20Linux\/[0-9.]+\/)"/, arr)
				if (arr[1] != "") link=arr[1]
			}
			/<span class="name">/ {
				match($0, /<span class="name">([^<]+)<\/span>/, arr)
				if (arr[1] != "") ver=arr[1]
			}
			/<span class="count">/ {
				match($0, /<span class="count">([0-9,]+)<\/span>/, arr)
				gsub(",", "", arr[1])
				count=arr[1]
			}
			/<\/tr>/ {
				if (ver != "" && count != "" && count+0 > 5 && link != "") {
					print ver "|https://sourceforge.net" link
				}
			}'
		`
	cmd := exec.Command("bash", "-c", bashScript)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error ejecutando scraping: %v", err)
	}
	raw := splitLines(string(out))
	var versiones []string
	for _, v := range raw {
		if v == "" {
			continue
		}
		// sep := "|"
		idx := -1
		for i, c := range v {
			if c == '|' {
				idx = i
				break
			}
		}
		if idx > 0 {
			versiones = append(versiones, v[:idx])
		} else {
			versiones = append(versiones, v)
		}
	}
	if len(versiones) == 0 {
		return nil, fmt.Errorf("no se encontraron versiones con más de 5 descargas")
	}
	return versiones, nil
}

func splitLines(s string) []string {
	var res []string
	curr := ""
	for _, c := range s {
		if c == '\n' {
			if curr != "" {
				res = append(res, curr)
				curr = ""
			}
		} else {
			curr += string(c)
		}
	}
	if curr != "" {
		res = append(res, curr)
	}
	return res
}
