package tui

import (
	"strconv"

	"github.com/volcanic001/vastago/internal/store"
)

type todoRow struct {
	todoIndex int
	heading   string
}

func todoVisualOrder(todos []store.Todo) []int {
	order := make([]int, 0, len(todos))
	for index, todo := range todos {
		if !todo.Completed() {
			order = append(order, index)
		}
	}
	for index, todo := range todos {
		if todo.Completed() {
			order = append(order, index)
		}
	}
	return order
}

func todoRows(todos []store.Todo) []todoRow {
	open, completed := 0, 0
	for _, todo := range todos {
		if todo.Completed() {
			completed++
		} else {
			open++
		}
	}
	rows := make([]todoRow, 0, len(todos)+2)
	if open > 0 {
		rows = append(rows, todoRow{todoIndex: -1, heading: "ABIERTOS · " + strconv.Itoa(open)})
	}
	completedHeadingAdded := false
	for _, index := range todoVisualOrder(todos) {
		if todos[index].Completed() && completed > 0 && !completedHeadingAdded {
			rows = append(rows, todoRow{todoIndex: -1, heading: "COMPLETADOS · " + strconv.Itoa(completed)})
			completedHeadingAdded = true
		}
		rows = append(rows, todoRow{todoIndex: index})
	}
	return rows
}

func (m Model) moveTodoSelection(offset int) Model {
	order := todoVisualOrder(m.db.Todos)
	for position, index := range order {
		if index == m.selectedTodo {
			m.selectedTodo = order[max(0, min(len(order)-1, position+offset))]
			return m
		}
	}
	if len(order) > 0 {
		m.selectedTodo = order[0]
	}
	return m
}
