# Interfaz

Verificado: 2026-10-29 contra `43f3393`.

## Hechos

- Las plantillas son Plush: el layout está en `templates/application.plush.html`, la lista y el alta en `templates/tasks/index.plush.html` y la edición en `templates/tasks/edit.plush.html`
- Los assets llevan una versión de caché en la URL: `app.css?v=5` y `app.js?v=4` — `templates/application.plush.html:7`, `:30`
- La lista muestra «Vence el DD/MM/AAAA» dentro de `<time datetime="AAAA-MM-DD">`, solo en las tareas que tienen fecha — `templates/tasks/index.plush.html:45-47`
- En el alta, el campo de fecha (`#task-due-date`, `type="date"`, `name="due_date"`) tiene una etiqueta visualmente oculta, «Fecha límite», y está entre el título y el botón «Añadir» — `templates/tasks/index.plush.html:89-97`
- En la edición, el campo de fecha tiene la etiqueta visible «Fecha límite (opcional)» — `templates/tasks/edit.plush.html:26-33`
- `.task-due` se coloca en la segunda fila de la cuadrícula de cada tarea, bajo el título, en gris y sin saltos de línea — `public/assets/app.css:146-154`
- `input[type="date"]` tiene el mismo estilo que `input[type="text"]`; en la fila del alta no crece (`flex: 0 0 auto`) y en móvil baja a una segunda línea junto a «Añadir» — `public/assets/app.css:296-316`, captura `validaciones/fecha-limite-tasks/lista-movil.png`
- En la edición, el título tiene `autofocus` siempre, también cuando el error está en la fecha — `templates/tasks/edit.plush.html:17`

## Reglas e invariantes

- Los errores se muestran por campo: el título usa `errors.Get("title")` y `#title-error`, y la fecha usa `errors.Get("due_date")` y `#due-date-error`. Un error en un campo no marca el otro — `templates/tasks/index.plush.html:87`, `:96`, `:100-109`
- Un campo con error lleva `aria-invalid="true"` y un `aria-describedby` que apunta a su mensaje, y el mensaje lleva `role="alert"` — `templates/tasks/index.plush.html:96`, `:106`
- Las plantillas no desreferencian `DueDate`: siempre usan `DueDateValue()` y `DueDateLabel()` — `templates/tasks/index.plush.html:45-46`
- Los controles se localizan por su nombre accesible (por ejemplo «Editar tarea: <título>» o «Eliminar tarea: <título>»), y las pruebas con Playwright dependen de ello — `templates/tasks/index.plush.html:52`, `:60`
- Si un cambio toca el CSS o el JS, hay que subir su `?v=` — `templates/application.plush.html:7`, `:30`. Es la convención que siguió este encargo, no una regla escrita en el repositorio.

## Origen

- [2026-10-26 · Fecha límite opcional en las tareas](../episodica/2026-10-26-fecha-limite-tasks.md)
