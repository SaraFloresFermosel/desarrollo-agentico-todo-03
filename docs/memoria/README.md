# Memoria del proyecto

Conocimiento del proyecto organizado en tres memorias, mantenido por la skill `/documentacion`.

- **Episódica** (`episodica/`): qué pasó en cada encargo y por qué se decidió así. No se reescribe.
- **Semántica** (`semantica/`): qué es cierto del sistema ahora. Se actualiza en su sitio.
- **Procedimental** (`procedimental/`): cómo se hacen las tareas repetibles, paso a paso.

## Episodios

- [2026-10-26 · Fecha límite opcional en las tareas](episodica/2026-10-26-fecha-limite-tasks.md) — issue #3: `due_date` opcional; por qué se eligió `date` y `*time.Time`; los problemas de gofmt y `adws` en Windows

## Semántica

- [Modelo Task](semantica/modelo-task.md) — campos, columnas, parseo y helpers de la fecha límite, validación del título y orden de la lista
- [Rutas y acciones](semantica/rutas-y-acciones.md) — tabla de rutas y respuestas, CSRF, qué puede cambiar cada acción y cuándo responde 422
- [Interfaz](semantica/interfaz.md) — plantillas Plush, errores por campo, accesibilidad, estilos de la fecha y versión de caché de los assets
- [Entorno y CI](semantica/entorno-y-ci.md) — PostgreSQL en Docker, pasos de la CI, tests con testify y particularidades de Windows

## Procedimientos

- [Añadir un campo opcional a Task](procedimental/anadir-campo-a-task.md) — al añadir a las tareas un dato no obligatorio que se edita en los formularios
