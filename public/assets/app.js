// Pide confirmación antes de enviar formularios marcados con data-confirm.
document.addEventListener("submit", function (event) {
  var form = event.target;
  if (form.dataset.confirm && !window.confirm(form.dataset.confirm)) {
    event.preventDefault();
  }

  if (!form.matches("[data-task-toggle]")) {
    return;
  }

  event.preventDefault();

  var task = form.closest(".task");
  var completed = task.classList.toggle("task-completed");
  var button = form.querySelector("button");
  var status = task.querySelector(".task-status");

  var label = completed ? "Reabrir" : "Completar";
  button.querySelector(".toggle-label").textContent = label;
  button.setAttribute("aria-pressed", String(completed));
  button.setAttribute("aria-label", label + " tarea: " + task.querySelector(".task-title").textContent);
  button.setAttribute("title", label + " tarea");

  if (completed && !status) {
    status = document.createElement("span");
    status.className = "task-status";
    status.textContent = "Completada";
    task.insertBefore(status, task.querySelector(".task-actions"));
  } else if (!completed && status) {
    status.remove();
  }

  fetch(form.action, { method: "POST", body: new FormData(form) });
});

// Reordenar tareas con arrastre (ratón/táctil con puntero) o con ↑ y ↓ en el asa.
(function () {
  var list = document.querySelector(".task-list");
  if (!list) {
    return;
  }

  var status = document.getElementById("reorder-status");
  var dragged = null;
  var dropTarget = null;
  var dropBefore = false;

  function clearDropIndicators() {
    list.querySelectorAll(".task-drop-before, .task-drop-after").forEach(function (item) {
      item.classList.remove("task-drop-before", "task-drop-after");
    });
  }

  function clearDragState() {
    clearDropIndicators();
    if (dragged) {
      dragged.classList.remove("task-dragging");
    }
    list.querySelectorAll('.task[draggable="true"]').forEach(function (item) {
      item.draggable = false;
    });
    dragged = null;
    dropTarget = null;
  }

  function announce(task) {
    var items = list.querySelectorAll(".task");
    var position = Array.prototype.indexOf.call(items, task) + 1;
    status.textContent = "Tarea movida a la posición " + position + " de " + items.length;
  }

  function saveOrder() {
    var body = new FormData();
    var token = document.querySelector('.new-task input[name="authenticity_token"]');
    if (token) {
      body.append("authenticity_token", token.value);
    }
    list.querySelectorAll(".task").forEach(function (item) {
      body.append("ids", item.dataset.taskId);
    });

    fetch(list.dataset.reorderUrl, { method: "POST", body: body })
      .then(function (response) {
        if (!response.ok) {
          window.location.reload();
        }
      })
      .catch(function () {
        window.location.reload();
      });
  }

  // Mueve task delante de ref (null = al final). Devuelve false si no cambia nada.
  function moveTask(task, ref) {
    if (ref === task || task.nextElementSibling === ref) {
      return false;
    }
    list.insertBefore(task, ref);
    return true;
  }

  list.addEventListener("pointerdown", function (event) {
    var handle = event.target.closest(".task-handle");
    if (!handle) {
      return;
    }
    handle.closest(".task").draggable = true;
  });

  document.addEventListener("pointerup", function () {
    list.querySelectorAll('.task[draggable="true"]').forEach(function (item) {
      if (item !== dragged) {
        item.draggable = false;
      }
    });
  });

  list.addEventListener("dragstart", function (event) {
    var task = event.target.closest(".task");
    if (!task || !task.draggable) {
      return;
    }
    dragged = task;
    task.classList.add("task-dragging");
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", task.dataset.taskId);
  });

  list.addEventListener("dragover", function (event) {
    if (!dragged) {
      return;
    }
    var target = event.target.closest(".task");
    if (!target || target === dragged) {
      return;
    }
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";

    var rect = target.getBoundingClientRect();
    dropBefore = event.clientY < rect.top + rect.height / 2;
    dropTarget = target;

    clearDropIndicators();
    target.classList.add(dropBefore ? "task-drop-before" : "task-drop-after");
  });

  list.addEventListener("drop", function (event) {
    if (!dragged || !dropTarget) {
      return;
    }
    event.preventDefault();

    var ref = dropBefore ? dropTarget : dropTarget.nextElementSibling;
    var moved = moveTask(dragged, ref);
    var task = dragged;
    clearDragState();

    if (moved) {
      announce(task);
      saveOrder();
    }
  });

  list.addEventListener("dragend", clearDragState);

  list.addEventListener("keydown", function (event) {
    var handle = event.target.closest(".task-handle");
    if (!handle || (event.key !== "ArrowUp" && event.key !== "ArrowDown")) {
      return;
    }
    event.preventDefault();

    var task = handle.closest(".task");
    var moved = false;
    if (event.key === "ArrowUp" && task.previousElementSibling) {
      moved = moveTask(task, task.previousElementSibling);
    } else if (event.key === "ArrowDown" && task.nextElementSibling) {
      moved = moveTask(task, task.nextElementSibling.nextElementSibling);
    }

    handle.focus();
    if (moved) {
      announce(task);
      saveOrder();
    }
  });
})();
