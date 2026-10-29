# Modelo Task

Verificado: 2026-10-29 contra `43f3393`.

## Hechos

- `Task` es la única entidad de la aplicación y vive en la tabla `tasks` — `models/task.go:29`, `migrations/20260901000000_create_tasks.up.fizz:1`
- Sus campos son `ID` (uuid), `Title`, `Completed`, `Position`, `DueDate`, `CreatedAt` y `UpdatedAt` — `models/task.go:30-36`
- Desde un formulario solo se rellena `Title` con `c.Bind`; los demás campos llevan `form:"-"` — `models/task.go:30-36`
- `DueDate` es un `*time.Time` opcional; `nil` significa que la tarea no tiene fecha límite — `models/task.go:34`
- En la base de datos, `tasks.due_date` es una columna `date` nullable, sin valor por defecto — `migrations/20261026000000_add_due_date_to_tasks.up.fizz:1`
- `tasks.title` es `string` de tamaño 200, `completed` vale `false` por defecto y `position` vale `0` por defecto — `migrations/20260901000000_create_tasks.up.fizz:3-4`, `migrations/20261019000000_add_position_to_tasks.up.fizz:1`
- `ParseDueDate` convierte el valor del formulario en una fecha: quita los espacios, devuelve `nil, nil` si queda vacío y `ErrInvalidDueDate` si no es una fecha `AAAA-MM-DD` real — `models/task.go:44-55`
- `DueDateValue()` devuelve la fecha como `AAAA-MM-DD` (para `value` y `datetime`) y `DueDateLabel()` como `DD/MM/AAAA` (para mostrarla); los dos devuelven `""` si no hay fecha — `models/task.go:58-71`
- `ReorderTasks` guarda como posiciones `1..N` el orden de los ids recibidos — `models/task.go:76-107`

## Reglas e invariantes

- El título es obligatorio después de quitarle los espacios, y tiene como máximo `TitleMaxLength` (200) caracteres contados como runas — `models/task.go:113-118`
- `Validate` no comprueba la fecha límite: su formato se valida al parsearla en las acciones — `models/task.go:110-121`, `actions/tasks.go:57-63`
- Una fecha como `2026-02-30` no es válida, porque `time.Parse` rechaza días que no existen — `models/task_test.go:135`
- La fecha se guarda y se lee como medianoche UTC, y los helpers la formatean sin convertir de zona, así que no se desplaza de día. Lo comprueba `Test_Task_DueDate_Persists` — `models/task_test.go:154`
- `ReorderTasks` exige que lleguen todas las tareas exactamente una vez; si faltan, sobran, se repiten o no existen, devuelve `ErrReorderMismatch` sin cambiar nada — `models/task.go:82-97`
- La lista se ordena por `position asc, created_at asc, id asc` — `actions/tasks.go:21`

## Origen

- [2026-10-26 · Fecha límite opcional en las tareas](../episodica/2026-10-26-fecha-limite-tasks.md)
