# Feature: fecha límite opcional en las tareas

Origen: issue [#3](https://github.com/SaraFloresFermosel/desarrollo-agentico-todo-03/issues/3) — «Quiero añadir fecha limite a las tasks». Cuerpo: «Vale con que sea simplemente un dia. Sin hora.»

## Encargo y criterios de aceptación

Requisitos explícitos:

- Una tarea puede tener una fecha límite.
- La fecha límite es solo un día (sin hora).

Inferencias (decisiones de este plan):

- La fecha límite es **opcional**: las tareas existentes y las nuevas sin fecha siguen siendo válidas.
- Se puede indicar al crear y al editar una tarea, y borrarla dejando el campo vacío al editar.
- Se muestra en la lista junto al título.

Criterios de aceptación:

1. El formulario de creación de `/` tiene un campo de fecha (`<input type="date" name="due_date">`) opcional con etiqueta accesible «Fecha límite».
2. Crear una tarea con `due_date=2026-11-02` la guarda con esa fecha; sin `due_date` (o vacío) se guarda sin fecha.
3. Un `due_date` que no sea `AAAA-MM-DD` válido responde 422, no crea la tarea y muestra «La fecha límite no es válida.» conservando título y fecha introducidos.
4. La lista muestra, para las tareas con fecha, «Vence el 02/11/2026» dentro de un `<time datetime="2026-11-02">`; las tareas sin fecha no muestran nada.
5. `/tasks/{id}/edit` muestra el campo de fecha precargado (`value="2026-11-02"`) o vacío.
6. `PUT /tasks/{id}` guarda título y fecha; un `due_date` vacío o ausente deja la tarea sin fecha; uno inválido responde 422 sin modificar la tarea. Completada, posición, `id` y `created_at` siguen sin poder cambiarse desde el formulario.
7. Las tareas existentes quedan sin fecha tras la migración y la reversión de la migración elimina la columna.

## Comportamiento actual y propuesta

Actual: `models.Task` tiene `ID`, `Title`, `Completed`, `Position`, `CreatedAt`, `UpdatedAt`. La tabla `tasks` no tiene ninguna columna de fecha de vencimiento. `TasksCreate` enlaza el formulario en `models.Task` (solo `title` es enlazable); `TasksUpdate` enlaza un struct anónimo con `title` y actualiza solo el título.

Propuesta: añadir la columna nullable `due_date` de tipo `date`, el campo `DueDate *time.Time` en el modelo (sin enlace automático de formulario), parsear explícitamente el valor `due_date` del formulario en las acciones con formato `2006-01-02`, y mostrar/editar la fecha en las plantillas existentes.

## Alcance

Incluido:

- Migración fizz que añade `tasks.due_date` (`date`, nullable).
- Campo de modelo y helpers de formato/parseo.
- Creación y edición de la fecha en `TasksCreate`/`TasksUpdate`.
- Campo de fecha en los formularios de alta y edición, y visualización en la lista.
- Estilos mínimos para el campo y la etiqueta de vencimiento.
- Tests de modelo y de acciones.

Excluido:

- Resaltar tareas vencidas o próximas a vencer.
- Ordenar o filtrar por fecha límite (el orden sigue siendo `position`).
- Hora, zona horaria, recordatorios o notificaciones.
- Cambios en reordenación, toggle o borrado.
- Cambios en `fixtures/tasks.toml` y en las migraciones de seed existentes.

## Diseño por capas

### Persistencia

Nueva migración `migrations/20261026000000_add_due_date_to_tasks.up.fizz`:

```
add_column("tasks", "due_date", "date", {"null": true})
```

y su `down`:

```
drop_column("tasks", "due_date")
```

Sigue el patrón de `20261019000000_add_position_to_tasks`. Sin default: las filas existentes quedan a `NULL`.

### Modelo (`models/task.go`)

- Campo nuevo tras `Position`:
  `DueDate *time.Time `json:"due_date" db:"due_date" form:"-"``
  (se usa puntero para representar `NULL` sin añadir `gobuffalo/nulls` como dependencia directa; `form:"-"` para que `c.Bind` no lo toque).
- Constante `DueDateLayout = "2006-01-02"`.
- `var ErrInvalidDueDate = errors.New("fecha límite no válida")`.
- `func ParseDueDate(raw string) (*time.Time, error)`: recorta espacios; vacío → `nil, nil`; si `time.Parse(DueDateLayout, raw)` falla → `nil, ErrInvalidDueDate`; si no, devuelve un puntero a la fecha (UTC, medianoche).
- `func (t Task) DueDateValue() string`: `""` si `DueDate == nil`, si no `t.DueDate.Format(DueDateLayout)` (para `value` y `datetime`).
- `func (t Task) DueDateLabel() string`: `""` si nil, si no `t.DueDate.Format("02/01/2006")`.
- `Validate` no cambia (la fecha es opcional y su formato se valida al parsear).

### Acciones (`actions/tasks.go`)

En cada acción:

- `TasksCreate`: tras `c.Bind(task)` y recortar el título, leer `raw := c.Param("due_date")` y llamar `models.ParseDueDate(raw)`.
  - Si devuelve error: construir `verrs := validate.NewErrors()`, añadir `verrs.Add("due_date", "La fecha límite no es válida.")`, y renderizar el índice con 422 igual que en la rama de errores existente (recargando `tasks`), poniendo además `c.Set("dueDate", raw)` para conservar el valor introducido. No se crea nada.
  - Si no: `task.DueDate = dueDate` y seguir con el flujo actual (`position`, `ValidateAndCreate`).
  - Para no duplicar el render del índice con errores, extraer la rama existente a un helper `renderIndexWithErrors(c, tx, task, verrs, dueDate string) error` que haga la consulta ordenada, los `c.Set` y el `Render(422, …)`, y usarlo en ambos casos.
- `TasksIndex`: añadir `c.Set("dueDate", "")`.
- `TasksCreate` (rama de errores de título): `c.Set("dueDate", task.DueDateValue())` vía el helper.
- `TasksUpdate`: añadir `DueDate string `form:"due_date" json:"due_date"`` al struct `input`; parsear con `models.ParseDueDate(input.DueDate)`.
  - Error: `verrs` con el mensaje de `due_date`, `c.Set("task", task)` (con el título ya asignado del input), `c.Set("dueDate", input.DueDate)`, `c.Set("errors", verrs)`, render 422 de `tasks/edit.plush.html` sin actualizar.
  - Correcto: `task.DueDate = dueDate` y seguir con `ValidateAndUpdate`; en la rama de errores existente poner también `c.Set("dueDate", input.DueDate)`.
  - Actualizar el comentario de la función: «actualiza el título y la fecha límite, conservando los demás datos de la tarea».
- `TasksEdit`: `c.Set("dueDate", task.DueDateValue())`.

La variable de plantilla `dueDate` (string) contiene siempre el valor a mostrar en el `<input type="date">`, ya sea el guardado o el que envió el usuario.

### Renderizado

`templates/tasks/index.plush.html`:

- En cada `<li>`, tras `<span class="task-title">…</span>`:
  ```
  <%= if (t.DueDateValue() != "") { %>
    <span class="task-due">Vence el <time datetime="<%= t.DueDateValue() %>"><%= t.DueDateLabel() %></time></span>
  <% } %>
  ```
- En el formulario `.new-task`, dentro de `.new-task-row`, antes del botón «Añadir»:
  ```
  <label for="task-due-date" class="sr-only">Fecha límite</label>
  <input id="task-due-date" type="date" name="due_date" value="<%= dueDate %>" title="Fecha límite (opcional)"
    <%= if (len(errors.Get("due_date")) > 0) { %>aria-invalid="true" aria-describedby="due-date-error"<% } %>>
  ```
- Bloque de errores: el `aria-describedby="title-error"` y `autofocus` del input de título solo deben activarse si hay errores de `title` (`len(errors.Get("title")) > 0`), y añadir tras el párrafo `title-error` uno análogo:
  ```
  <%= if (len(errors.Get("due_date")) > 0) { %>
    <p id="due-date-error" class="notice field-error" role="alert"><%= for (msg) in errors.Get("due_date") { %><%= msg %><% } %></p>
  <% } %>
  ```
  y envolver el `title-error` existente en la condición de `errors.Get("title")` en vez de `errors.HasAny()`.

`templates/tasks/edit.plush.html`:

- Tras el bloque del título, añadir:
  ```
  <label for="task-due-date">Fecha límite <span class="hint">(opcional)</span></label>
  <input id="task-due-date" type="date" name="due_date" value="<%= dueDate %>"
    <%= if (len(errors.Get("due_date")) > 0) { %>aria-invalid="true" aria-describedby="due-date-error"<% } %>>
  ```
  con su párrafo `due-date-error` igual que en el índice.
- Cambiar las condiciones de error del título a `len(errors.Get("title")) > 0` para que un error de fecha no marque el título como inválido.
- Cambiar el texto de ayuda a «Hasta 200 caracteres. Deja la fecha vacía si la tarea no tiene fecha límite.» (manteniendo el `id="title-help"`).

### CSS (`public/assets/app.css`)

- Extender la regla `input[type="text"]` a `input[type="text"], input[type="date"]` para que el campo de fecha tenga el mismo aspecto.
- `.new-task-row input[type="date"] { flex: 0 0 auto; width: auto; }` para que el título siga ocupando el espacio restante (y comprobar que el `@media` de pantallas estrechas con `flex-wrap` deja la fecha en una línea razonable; si `.new-task-row input { flex-basis: … }` afecta a la fecha, limitar ese selector a `input[type="text"]`).
- `.task-due { font-size: 0.85rem; color: var(--color-muted); white-space: nowrap; }`.
- Subir la versión de caché `?v=4` → `?v=5` del CSS en `templates/application.plush.html` (el JS no cambia, pero se puede dejar igual).

No hay cambios en JavaScript.

## Archivos relevantes

| Ruta | Responsabilidad |
| --- | --- |
| `migrations/20261026000000_add_due_date_to_tasks.up.fizz` / `.down.fizz` | Nueva columna `due_date` (crear). |
| `models/task.go` | Campo `DueDate`, `ParseDueDate`, `DueDateValue`, `DueDateLabel`. |
| `models/task_test.go` | Tests de parseo, formato y persistencia de la fecha. |
| `actions/tasks.go` | Lectura/validación de `due_date` en crear y actualizar; `dueDate` para las plantillas. |
| `actions/tasks_test.go` | Tests HTTP de los criterios de aceptación. |
| `templates/tasks/index.plush.html` | Mostrar la fecha en la lista y el campo en el alta. |
| `templates/tasks/edit.plush.html` | Campo de fecha en la edición. |
| `public/assets/app.css` | Estilos del campo de fecha y de `.task-due`. |
| `templates/application.plush.html` | Versión de caché del CSS. |

## Implementación paso a paso

1. Crear la migración `up`/`down` descrita.
2. Modificar `models/task.go` (campo, constante, error, `ParseDueDate`, `DueDateValue`, `DueDateLabel`).
3. Modificar `actions/tasks.go`: helper `renderIndexWithErrors`, `TasksIndex`, `TasksCreate`, `TasksEdit`, `TasksUpdate` como se describe.
4. Modificar las plantillas `index` y `edit`.
5. Ajustar `public/assets/app.css` y la versión en `application.plush.html`.
6. Añadir los tests de la sección siguiente.
7. Ejecutar `gofmt -l .` (y `gofmt -w` sobre los archivos tocados si hace falta) y el resto de validaciones.

## Tests y validación

Tests de modelo (`models/task_test.go`):

- `Test_ParseDueDate`: `""` y `"  "` → `nil, nil`; `"2026-11-02"` → fecha 2026-11-02; `"2026-02-30"`, `"02/11/2026"` y `"mañana"` → `ErrInvalidDueDate`.
- `Test_Task_DueDate_Persists`: crear tarea con `DueDate`, recargar y comprobar `DueDateValue() == "2026-11-02"` y `DueDateLabel() == "02/11/2026"`; una tarea sin fecha recarga con `DueDate == nil` y ambos helpers devuelven `""`.

Tests de acciones (`actions/tasks_test.go`):

- `Test_TasksCreate_WithDueDate`: POST `url.Values{"title": {"Con fecha"}, "due_date": {"2026-11-02"}}` → 303; la tarea guardada tiene `DueDateValue() == "2026-11-02"`.
- `Test_TasksCreate_WithoutDueDate`: POST solo con título (y con `due_date` vacío) → 303 y `DueDate == nil`.
- `Test_TasksCreate_InvalidDueDate`: `due_date=no-es-fecha` → 422, cuerpo contiene «La fecha límite no es válida.» y `value="Con fecha"`; no se crea ninguna tarea.
- `Test_TasksIndex_ShowsDueDate`: con una tarea con fecha y otra sin ella, el cuerpo contiene `datetime="2026-11-02"` y «02/11/2026» una sola vez (`strings.Count`).
- `Test_TasksEdit_PrefillsDueDate`: el formulario contiene `name="due_date"` y `value="2026-11-02"`.
- `Test_TasksUpdate_SetsAndClearsDueDate`: PUT con `due_date=2026-12-24` guarda la fecha; un segundo PUT con `due_date=""` la deja a `nil`; `Completed` no cambia.
- `Test_TasksUpdate_InvalidDueDate`: 422 con el mensaje, y la tarea recargada conserva título y fecha originales.
- Los tests existentes (incluido `Test_TasksUpdate_PersistsTitleOnly`, que no envía `due_date`) deben seguir pasando sin cambios.

Validación (requiere PostgreSQL accesible según `database.yml` y `GO_ENV=test`):

```
gofmt -l .
go vet ./...
soda migrate -e test        # o el mecanismo equivalente disponible; aplica la nueva migración
go test -p 1 ./...
go build ./...
```

Comprobación manual opcional con `/env:start`: crear una tarea con fecha, verla en la lista, editarla, borrar la fecha.

## Riesgos y recuperación

- **Escaneo de `date` en `*time.Time`**: el driver de PostgreSQL devuelve `time.Time` a medianoche UTC; los helpers formatean sin conversión de zona, así que no hay desplazamiento de día. Si un test mostrara un día desplazado, formatear con `t.DueDate.UTC()`.
- **`c.Param("due_date")` en `TasksCreate`**: Buffalo resuelve parámetros de formulario con `Param`; si no lo hiciera en tests, usar `c.Request().FormValue("due_date")`.
- **Plush y punteros**: las plantillas solo usan los métodos `DueDateValue`/`DueDateLabel` (receptor por valor), evitando desreferenciar `nil` en Plush.
- **Migración**: aditiva y nullable; reversible con el `down`. Si falla en un entorno, `soda migrate down` revierte solo esta migración.
- **Navegadores sin `type="date"`**: muestran un campo de texto; el servidor exige `AAAA-MM-DD` y responde 422 con mensaje claro en otro formato.
- **Accesibilidad de errores**: separar los errores de título y fecha evita que un error de fecha marque como inválido el título; los tests de `aria-describedby="title-error"` existentes deben seguir pasando.
