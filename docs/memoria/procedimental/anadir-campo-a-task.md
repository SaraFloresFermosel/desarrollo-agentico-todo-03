# Añadir un campo opcional a Task

Cuándo usarlo: hay que añadir a las tareas un dato nuevo que no es obligatorio (una fecha, una nota, una prioridad…) y que se edita desde los formularios de alta y edición. Es el patrón que se siguió con `due_date`. Primero hay que planificarlo con `/feature`; esto describe el corte vertical que ese plan debería tener.

## Requisitos

- PostgreSQL levantado: `docker compose up -d db` (ver [Entorno y CI](../semantica/entorno-y-ci.md)).
- `soda` instalado (la CI usa `go install github.com/gobuffalo/pop/v6/soda@v6.4.1`).

## Pasos

1. **Migración nullable y reversible.** Crea `migrations/<AAAAMMDDhhmmss>_add_<campo>_to_tasks.up.fizz` con `add_column("tasks", "<campo>", "<tipo>", {"null": true})` y su `.down.fizz` con `drop_column("tasks", "<campo>")`. Sin valor por defecto: así las tareas existentes quedan a `NULL`. Modelo: `migrations/20261026000000_add_due_date_to_tasks.*.fizz`.
2. **Aplica la migración y comprueba que se puede revertir** — `soda migrate -e test`, después `soda migrate down -e test` y otra vez `soda migrate -e test`. Entre medias, mira el esquema con `docker compose exec -T db psql -U postgres -d todo_test -c '\d tasks'`: el campo nuevo debe aparecer sin `not null` en la columna `Nullable` y sin valor en `Default`.
3. **Campo en el modelo** (`models/task.go`): usa un puntero (`*time.Time`, `*string`…) para representar `NULL` sin añadir dependencias, y la etiqueta `form:"-"` para que `c.Bind` no lo rellene: `<Campo> *<tipo> \`json:"<campo>" db:"<campo>" form:"-"\``.
4. **Parseo explícito**: escribe en el modelo una función `Parse<Campo>(raw string) (*<tipo>, error)` que quite los espacios, devuelva `nil, nil` si queda vacío y un error propio (`var ErrInvalid<Campo> = errors.New(…)`) si el formato no es válido.
5. **Helpers para las plantillas** con receptor por valor, que devuelvan `""` si el campo es `nil`: uno con el valor del `<input>` y otro con el texto que se muestra. Plush nunca debe desreferenciar el puntero.
6. **Acciones** (`actions/tasks.go`):
   - en `TasksCreate`, lee el valor con `c.Param("<campo>")`; en `TasksUpdate`, añádelo al struct `input` con `form:"<campo>"`;
   - si `Parse<Campo>` falla, crea `validate.NewErrors()`, añade el error con `verrs.Add("<campo>", "<mensaje>")` y responde 422: en el alta con `renderIndexWithErrors`, en la edición renderizando `tasks/edit.plush.html`. No guardes nada;
   - pasa a la plantilla una variable string con el valor tal como llegó, para que el usuario lo vea si se equivocó; inicialízala en `TasksIndex` y `TasksEdit`.
7. **Plantillas**: añade el campo con su `<label for>`. Asígnale `aria-invalid` y `aria-describedby="<campo>-error"` solo cuando `len(errors.Get("<campo>")) > 0`, y pon el mensaje en un `<p id="<campo>-error" class="notice field-error" role="alert">`. No uses `errors.HasAny()`, porque marcaría como inválidos campos que están bien.
8. **CSS**: si cambias `public/assets/app.css`, sube su `?v=` en `templates/application.plush.html`. En móvil, comprueba que no aparece scroll horizontal.
9. **Tests**: en `models/task_test.go`, el parseo (vacío, espacios, válido e inválidos) y la persistencia con y sin valor. En `actions/tasks_test.go`, crear con y sin valor, crear con un valor inválido (422, no se crea nada y vuelve lo que escribió el usuario), mostrarlo en la lista, precargarlo en la edición, actualizarlo y borrarlo, y actualizar con un valor inválido (422 sin modificar nada).
10. **Valida** con `/validacion plans/<slug>.md`.

## Cómo saber que ha funcionado

- `soda migrate -e test` termina con `> add_<campo>_to_tasks` y `Successfully applied 1 migrations.`; después del `down`, la columna ya no existe.
- Los tests nuevos pasan: `go test -p 1 ./actions ./models -v -run 'Test_ActionSuite|Test_ModelSuite' -testify.m '<Campo>'`, y `Test_TasksUpdate_PersistsTitleOnly` sigue en PASS.
- Un `POST /tasks` con un valor inválido devuelve `422 Unprocessable Entity` y el número de tareas no cambia.

## Errores conocidos

| Síntoma | Causa | Solución |
| --- | --- | --- |
| En el navegador no se puede escribir un valor inválido en un `<input type="date">` | El propio navegador lo impide | Para probar el 422, cambia el `type` del campo a `text` con `eval` en Playwright, o prueba con `curl` |
| Después del 422, el campo de fecha aparece vacío en el navegador aunque el servidor devolvió `value="no-es-fecha"` | Un `type="date"` descarta los valores que no son válidos | Es lo esperado. Para comprobar que el valor se devuelve, mira el HTML de la respuesta |
| `gofmt -l .` lista archivos que no has tocado | Finales de línea CRLF en Windows | Comprueba el contenido del commit, como explica [Entorno y CI](../semantica/entorno-y-ci.md) |

## Origen

- [2026-10-26 · Fecha límite opcional en las tareas](../episodica/2026-10-26-fecha-limite-tasks.md)
