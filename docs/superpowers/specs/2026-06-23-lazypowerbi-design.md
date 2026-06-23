# Diseño: lazypowerbi

**Fecha:** 2026-06-23
**Estado:** Aprobado para planificación

## 1. Objetivo y alcance

Crear `lazypowerbi`, una TUI para explorar workspaces de Power BI, como proyecto nuevo e independiente en `C:\Repos\lazypowerbi`, tomando como punto de partida la arquitectura de `lazyazure` (este repositorio). `lazyazure` es un proyecto en producción y no se modifica como parte de este trabajo — solo se lee/copia desde aquí.

Este proyecto tiene un segundo objetivo explícito además del producto en sí: es un proyecto **didáctico** para que el usuario (Python, nuevo en Go) aprenda Go construyéndolo paso a paso, con explicaciones de conceptos de lenguaje y de arquitectura a medida que se implementan, no solo código entregado. Esto condiciona cómo se ejecutará el plan de implementación (ver nota en sección 8), más que el propio diseño técnico.

### Dentro de alcance
- Navegación jerárquica: Workspaces → Items (Datasets/Reports/Dashboards/Dataflows mezclados) → panel de detalle.
- Autenticación contra la API REST de Power BI **a nivel de usuario** (`/v1.0/myorg/...`), vía `azidentity` (mismo mecanismo que lazyazure: az login, env vars, managed identity, etc.).
- Modo demo con datos mock (paridad con `LAZYAZURE_DEMO=1/2`).
- Búsqueda/filtro en paneles, copiar/abrir enlace al portal de Power BI, vistas Summary/JSON en el panel de detalle.
- Reutilización máxima de código no específico de dominio: `pkg/gui/panels`, `pkg/utils` (salvo `portal_urls.go`), `pkg/tasks`, `vendor_gocui`.

### Fuera de alcance (iteraciones futuras)
- Power BI Admin API (`/admin/groups`) — solo API de usuario por ahora. Quien no sea miembro de un workspace no lo verá; es la paridad de comportamiento más fiel con cómo lazyazure depende de RBAC real del usuario.
- Operaciones de escritura/modificación sobre recursos de Power BI.
- Autenticación dedicada de service principal más allá de lo que ya cubre `DefaultAzureCredential`.
- Migrar `tools/`, `scripts/`, `AGENTS.md`, `PLAN.md`, README, assets de demo de lazyazure — se crean versiones propias cuando corresponda, no como parte de este diseño.

## 2. Modelo de dominio (`pkg/domain`)

```go
type Workspace struct {
    ID                    string
    Name                  string
    Type                  string // "Workspace" | "PersonalGroup"
    IsReadOnly            bool
    IsOnDedicatedCapacity bool
    CapacityID            string
}

type Item struct {
    ID          string
    Name        string
    Kind        string // "Dataset" | "Report" | "Dashboard" | "Dataflow"
    WorkspaceID string
    WebURL      string
    Properties  map[string]interface{}
}
```

Mismo patrón que `Subscription`/`Resource` en lazyazure: cada tipo implementa `DisplayString()`, `GetID()`, `GetDisplaySuffix()` (en `Item`, el sufijo es `Kind`).

`domain.User` se reutiliza **sin cambios** — los claims del JWT de Entra ID (`tid`, `oid`, `upn`, `name`) son los mismos independientemente del scope del token solicitado.

No se necesita un equivalente a `pkg/resources/display_names.go` (mapeo ARM type → nombre legible vía JSON embebido): solo hay 4 `Kind` fijos y conocidos, basta un mapa literal pequeño en `domain`.

## 3. Cliente de la API de Power BI (`pkg/powerbi`)

No existe SDK oficial de Go para la API REST de Power BI (a diferencia de Azure Resource Manager, que sí tiene `armresources`/`armsubscriptions` generados por Microsoft). Se construyen las peticiones HTTP directamente con `net/http` + `encoding/json`.

**`client.go`** (equivalente a `azure.Client`):
```go
type Client struct {
    credential azcore.TokenCredential
    httpClient *http.Client
}
```
- `NewClient()` usa `azidentity.NewDefaultAzureCredential`, igual mecanismo de auth que lazyazure.
- Scope de token: `"https://analysis.windows.net/powerbi/api/.default"` (en vez de `management.azure.com`).
- `GetUserInfo`/`parseAzureToken` se reutilizan sin cambios.
- Método interno `doRequest(ctx, method, path string) ([]byte, error)` centraliza: obtención de token, construcción de la petición a `https://api.powerbi.com/v1.0/myorg` + `path`, header `Authorization: Bearer`, ejecución y manejo de errores HTTP. Evita repetir auth+HTTP en cada cliente de recurso.

**`workspaces.go`** (equivalente a `subscriptions.go`): `ListWorkspaces(ctx) ([]*domain.Workspace, error)` vía `GET /groups`, con struct intermedio para deserializar el JSON y mapeo a `domain.Workspace`.

**`items.go`** (equivalente a `resources.go`): `ListItemsByWorkspace(ctx, workspaceID) ([]*domain.Item, error)` lanza 4 peticiones (`/groups/{id}/datasets`, `/reports`, `/dashboards`, `/dataflows`), cada una a su struct intermedio, fusionadas en una sola lista con `Kind` etiquetado según el endpoint de origen.

**`factory.go`**: implementa `gui.PowerBIClientFactory`, mismo rol que `azure/factory.go`.

## 4. GUI (`pkg/gui`)

Cambio estructural principal respecto a lazyazure: de **3 paneles de lista apilados** (Subscriptions/ResourceGroups/Resources) a **2** (Workspaces/Items), porque la jerarquía de Power BI tiene un nivel menos que Azure.

