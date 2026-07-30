// Package tui предоставляет терминальный интерфейс для GophKeeper.
package tui

import (
	"context"

	"github.com/PolyakovEvg/gophkeeper/internal/client/app"
	models "github.com/PolyakovEvg/gophkeeper/internal/models"
)

// Состояния экрана
type screen int

const (
	screenMain screen = iota
	screenLogin
	screenRegister
	screenUnlock
	screenList
	screenAddType
	screenAddForm
	screenView
	screenDeleteConfirm
)

// inputField представляет поле ввода
type inputField struct {
	label string
	value string
}

// model - состояние TUI
type model struct {
	app      *app.App
	screen   screen
	cursor   int
	quitting bool
	err      string
	success  string
	login    string

	inputs   []inputField
	inputIdx int

	items    []models.LocalItem
	selected *models.LocalItem
	itemType models.ItemType

	ctx context.Context
}
