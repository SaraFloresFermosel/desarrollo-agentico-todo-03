package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// TitleMaxLength es el número máximo de caracteres permitido en el título.
const TitleMaxLength = 200

// ErrReorderMismatch indica que la lista de identificadores no contiene exactamente
// las tareas guardadas (faltan, sobran, se repiten o no existen).
var ErrReorderMismatch = errors.New("la lista de tareas no coincide con las guardadas")

// Task representa una tarea de la lista.
type Task struct {
	ID        uuid.UUID `json:"id" db:"id" form:"-"`
	Title     string    `json:"title" db:"title" form:"title"`
	Completed bool      `json:"completed" db:"completed" form:"-"`
	Position  int       `json:"position" db:"position" form:"-"`
	CreatedAt time.Time `json:"created_at" db:"created_at" form:"-"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" form:"-"`
}

// Tasks es una colección de Task.
type Tasks []Task

// ReorderTasks guarda como posiciones 1..N el orden indicado por ids.
// ids debe contener cada tarea exactamente una vez; si no, devuelve ErrReorderMismatch
// sin modificar nada.
func ReorderTasks(tx *pop.Connection, ids []uuid.UUID) error {
	tasks := Tasks{}
	if err := tx.All(&tasks); err != nil {
		return err
	}

	if len(ids) != len(tasks) {
		return ErrReorderMismatch
	}

	byID := make(map[uuid.UUID]*Task, len(tasks))
	for i := range tasks {
		byID[tasks[i].ID] = &tasks[i]
	}

	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if _, ok := byID[id]; !ok || seen[id] {
			return ErrReorderMismatch
		}
		seen[id] = true
	}

	for i, id := range ids {
		task := byID[id]
		task.Position = i + 1
		if err := tx.Update(task); err != nil {
			return err
		}
	}
	return nil
}

// Validate se ejecuta en cada llamada a los métodos "pop.Validate*".
func (t *Task) Validate(tx *pop.Connection) (*validate.Errors, error) {
	verrs := validate.NewErrors()

	title := strings.TrimSpace(t.Title)
	if title == "" {
		verrs.Add("title", "El título es obligatorio.")
	} else if len([]rune(title)) > TitleMaxLength {
		verrs.Add("title", fmt.Sprintf("El título no puede superar los %d caracteres.", TitleMaxLength))
	}

	return verrs, nil
}
