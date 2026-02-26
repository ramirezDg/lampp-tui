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

## Estructura recomendada de carpetas

```
lampp-tui/
├── main.go           # Lógica principal de la TUI
├── ui.go             # Funciones visuales y helpers
├── go.mod / go.sum   # Dependencias
├── README.md         # Documentación principal
├── docs/             # Documentación avanzada
│   └── README.md
├── assets/                 # Recursos estáticos (imágenes, ejemplos)
│   └── README.md
├── cmd/                    # Entrypoints ejecutables
│   └── lampp-tui/main.go   # main para la TUI
├── internal/               # Lógica interna no exportada
│   ├── tui/                # UI y componentes Bubble Tea
│   │   ├── model.go        # Modelo de estado de la UI
│   │   ├── update.go       # Lógica de actualización (mensajes)
│   │   ├── view.go         # Renderizado visual
│   │   └── styles.go       # Estilos Lipgloss
│   ├── services/           # Lógica de negocio y servicios
│   │   ├── downloader.go   # Descarga de archivos
│   │   ├── http_client.go  # Cliente HTTP reutilizable
│   │   ├── xampp.go        # Control y estado de XAMPP
│   │   └── version_fetcher.go # Obtención de versiones XAMPP
│   └── state/              # Estado global (opcional)
│       └── app_state.go
└── docs/                   # Documentación avanzada
   └── README.md
```

Esta estructura sigue las mejores prácticas de Go y Bubble Tea:

- Separación clara entre UI, servicios y estado.
- Lógica de UI en `internal/tui/`.
- Lógica de negocio y servicios en `internal/services/`.
- Entrypoint limpio en `cmd/`.
- Recursos y documentación organizados.

Puedes extender `internal/` para más módulos o componentes si tu proyecto crece.

Puedes crear las carpetas `internal/` y `assets/` si tu proyecto crece o necesitas organizar mejor el código y recursos.

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
    go run main.go ui.go
    ```

## Personalización

Puedes modificar los servicios, puertos y configuraciones iniciales en la función `initialModel()` de `main.go`.

## Licencia

MIT

---

Hecho con [Bubble Tea](https://github.com/charmbracelet/bubbletea) y [Lipgloss](https://github.com/charmbracelet/lipgloss).
