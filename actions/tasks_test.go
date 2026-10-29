package actions

import (
	"net/url"
	"regexp"
	"strings"

	"todo/models"

	"github.com/gofrs/uuid"
)

func (as *ActionSuite) Test_TasksIndex_EmptyState() {
	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "No hay tareas todavía")
}

func (as *ActionSuite) Test_TasksIndex_ListsTasks() {
	as.LoadFixture("tres tareas")

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)

	body := res.Body.String()
	as.Contains(body, "Comprar material de oficina")
	as.Contains(body, "Llamar al gestor")
	as.Contains(body, "Publicar la oferta de empleo")
	as.Contains(body, "Completada")
}

func (as *ActionSuite) Test_TasksCreate_Valid() {
	res := as.HTML("/tasks").Post(&models.Task{Title: "  Nueva tarea de prueba  "})
	as.Equal(303, res.Code)
	as.Equal("/", res.Header().Get("Location"))

	task := &models.Task{}
	as.NoError(as.DB.First(task))
	as.Equal("Nueva tarea de prueba", task.Title)
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksCreate_EmptyTitle_Invalid() {
	res := as.HTML("/tasks").Post(&models.Task{Title: ""})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksCreate_OnlySpaces_Invalid() {
	res := as.HTML("/tasks").Post(&models.Task{Title: "     "})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksToggle_CompletesAndReopens() {
	task := &models.Task{Title: "Pendiente de completar"}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s/toggle", task.ID).Put(url.Values{})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.True(task.Completed)

	res = as.HTML("/tasks/%s/toggle", task.ID).Put(url.Values{})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksToggle_FormPostPersists() {
	as.LoadFixture("tres tareas")

	task := &models.Task{}
	as.NoError(as.DB.Where("completed = ?", false).First(task))

	page := as.HTML("/").Get().Body.String()
	token := regexp.MustCompile(`name="authenticity_token" value="([^"]+)"`).FindStringSubmatch(page)
	as.Require().Len(token, 2)

	res := as.HTML("/tasks/%s/toggle", task.ID).Post(url.Values{
		"_method":            {"PUT"},
		"authenticity_token": {token[1]},
	})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.True(task.Completed)
}

func (as *ActionSuite) Test_TasksToggle_GetDoesNotToggle() {
	task := &models.Task{Title: "GET no debe alternar"}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s/toggle", task.ID).Get()
	as.Equal(404, res.Code)

	as.NoError(as.DB.Reload(task))
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksToggle_NotFound() {
	id := uuid.Must(uuid.NewV4())
	res := as.HTML("/tasks/%s/toggle", id).Put(url.Values{})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksToggle_InvalidID() {
	res := as.HTML("/tasks/no-es-un-uuid/toggle").Put(url.Values{})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksDestroy() {
	task := &models.Task{Title: "Tarea que se elimina"}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s", task.ID).Delete()
	as.Equal(303, res.Code)

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksDestroy_NotFound() {
	id := uuid.Must(uuid.NewV4())
	res := as.HTML("/tasks/%s", id).Delete()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) taskForEditing(completed bool) *models.Task {
	task := &models.Task{Title: "Título original", Completed: completed}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())
	as.NoError(as.DB.Reload(task))
	return task
}

func (as *ActionSuite) Test_TasksIndex_EditLink() {
	task := as.taskForEditing(false)
	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "/tasks/"+task.ID.String()+"/edit")
	as.Contains(res.Body.String(), "Editar tarea: Título original")
}

func (as *ActionSuite) Test_TasksEdit_PrefillsTitle() {
	task := as.taskForEditing(true)
	res := as.HTML("/tasks/%s/edit", task.ID).Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), `value="Título original"`)
	as.Contains(res.Body.String(), `value="PUT"`)
	as.Contains(res.Body.String(), "Guardar cambios")
	as.Contains(res.Body.String(), "Cancelar")
	as.NoError(as.DB.Reload(task))
	as.True(task.Completed)
}

func (as *ActionSuite) Test_TasksEdit_NotFound() {
	res := as.HTML("/tasks/%s/edit", uuid.Must(uuid.NewV4())).Get()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksEdit_InvalidID() {
	res := as.HTML("/tasks/no-es-un-uuid/edit").Get()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksEdit_EscapesTitle() {
	task := &models.Task{Title: `<img src=x onerror="alert(1)">`}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())
	res := as.HTML("/tasks/%s/edit", task.ID).Get()
	as.Equal(200, res.Code)
	as.NotContains(res.Body.String(), `<img src=x`)
	as.Contains(res.Body.String(), "&lt;img")
}

func (as *ActionSuite) Test_TasksUpdate_PersistsTitleOnly() {
	task := as.taskForEditing(true)
	id, createdAt, updatedAt := task.ID, task.CreatedAt, task.UpdatedAt
	res := as.HTML("/tasks/%s", id).Put(url.Values{
		"title": {"  Título actualizado  "}, "completed": {"false"},
		"id": {uuid.Must(uuid.NewV4()).String()}, "created_at": {"2000-01-01T00:00:00Z"},
	})
	as.Equal(303, res.Code)
	as.Equal("/", res.Header().Get("Location"))
	as.NoError(as.DB.Reload(task))
	as.Equal("Título actualizado", task.Title)
	as.Equal(id, task.ID)
	as.True(task.Completed)
	as.True(task.CreatedAt.Equal(createdAt))
	as.False(task.UpdatedAt.Before(updatedAt))
	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(1, count)
}

func (as *ActionSuite) Test_TasksUpdate_PendingTaskRemainsPending() {
	task := as.taskForEditing(false)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"Nuevo título"}, "completed": {"true"}})
	as.Equal(303, res.Code)
	as.NoError(as.DB.Reload(task))
	as.Equal("Nuevo título", task.Title)
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksUpdate_EmptyTitle() {
	task := as.taskForEditing(false)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {""}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")
	as.Contains(res.Body.String(), `aria-describedby="title-error"`)
	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
}

func (as *ActionSuite) Test_TasksUpdate_OnlySpaces() {
	task := as.taskForEditing(true)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"   "}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")
	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
	as.True(task.Completed)
}

func (as *ActionSuite) Test_TasksUpdate_TitleTooLong() {
	task := as.taskForEditing(false)
	title := strings.Repeat("á", models.TitleMaxLength+1)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {title}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título no puede superar los 200 caracteres.")
	as.Contains(res.Body.String(), title)
	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
}

func (as *ActionSuite) Test_TasksUpdate_TitleAtLimit() {
	task := as.taskForEditing(false)
	title := strings.Repeat("á", models.TitleMaxLength)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {title}})
	as.Equal(303, res.Code)
	as.NoError(as.DB.Reload(task))
	as.Equal(title, task.Title)
}

