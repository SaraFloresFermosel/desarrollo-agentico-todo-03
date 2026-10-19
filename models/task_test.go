package models

import (
	"strings"

	"github.com/gofrs/uuid"
)

func (ms *ModelSuite) Test_Task_Create_Valid() {
	task := &Task{Title: "Comprar café"}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	ms.NotZero(task.ID)
	ms.False(task.Completed)

	count, err := ms.DB.Count(&Task{})
	ms.NoError(err)
	ms.Equal(1, count)
}

func (ms *ModelSuite) Test_Task_Title_Required() {
	task := &Task{Title: ""}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.True(verrs.HasAny())
	ms.Contains(verrs.Get("title"), "El título es obligatorio.")

	count, err := ms.DB.Count(&Task{})
	ms.NoError(err)
	ms.Equal(0, count)
}

func (ms *ModelSuite) Test_Task_Title_OnlySpaces_Invalid() {
	task := &Task{Title: "   \t  "}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.True(verrs.HasAny())
	ms.Contains(verrs.Get("title"), "El título es obligatorio.")
}

func (ms *ModelSuite) Test_Task_Title_MaxLength() {
	task := &Task{Title: strings.Repeat("a", TitleMaxLength)}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())
}

func (ms *ModelSuite) Test_Task_Title_TooLong_Invalid() {
	task := &Task{Title: strings.Repeat("a", TitleMaxLength+1)}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.True(verrs.HasAny())
	ms.NotEmpty(verrs.Get("title"))
}

func (ms *ModelSuite) Test_Task_Toggle_Persists() {
	task := &Task{Title: "Revisar el correo"}
	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	task.Completed = true
	ms.NoError(ms.DB.Update(task))

	reloaded := &Task{}
	ms.NoError(ms.DB.Find(reloaded, task.ID))
	ms.True(reloaded.Completed)
}

func (ms *ModelSuite) createTasks(titles ...string) Tasks {
	tasks := Tasks{}
	for _, title := range titles {
		task := Task{Title: title}
		verrs, err := ms.DB.ValidateAndCreate(&task)
		ms.NoError(err)
		ms.False(verrs.HasAny())
		tasks = append(tasks, task)
	}
	return tasks
}

func (ms *ModelSuite) positions(tasks Tasks) map[uuid.UUID]int {
	result := map[uuid.UUID]int{}
	for _, t := range tasks {
		task := &Task{}
		ms.NoError(ms.DB.Find(task, t.ID))
		result[t.ID] = task.Position
	}
	return result
}

func (ms *ModelSuite) Test_Tasks_Reorder_PersistsNewOrder() {
	tasks := ms.createTasks("A", "B", "C")
	ids := []uuid.UUID{tasks[2].ID, tasks[0].ID, tasks[1].ID}

	ms.NoError(ReorderTasks(ms.DB, ids))

	for i, id := range ids {
		task := &Task{}
		ms.NoError(ms.DB.Find(task, id))
		ms.Equal(i+1, task.Position)
	}
}

func (ms *ModelSuite) Test_Tasks_Reorder_RejectsMismatchedSet() {
	tasks := ms.createTasks("A", "B", "C")
	before := ms.positions(tasks)

	err := ReorderTasks(ms.DB, []uuid.UUID{tasks[0].ID, tasks[1].ID})
	ms.ErrorIs(err, ErrReorderMismatch)

	err = ReorderTasks(ms.DB, []uuid.UUID{tasks[0].ID, tasks[1].ID, tasks[2].ID, uuid.Must(uuid.NewV4())})
	ms.ErrorIs(err, ErrReorderMismatch)

	ms.Equal(before, ms.positions(tasks))
}

func (ms *ModelSuite) Test_Tasks_Reorder_RejectsDuplicates() {
	tasks := ms.createTasks("A", "B", "C")
	before := ms.positions(tasks)

	err := ReorderTasks(ms.DB, []uuid.UUID{tasks[0].ID, tasks[0].ID, tasks[2].ID})
	ms.ErrorIs(err, ErrReorderMismatch)

	ms.Equal(before, ms.positions(tasks))
}
