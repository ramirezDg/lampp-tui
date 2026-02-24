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

## Estructura del código

- **main.go**: Lógica principal de la TUI, manejo de eventos de teclado, renderizado de la tabla y navegación.
- **ui.go**: Funciones auxiliares para el diseño visual (banner, cajas, área de texto, pie de página).

## Ejemplo de uso

Al ejecutar el programa, verás una tabla como esta:

```
Servicio           PID        Puerto      Config
╭───────────────╮  0         80          httpd.conf
│    Apache     │  0         3306        my.ini
│   stopped     │  0         21          vsftpd.conf
╰───────────────╯
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