func (as *ActionSuite) Test_TasksUpdate_NotFound() {
	res := as.HTML("/tasks/%s", uuid.Must(uuid.NewV4())).Put(url.Values{"title": {"Título nuevo"}})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksUpdate_InvalidID() {
	res := as.HTML("/tasks/no-es-un-uuid").Put(url.Values{"title": {"Título nuevo"}})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) tasksInPositionOrder(titles ...string) models.Tasks {
	for i, title := range titles {
		task := &models.Task{Title: title}
		verrs, err := as.DB.ValidateAndCreate(task)
		as.NoError(err)
		as.False(verrs.HasAny())
		task.Position = i + 1
		as.NoError(as.DB.Update(task))
	}
	tasks := models.Tasks{}
	as.NoError(as.DB.Order("position asc").All(&tasks))
	return tasks
}

func (as *ActionSuite) Test_TasksIndex_OrdersByPosition() {
	first := &models.Task{Title: "Primera creada"}
	verrs, err := as.DB.ValidateAndCreate(first)
	as.NoError(err)
	as.False(verrs.HasAny())
	second := &models.Task{Title: "Segunda creada"}
	verrs, err = as.DB.ValidateAndCreate(second)
	as.NoError(err)
	as.False(verrs.HasAny())

	second.Position = 1
	as.NoError(as.DB.Update(second))
	first.Position = 2
	as.NoError(as.DB.Update(first))

	body := as.HTML("/").Get().Body.String()
	as.Less(strings.Index(body, "Segunda creada"), strings.Index(body, "Primera creada"))
}

func (as *ActionSuite) Test_TasksIndex_PositionTiesFallBackToCreatedAt() {
	first := &models.Task{Title: "Creada antes"}
	verrs, err := as.DB.ValidateAndCreate(first)
	as.NoError(err)
	as.False(verrs.HasAny())
	second := &models.Task{Title: "Creada después"}
	verrs, err = as.DB.ValidateAndCreate(second)
	as.NoError(err)
	as.False(verrs.HasAny())

	first.Position = 1
	as.NoError(as.DB.Update(first))
	second.Position = 1
	as.NoError(as.DB.Update(second))

	body := as.HTML("/").Get().Body.String()
	as.Less(strings.Index(body, "Creada antes"), strings.Index(body, "Creada después"))
}

func (as *ActionSuite) Test_TasksCreate_AppendsAtEnd() {
	existing := &models.Task{Title: "Ya existente"}
	verrs, err := as.DB.ValidateAndCreate(existing)
	as.NoError(err)
	as.False(verrs.HasAny())
	existing.Position = 7
	as.NoError(as.DB.Update(existing))

	res := as.HTML("/tasks").Post(&models.Task{Title: "Nueva al final"})
	as.Equal(303, res.Code)

	created := &models.Task{}
	as.NoError(as.DB.Where("title = ?", "Nueva al final").First(created))
	as.Equal(8, created.Position)
}

func (as *ActionSuite) Test_TasksReorder_Persists() {
	tasks := as.tasksInPositionOrder("Tarea A", "Tarea B", "Tarea C")
	ids := []string{tasks[2].ID.String(), tasks[0].ID.String(), tasks[1].ID.String()}

	res := as.HTML("/tasks/reorder").Post(url.Values{"ids": ids})
	as.Equal(204, res.Code)

	body := as.HTML("/").Get().Body.String()
	as.Less(strings.Index(body, `data-task-id="`+ids[0]+`"`), strings.Index(body, `data-task-id="`+ids[1]+`"`))
	as.Less(strings.Index(body, `data-task-id="`+ids[1]+`"`), strings.Index(body, `data-task-id="`+ids[2]+`"`))
}

func (as *ActionSuite) Test_TasksReorder_MismatchReturns422AndKeepsOrder() {
	tasks := as.tasksInPositionOrder("Tarea A", "Tarea B", "Tarea C")

	res := as.HTML("/tasks/reorder").Post(url.Values{"ids": {tasks[2].ID.String(), tasks[0].ID.String()}})
	as.Equal(422, res.Code)

	after := models.Tasks{}
	as.NoError(as.DB.Order("position asc").All(&after))
	as.Equal(tasks, after)
}

func (as *ActionSuite) Test_TasksReorder_InvalidUUID() {
	tasks := as.tasksInPositionOrder("Tarea A", "Tarea B")

	res := as.HTML("/tasks/reorder").Post(url.Values{"ids": {"no-es-un-uuid", tasks[1].ID.String()}})
	as.Equal(422, res.Code)
}

func (as *ActionSuite) Test_TasksReorder_EmptyList() {
	res := as.HTML("/tasks/reorder").Post(url.Values{})
	as.Equal(204, res.Code)
}

func (as *ActionSuite) Test_TasksIndex_RendersHandle() {
	task := as.taskForEditing(false)

	body := as.HTML("/").Get().Body.String()
	as.Contains(body, `class="task-handle"`)
	as.Contains(body, `aria-label="Reordenar tarea: Título original"`)
	as.Contains(body, `data-task-id="`+task.ID.String()+`"`)
}

func (as *ActionSuite) Test_TasksIndex_HandleTitleEscaped() {
	task := &models.Task{Title: `<img src=x onerror="alert(1)">`}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	body := as.HTML("/").Get().Body.String()
	as.NotContains(body, `<img src=x`)
	as.Contains(body, "Reordenar tarea: &lt;img")
}

func (as *ActionSuite) taskWithDueDate(title, dueDate string) *models.Task {
	date, err := models.ParseDueDate(dueDate)
	as.NoError(err)
	task := &models.Task{Title: title, DueDate: date}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())
	return task
}

