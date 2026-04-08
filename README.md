# xampp-tui

**xampp-tui** es una interfaz de usuario en modo texto (TUI) para gestionar servicios tipo XAMPP (Apache, MySQL, FTP) desde la terminal, desarrollada en Go utilizando las librerías [Bubble Tea](https://github.com/charmbracelet/bubbletea) y [Lipgloss](https://github.com/charmbracelet/lipgloss).

## Características

- **Visualización de servicios**: Muestra el estado de Apache, MySQL y FTP en una tabla con columnas para Servicio, PID, Puerto y Configuración.
- **Navegación con teclado**: Usa las teclas de dirección o `w`, `a`, `s`, `d` para moverte entre filas y columnas.
- **Cambio de estado**: Pulsa `Enter` o `Espacio` sobre un servicio para alternar entre "running" y "stopped".
- **Área de logs**: Visualiza mensajes o logs de acciones en la parte inferior de la interfaz.
- **Diseño atractivo**: Usa cajas y estilos para resaltar la información y la selección actual.
- **Atajos**:
    - `q` o `Ctrl+C`: Salir
    - `↑/↓/w/s`: Moverse entre servicios
    - `←/→/a/d`: Cambiar de columna
    - `Enter`/`Espacio`: Cambiar estado o activar acción según la columna

## Estructura real del proyecto

```
xampp-tui
├── README.md
├── assets
│   └── README.md
├── cmd
│   ├── lampp-tui
│   │   ├── downloads
│   │   └── main.go
│   └── logs
│       └── app.log
├── go.mod
├── go.sum
├── install.sh
└── internal
    ├── installer
    │   ├── downloader.go
    │   └── versions.go
    ├── logger
    │   └── logger.go
    ├── tui
    │   ├── model.go
    │   ├── render.go
    │   ├── styles.go
    │   ├── update.go
    │   └── view.go
    └── xampp
        ├── service.go
        └── validator.go
```

**Notas sobre la estructura:**

- El entrypoint principal está en `cmd/lampp-tui/main.go`.
- Los instaladores y lógica de descarga están en `internal/installer/`.
- El logger propio está en `internal/logger/logger.go` y los logs se guardan en `cmd/logs/app.log`.
- La lógica de la interfaz TUI está en `internal/tui/`.
- La lógica de gestión de servicios XAMPP está en `internal/xampp/`.
- Los recursos/documentación adicional van en `assets/`.
- El directorio `downloads/` dentro de `cmd/lampp-tui/` almacena descargas temporales de instaladores.
- El archivo `install.sh` automatiza la instalación y el servicio systemd.

Esta estructura sigue buenas prácticas de Go y Bubble Tea, separando claramente la UI, lógica de negocio y utilidades.

## Versionado

Este proyecto sigue [SemVer](https://semver.org/lang/es/) para el control de versiones. Usa etiquetas Git para marcar lanzamientos:

```bash
git tag v1.0.0
git push origin v1.0.0
```

Consulta el historial de versiones en la sección de [Releases](https://github.com/ramirezDg/lampp-tui/releases) de GitHub.

## Documentación avanzada

Encuentra más detalles en [docs/README.md](docs/README.md).

## Ejemplo de uso

Al ejecutar el programa, verás una tabla como esta:

```
Servicio           PID        Puerto      Config
Apache             0          80          httpd.conf
MySql              0          3306        my.ini
FTP                0          21          vsftpd.conf

...
Logs De Acciones
[q, ctrl+c] quit | [↑, w, k] up | [↓, s, j] down | [enter, space] toggle state
```

## Requisitos

- Go 1.18 o superior
- Linux (probado en terminal)
- gawk (procesamiento de texto en scripts y utilidades)
    - Instálalo en Debian/Ubuntu con: `sudo apt install gawk`
    - En Arch/Manjaro: `sudo pacman -S gawk`
    - En Fedora: `sudo dnf install gawk`
- Dependencias:
    - [Bubble Tea](https://github.com/charmbracelet/bubbletea)
    - [Lipgloss](https://github.com/charmbracelet/lipgloss)

## Instalación

1. Clona el repositorio:
    ```bash
    git clone https://github.com/ramirezDg/lampp-tui.git
    cd lampp-tui
    ```
2. Instala dependencias:

    ```bash
    go mod tidy
    ```

3. Ejecuta la aplicación:

    ```bash
    go run cmd/lampp-tui/main.go
    ```

    O bien, instala y ejecuta como servicio con:

    ```bash
    ./install.sh
    sudo systemctl status xampp-tui
    sudo journalctl -u xampp-tui -f
    ```

## Personalización

Puedes modificar los servicios, puertos y configuraciones iniciales en la función `InitialModel()` de `internal/tui/model.go`.

## Licencia

MIT

---

Hecho con [Bubble Tea](https://github.com/charmbracelet/bubbletea) y [Lipgloss](https://github.com/charmbracelet/lipgloss).
