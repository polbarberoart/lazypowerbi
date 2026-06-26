# lazypowerbi

A terminal UI (TUI) for exploring Power BI workspaces, reports, datasets and dashboards — without leaving your terminal.

> **Proyecto didáctico.** lazypowerbi es un proyecto personal desarrollado en Go con ayuda de agentes de IA (Claude Code) como ejercicio para aprender el lenguaje y el desarrollo de herramientas de terminal. Está inspirado en [lazyazure](https://github.com/matsest/lazyazure) pero orientado a explorar recursos de Power BI.

---

## ¿Qué hace?

lazypowerbi te permite navegar por tus workspaces e items de Power BI desde el terminal, accediendo rápidamente a información que normalmente requiere navegar por el portal web: IDs de datasets, URLs de conexión, metadatos de reports, etc.

```
┌─────────────────────────────────────────────────────────────────┐
│  lazypowerbi                                                    │
├─────────────────┬───────────────────────────────────────────────┤
│  Workspaces     │  Details                                      │
│                 │                                               │
│ > Sales         │  ID:       r-1                               │
│   Finance       │  Name:     Sales Report                      │
│   Marketing     │  Kind:     Report                            │
├─────────────────┤  Workspace: ws-1                             │
│  Items          │  WebURL:   https://...                       │
│                 │                                               │
│ > Sales Report  │                                               │
│   Sales Dataset │                                               │
│   Sales Dash..  │                                               │
├─────────────────┴───────────────────────────────────────────────┤
│  user@example.com  │  tenant-id  │  ↑↓/jk: nav  Tab  q: quit  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Características

- Navegar workspaces e items (reports, datasets, dashboards) con teclado
- Ver detalles de cada elemento: ID, nombre, tipo, workspace, URL
- Autenticación con Azure mediante `DefaultAzureCredential`
- Carga asíncrona de datos sin bloquear la interfaz

---

## Requisitos

- [Go 1.21+](https://golang.org/dl/)
- [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli) — para autenticación
- Acceso a uno o más workspaces de Power BI

---

## Inicio rápido

### 1. Autenticarse con Azure

```bash
az login
```

Si trabajas con varios tenants, especifica el tuyo:

```bash
az login --tenant <tenant-id>
```

### 2. Clonar el repositorio

```bash
git clone https://github.com/polbarberoart/lazypowerbi.git
cd lazypowerbi
```

### 3. Instalar dependencias

```bash
go mod download
```

### 4. Ejecutar

```bash
go run .
```

O compilar un binario:

```bash
go build -o lazypowerbi .
./lazypowerbi
```

---

## Controles

| Tecla | Acción |
|-------|--------|
| `↑` / `k` | Subir en la lista |
| `↓` / `j` | Bajar en la lista |
| `Tab` | Cambiar panel activo (Workspaces ↔ Items) |
| `q` / `Ctrl+C` | Salir |

---

## Estructura del proyecto

```
lazypowerbi/
├── main.go                  # Punto de entrada
├── internal/
│   └── ui/                  # TUI (gocui)
│       ├── app.go           # Struct principal y ciclo de vida
│       ├── layout.go        # Distribución de paneles
│       ├── views.go         # Render de contenido
│       ├── keys.go          # Keybindings
│       ├── load.go          # Carga asíncrona de datos
│       └── interfaces.go    # Interfaces para clientes
├── pkg/
│   ├── domain/              # Tipos de dominio (Workspace, Item, User)
│   └── powerbi/             # Cliente HTTP para la API REST de Power BI
└── vendor_gocui/            # Fork de gocui (de jesseduffield/lazygit)
```

---

## Tecnologías

- **[Go](https://golang.org/)** — lenguaje principal
- **[gocui](https://github.com/jesseduffield/gocui)** — librería TUI (fork de lazygit)
- **[Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go)** — autenticación con Azure AD
- **Power BI REST API** — sin SDK oficial para Go; el cliente está construido desde cero en `pkg/powerbi`

---

## Inspiración

Este proyecto está inspirado en [lazyazure](https://github.com/matsest/lazyazure) de [@matsest](https://github.com/matsest), una herramienta similar para explorar recursos de Azure. lazypowerbi aplica los mismos principios — navegación rápida con teclado, TUI minimalista — pero orientados al ecosistema de Power BI.

---

## Roadmap

Ver [`docs/ROADMAP.md`](docs/ROADMAP.md) para el plan de mejoras y nuevas funcionalidades previstas.

---

## Licencia

MIT