func (as *ActionSuite) Test_TasksCreate_WithDueDate() {
	res := as.HTML("/tasks").Post(url.Values{"title": {"Con fecha"}, "due_date": {"2026-11-02"}})
	as.Equal(303, res.Code)

	task := &models.Task{}
	as.NoError(as.DB.First(task))
	as.Equal("Con fecha", task.Title)
	as.Equal("2026-11-02", task.DueDateValue())
}

func (as *ActionSuite) Test_TasksCreate_WithoutDueDate() {
	res := as.HTML("/tasks").Post(url.Values{"title": {"Sin campo de fecha"}})
	as.Equal(303, res.Code)
	res = as.HTML("/tasks").Post(url.Values{"title": {"Con fecha vacía"}, "due_date": {""}})
	as.Equal(303, res.Code)

	tasks := models.Tasks{}
	as.NoError(as.DB.All(&tasks))
	as.Len(tasks, 2)
	for _, task := range tasks {
		as.Nil(task.DueDate)
	}
}

func (as *ActionSuite) Test_TasksCreate_InvalidDueDate() {
	res := as.HTML("/tasks").Post(url.Values{"title": {"Con fecha"}, "due_date": {"no-es-fecha"}})
	as.Equal(422, res.Code)
	body := res.Body.String()
	as.Contains(body, "La fecha límite no es válida.")
	as.Contains(body, `value="Con fecha"`)
	as.Contains(body, `aria-describedby="due-date-error"`)
	as.NotContains(body, `id="title-error"`)

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksIndex_ShowsDueDate() {
	as.taskWithDueDate("Con fecha", "2026-11-02")
	as.taskWithDueDate("Sin fecha", "")

	body := as.HTML("/").Get().Body.String()
	as.Equal(1, strings.Count(body, `datetime="2026-11-02"`))
	as.Equal(1, strings.Count(body, "02/11/2026"))
	as.Equal(1, strings.Count(body, "Vence el"))
}

func (as *ActionSuite) Test_TasksEdit_PrefillsDueDate() {
	task := as.taskWithDueDate("Con fecha", "2026-11-02")

	res := as.HTML("/tasks/%s/edit", task.ID).Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), `name="due_date"`)
	as.Contains(res.Body.String(), `value="2026-11-02"`)
}

func (as *ActionSuite) Test_TasksUpdate_SetsAndClearsDueDate() {
	task := as.taskForEditing(true)

	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"Título original"}, "due_date": {"2026-12-24"}})
	as.Equal(303, res.Code)
	as.NoError(as.DB.Reload(task))
	as.Equal("2026-12-24", task.DueDateValue())
	as.True(task.Completed)

	res = as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"Título original"}, "due_date": {""}})
	as.Equal(303, res.Code)
	as.NoError(as.DB.Reload(task))
	as.Nil(task.DueDate)
	as.True(task.Completed)
}

func (as *ActionSuite) Test_TasksUpdate_InvalidDueDate() {
	task := as.taskWithDueDate("Título original", "2026-11-02")

	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"Título nuevo"}, "due_date": {"2026-13-01"}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "La fecha límite no es válida.")
	as.Contains(res.Body.String(), `aria-describedby="due-date-error"`)

	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
	as.Equal("2026-11-02", task.DueDateValue())
}