**Paneles (`setupViews`):** `auth`, `workspaces` (antes `subscriptions`), `items` (antes `resources`; se elimina el panel intermedio `resourcegroups`), `main`, `status`. Reparto de alturas de 2 franjas en vez de 3.

**Struct `Gui`:** se eliminan `selectedRG`, `resourceGroups`, `rgList`, `resourceGroupsView`, `loadingRGs` y equivalentes. Se renombra `selectedSub`→`selectedWorkspace`, `subList`→`workspaceList`, `resList`→`itemList`.

**Navegación:** `onWorkspaceEnter` carga Items directamente (sin paso intermedio de ResourceGroups). Se elimina `onRGEnter` y sus bindings. `switchPanel`/`switchPanelReverse` ciclan por 3 vistas (`workspaces → items → main`) en vez de 4.

**Caché (`cache.go`):** de 2 niveles (RG cache + Resource cache) a 1 nivel (Items por Workspace, con TTL) — más simple que el original.

**Se reutiliza sin tocar:** `pkg/gui/panels/*` (FilteredList genérico sobre `T any`, SearchBar, MainPanelSearch — agnósticos de dominio), `pkg/gui/interfaces.go` como patrón (renombrando a `PowerBIClient`/`PowerBIClientFactory`/`WorkspacesClient`/`ItemsClient`), toda la lógica de scroll/tabs/colores/búsqueda del panel principal.

## 5. `pkg/utils`

Reutilizado prácticamente sin cambios: `logger.go`, `clipboard.go`, `browser.go`, `metrics.go`. Cambia solo el prefijo de variables de entorno (`LAZYAZURE_DEBUG` → `LAZYPOWERBI_DEBUG`) y la ruta de log (`~/.lazyazure/` → `~/.lazypowerbi/`).

`portal_urls.go` cambia de contenido: construye URLs de `app.powerbi.com` en vez de `portal.azure.com`. No requiere `tenantID` en la firma (Power BI usa el contexto de sesión del navegador):
```go
func BuildWorkspacePortalURL(workspaceID string) string
func BuildItemPortalURL(workspaceID, kind, itemID string) string
```

## 6. Modo demo (`pkg/demo`)

Mismo patrón: `DemoClient` implementa las mismas interfaces que el cliente real, devolviendo datos fijos en memoria.
```go
type DemoData struct {
    User       *domain.User
    Workspaces []*domain.Workspace
    Items      map[string][]*domain.Item // key: workspace ID
}
```
Un solo nivel de mapa (lazyazure tenía dos anidados: ResourceGroups + Resources). `LAZYPOWERBI_DEMO=1` (pequeño) / `=2` (grande), mismo mecanismo que lazyazure.

## 7. Estructura de carpetas y repositorio

- Module path: `github.com/polbarbero/lazypowerbi`
- Ubicación: `C:\Repos\lazypowerbi` (repo nuevo e independiente desde el inicio)

```
C:\Repos\lazypowerbi\
├── go.mod                  (module github.com/polbarbero/lazypowerbi)
├── main.go                 (adaptado: env vars LAZYPOWERBI_*, textos de help)
├── main_test.go
├── pkg/
│   ├── domain/              (Workspace, Item, User copiado tal cual)
│   ├── powerbi/             (antes "azure": client.go, workspaces.go, items.go, factory.go)
│   ├── demo/                (DemoData + DemoClient adaptados)
│   ├── gui/                 (gui.go, cache.go, interfaces.go — adaptado a 2 niveles de lista)
│   │   └── panels/          (copiado sin cambios)
│   ├── tasks/               (copiado sin cambios)
│   └── utils/               (copiado; portal_urls.go adaptado, env vars renombradas)
├── vendor_gocui/            (copiado sin cambios — fork vendorizado de gocui)
├── docs/
├── .github/                 (workflows adaptados)
├── .goreleaser.yml          (adaptado)
└── Makefile                 (adaptado)
```

No se migran `tools/`, `scripts/`, `AGENTS.md`, `PLAN.md`, `README.md`, `demo.gif/png` — son específicos de lazyazure; se redactan versiones propias para lazypowerbi fuera de este diseño.

## 8. Decisiones de auth/API confirmadas

- **API de usuario únicamente** (`/v1.0/myorg/...`), sin soporte de Admin API en esta versión. Justificación discutida: el rol de admin de Fabric/Power BI es un permiso separado del rol de admin de Entra ID, y exigirlo limitaría la adopción; la API de usuario da paridad de comportamiento con cómo lazyazure depende de RBAC real sin requerir configuración adicional.
- Tipos de item soportados: Datasets, Reports, Dashboards, Dataflows (los 4 endpoints estándar accesibles a cualquier miembro de un workspace).

## 9. Nota para la fase de implementación

El usuario está aprendiendo Go viniendo de Python. El plan de implementación y su ejecución deben tratar este proyecto como didáctico: explicar conceptos de Go sin asumir nada (punteros vs valores, interfaces, goroutines/canales, manejo de errores `if err != nil`, structs, generics, paquetes/módulos, `defer`, valores zero), y avanzar de forma incremental con explicaciones en cada paso en lugar de entregar bloques grandes de código ya terminados — incluso cuando gran parte sea adaptación directa de código ya existente en lazyazure.

Se han instalado las skills `samber/cc-skills-golang@{golang-code-style, golang-error-handling, golang-testing, golang-design-patterns, golang-performance, golang-security}` para apoyar esta fase con convenciones idiomáticas de Go.
