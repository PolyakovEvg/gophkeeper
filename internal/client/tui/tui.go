// Package tui предоставляет терминальный интерфейс для GophKeeper.
package tui

import (
	"context"

	"github.com/PolyakovEvg/gophkeeper/internal/client/app"

	tea "github.com/charmbracelet/bubbletea"
)

// Run запускает TUI.
func Run(application *app.App) error {
	m := model{
		app:    application,
		screen: screenMain,
		ctx:    context.Background(),
	}

	if application.HasSession() {
		err := application.RestoreSession()
		if err == nil {
			m.screen = screenList
			m.login = application.GetLogin()
			m.items, _ = application.ListItems()
		} else {
			m.screen = screenUnlock
			m.inputs = []inputField{{label: "Master-пароль", value: ""}}
			m.inputIdx = 0
		}
	}

	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
