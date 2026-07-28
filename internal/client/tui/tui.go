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

	// Проверяем, есть ли сохранённая сессия (токен)
	// Если есть, пытаемся восстановить из keychain
	if application.HasSession() {
		// Пробуем автоматически восстановить сессию из keychain
		err := application.RestoreSession()
		if err == nil {
			// Сессия успешно восстановлена из keychain
			m.screen = screenList
			m.login = application.GetLogin()
			m.items, _ = application.ListItems()
		} else {
			// Пароль не найден в keychain, запрашиваем у пользователя
			m.screen = screenUnlock
			m.inputs = []inputField{{label: "Master-пароль", value: ""}}
			m.inputIdx = 0
		}
	}

	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
