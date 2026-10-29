# Episodio: fecha límite opcional en las tareas

2026-10-26 · feature · rama `adw/3-fecha-limite-tasks` · commit `be595cc` (validado el 2026-10-29 contra ese commit) · issue [#3](https://github.com/SaraFloresFermosel/desarrollo-agentico-todo-03/issues/3) · [plan](../../../plans/fecha-limite-tasks.md) · [validación](../../../validaciones/fecha-limite-tasks.md)

## Contexto

La issue #3 pedía añadir una fecha límite a las tareas, «simplemente un dia. Sin hora.». Antes del cambio, `Task` tenía `ID`, `Title`, `Completed`, `Position` y las marcas de tiempo; la tabla `tasks` no tenía ninguna fecha de vencimiento y `TasksUpdate` solo actualizaba el título.

## Qué se hizo

- **Persistencia:** columna `tasks.due_date` de tipo `date` y nullable, sin valor por defecto, con su `down` — `migrations/20261026000000_add_due_date_to_tasks.up.fizz:1`.
- **Modelo:** campo `DueDate *time.Time` con `form:"-"` (`models/task.go:34`), `ParseDueDate` (`models/task.go:44`) y los helpers `DueDateValue` / `DueDateLabel` (`models/task.go:58`, `:66`).
- **Acciones:** `TasksCreate` y `TasksUpdate` leen `due_date` y lo parsean a mano; una fecha no válida responde 422 (`actions/tasks.go:57-63`, `:120-128`). El render del índice con errores pasa a un helper, `renderIndexWithErrors` (`actions/tasks.go:34`).
- **Interfaz:** campo de fecha en el alta y en la edición, y «Vence el DD/MM/AAAA» en un `<time>` en la lista (`templates/tasks/index.plush.html:45-47`, `:89-97`; `templates/tasks/edit.plush.html:26-38`). Estilos en `public/assets/app.css:146` y `:297-316`; la caché del CSS pasa a `?v=5` (`templates/application.plush.html:7`).
- **Tests:** 7 tests de acciones (`actions/tasks_test.go:408-481`) y 2 de modelo (`models/task_test.go:135`, `:154`).

## Decisiones

| Decisión | Alternativa descartada | Motivo |
| --- | --- | --- |
| La fecha es **opcional** | Obligatoria | La issue no lo decía; con una fecha obligatoria las tareas existentes dejarían de ser válidas. Es una inferencia del plan, no un requisito. |
| Columna `date` | `timestamp` | La issue pide «sin hora», y con `date` no hay que tratar zonas horarias. |
| `*time.Time` para representar `NULL` | `nulls.Time` de `gobuffalo/nulls` | Así no se añade `gobuffalo/nulls` como dependencia directa. |
| `form:"-"` y parseo explícito en las acciones | Dejar que `c.Bind` rellene el campo | Una fecha mal escrita da un error controlado (422 y «La fecha límite no es válida.») en lugar de un fallo del binding. |
| Los helpers `DueDateValue` / `DueDateLabel` reciben `Task` por valor y devuelven `""` si no hay fecha | Desreferenciar `DueDate` en Plush | Evita desreferenciar un puntero `nil` dentro de la plantilla. |
| La variable de plantilla `dueDate` (string) lleva lo que se muestra en el campo | Usar `task.DueDate` | Permite devolver al formulario el valor tal como lo escribió el usuario cuando no es válido. |
| Un error por campo (`errors.Get("title")` y `errors.Get("due_date")`) | `errors.HasAny()` | Un error en la fecha no debe marcar el título como inválido ni moverle el foco. |
| Fuera de alcance: marcar las tareas vencidas, ordenar o filtrar por fecha, hora y recordatorios | — | La issue no lo pedía; la lista sigue ordenada por `position`. |

**Diferencias entre el plan y el código:** `.task-due` añade `grid-column`, `grid-row`, `margin-top` y `padding-bottom` (`public/assets/app.css:146-150`), que el plan no mencionaba, para colocar la fecha bajo el título. En la edición, el título sigue con `autofocus` sin condición (`templates/tasks/edit.plush.html:17`): ya era así y el plan no pedía cambiarlo.

## Problemas y cómo se resolvieron

- **`gofmt -l .` listaba casi todos los `.go`** (es un problema del entorno, no del cambio). El checkout usa `core.autocrlf=true` y gofmt marca los finales de línea CRLF. Se comprobó el contenido de `HEAD`, que es lo que ve la CI: 0 archivos con problemas.
- **Fallaban 6 tests de `todo/adws` en Windows** (es un fallo previo y ajeno al cambio). Fallan igual en `main` (`e5eaecf`). El `gh` falso de `adws/trigger_test.go:55` es un script bash, y en Windows se resuelve antes el `gh.exe` real.
- **El navegador no deja escribir una fecha no válida en un `type="date"`.** Para probar el 422 desde Playwright se cambió el `type` del campo a `text` con `eval`. Después, el campo aparece vacío porque el navegador descarta el valor no válido.
- **Hueco en los tests:** `Test_TasksCreate_InvalidDueDate` no comprueba que se conserve la fecha introducida (`value="no-es-fecha"`). Eso solo lo demuestra la prueba HTTP del informe.

## Resultado

**Validado con observaciones.** Se cumplen los 7 criterios, con tests, peticiones HTTP y pruebas en el navegador (incluido un vídeo). Pasan `go vet`, `go build`, las migraciones en los dos sentidos (`soda migrate` / `soda migrate down`) y los tests de `actions` y `models`. `go test ./...` solo falla en `todo/adws`, por el problema de Windows que ya existía.

En la misma rama, el commit siguiente (`43f3393`) añadió la skill `/validacion`.

## Lecciones

- Patrón para añadir un campo opcional a `Task`: migración nullable, puntero, `form:"-"`, parseo en la acción, helpers por valor y errores por campo → [Añadir un campo opcional a Task](../procedimental/anadir-campo-a-task.md).
- Qué guarda `Task` y sus reglas → [Modelo Task](../semantica/modelo-task.md).
- Contrato de las rutas, las respuestas 422 y las variables de plantilla → [Rutas y acciones](../semantica/rutas-y-acciones.md).
- Accesibilidad de los formularios, errores por campo y versión de caché de los assets → [Interfaz](../semantica/interfaz.md).
- `gofmt` con CRLF y los tests de `adws` en Windows → [Entorno y CI](../semantica/entorno-y-ci.md). El procedimiento ya está en la skill `/validacion`, así que no se ha copiado.
- Forzar `type="text"` para probar una fecha no válida en el navegador: solo episódica. Ya lo recoge el paso 5 de la sección 5 de `/validacion`.
