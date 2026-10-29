# Rutas y acciones

Verificado: 2026-10-29 contra `43f3393`.

## Hechos

| Método y ruta | Acción | Respuesta | Evidencia |
| --- | --- | --- | --- |
| `GET /` | `TasksIndex` | 200 con la lista y el formulario de alta | `actions/app.go:51`, `actions/tasks.go:17` |
| `POST /tasks` | `TasksCreate` | 303 a `/` con el aviso «Tarea creada.»; 422 con el índice si hay errores | `actions/app.go:52`, `actions/tasks.go:48-85` |
| `POST /tasks/reorder` | `TasksReorder` | 204; 422 si los ids no son válidos o no coinciden con las tareas | `actions/app.go:53`, `actions/tasks.go:164-191` |
| `GET /tasks/{task_id}/edit` | `TasksEdit` | 200; 404 si la tarea no existe | `actions/app.go:54`, `actions/tasks.go:88-100` |
| `PUT /tasks/{task_id}` | `TasksUpdate` | 303 a `/` con el aviso «Tarea actualizada.»; 422 con la edición si hay errores | `actions/app.go:55`, `actions/tasks.go:103-144` |
| `PUT /tasks/{task_id}/toggle` | `TasksToggle` | 303 a `/` | `actions/app.go:56`, `actions/tasks.go:147-161` |
| `DELETE /tasks/{task_id}` | `TasksDestroy` | 303 a `/` con el aviso «Tarea eliminada.» | `actions/app.go:57`, `actions/tasks.go:194-208` |
| `GET /healthz` | `HealthCheck` | 200 `{"status":"ok"}` | `actions/app.go:58` |

- Todas las rutas pasan por `csrf.New` y por una transacción de pop por petición (`c.Value("tx")`) — `actions/app.go:45`, `:49`
- Los formularios envían `PUT` y `DELETE` como `POST` con un campo oculto `_method` — `templates/tasks/edit.plush.html:8`, `templates/tasks/index.plush.html:30`, `:59`
- Las plantillas de alta y edición reciben `task`, `errors` y `dueDate`, que es el texto que se muestra en el campo de fecha — `actions/tasks.go:25-28`, `:96-98`

## Reglas e invariantes

- `TasksCreate` añade la tarea al final: `position = MAX(position) + 1` — `actions/tasks.go:66-72`
- `TasksCreate` quita los espacios del principio y del final del título antes de validarlo — `actions/tasks.go:55`
- `TasksUpdate` solo cambia el título y la fecha límite. Aunque la petición mande `completed`, `position`, `id` o `created_at`, no los modifica — `actions/tasks.go:111-118`; lo cubre `Test_TasksUpdate_PersistsTitleOnly` para `completed`, `id` y `created_at` (`actions/tasks_test.go:196`). `position` solo se comprobó con una petición HTTP en la validación del 2026-10-29
- Si `due_date` llega vacío o no llega, la tarea queda sin fecha — `actions/tasks.go:120-129`, `models/task.go:46-48`
- Si `due_date` no es una fecha válida, la respuesta es 422 con el error «La fecha límite no es válida.» en el campo `due_date`. No se crea ni se modifica nada, y el formulario vuelve con lo que escribió el usuario — `actions/tasks.go:59-62`, `:121-127`
- Al volver a mostrar el índice con errores se usa siempre `renderIndexWithErrors`, que recarga la lista ordenada — `actions/tasks.go:34-45`

## Origen

- [2026-10-26 · Fecha límite opcional en las tareas](../episodica/2026-10-26-fecha-limite-tasks.md)
