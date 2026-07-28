// Package tui предоставляет терминальный интерфейс для GophKeeper.
package tui

import (
	"fmt"
	"strings"
	"time"

	models "github.com/PolyakovEvg/gophkeeper/internal/models"
)

// View - рендеринг текущего экрана
func (m model) View() string {
	if m.quitting {
		return ""
	}

	switch m.screen {
	case screenMain:
		return m.renderMain()
	case screenLogin:
		return m.renderLogin()
	case screenRegister:
		return m.renderRegister()
	case screenUnlock:
		return m.renderUnlock()
	case screenList:
		return m.renderList()
	case screenAddType:
		return m.renderAddType()
	case screenAddForm:
		return m.renderAddForm()
	case screenView:
		return m.renderView()
	case screenDeleteConfirm:
		return m.renderDeleteConfirm()
	default:
		return ""
	}
}

func (m model) renderMain() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("🔐 GophKeeper"))
	b.WriteString("\n\n")

	options := []string{"Войти", "Регистрация", "Выход"}
	for i, opt := range options {
		cursor := "  "
		if i == m.cursor {
			cursor = selectedStyle.Render("→ ")
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, opt))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓ - навигация, Enter - выбор, q - выход"))
	return boxStyle.Render(b.String())
}

func (m model) renderLogin() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("🔑 Вход"))
	b.WriteString("\n\n")

	for i, inp := range m.inputs {
		label := inp.label + ": "
		if i == m.inputIdx {
			label = selectedStyle.Render("→ ") + label
		} else {
			label = "  " + label
		}
		value := inp.value
		if inp.label == "Пароль" {
			value = strings.Repeat("•", len(value))
		}
		b.WriteString(fmt.Sprintf("%s%s\n", label, value))
		if i == m.inputIdx {
			b.WriteString(helpStyle.Render("   Введите значение и нажмите Enter\n"))
		}
	}

	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("❌ " + m.err))
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Tab/Enter - следующее поле, Esc - назад"))
	return boxStyle.Render(b.String())
}

func (m model) renderRegister() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("📝 Регистрация"))
	b.WriteString("\n\n")

	for i, inp := range m.inputs {
		label := inp.label + ": "
		if i == m.inputIdx {
			label = selectedStyle.Render("→ ") + label
		} else {
			label = "  " + label
		}
		value := inp.value
		if inp.label == "Пароль" {
			value = strings.Repeat("•", len(value))
		}
		b.WriteString(fmt.Sprintf("%s%s\n", label, value))
	}

	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("❌ " + m.err))
	}
	if m.success != "" {
		b.WriteString("\n")
		b.WriteString(successStyle.Render("✓ " + m.success))
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Tab/Enter - следующее поле, Esc - назад"))
	return boxStyle.Render(b.String())
}

func (m model) renderUnlock() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("🔓 Введите master-пароль"))
	b.WriteString("\n\n")

	for i, inp := range m.inputs {
		label := inp.label + ": "
		if i == m.inputIdx {
			label = selectedStyle.Render("→ ") + label
		} else {
			label = "  " + label
		}
		value := strings.Repeat("•", len(inp.value))
		b.WriteString(fmt.Sprintf("%s%s\n", label, value))
		if i == m.inputIdx {
			b.WriteString(helpStyle.Render("   Введите значение и нажмите Enter\n"))
		}
	}

	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("❌ " + m.err))
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Enter - подтвердить, Esc - выйти"))
	return boxStyle.Render(b.String())
}

func (m model) renderList() string {
	var b strings.Builder

	if m.login != "" {
		b.WriteString(headerStyle.Render("👤 " + m.login))
		b.WriteString("\n\n")
	}

	b.WriteString(titleStyle.Render("🔐 Ваши записи"))
	b.WriteString("\n\n")

	if len(m.items) == 0 {
		b.WriteString(menuStyle.Render("Нет записей. Нажмите 'a' для добавления."))
	} else {
		for i, it := range m.items {
			cursor := "  "
			if i == m.cursor {
				cursor = selectedStyle.Render("→ ")
			}
			icon := iconForType(it.Payload.Type)
			b.WriteString(fmt.Sprintf("%s%s %s [%s]\n", cursor, icon, it.Payload.Title, it.Payload.Type))
		}
	}

	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("❌ " + m.err))
	}
	if m.success != "" {
		b.WriteString("\n")
		b.WriteString(successStyle.Render("✓ " + m.success))
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("↑/↓ - навигация | Enter - просмотр | a - добавить | s - синхронизация | L - выход из аккаунта | q - выход"))
	return boxStyle.Render(b.String())
}

func (m model) renderAddType() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("➕ Добавить запись"))
	b.WriteString("\n\n")

	types := []struct {
		name string
	}{
		{"🔐 Логин/Пароль"},
		{"📄 Текст"},
		{"📁 Файл"},
		{"💳 Банковская карта"},
		{"⏱️ OTP секрет"},
	}

	for i, t := range types {
		cursor := "  "
		if i == m.cursor {
			cursor = selectedStyle.Render("→ ")
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, t.name))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓ - навигация, Enter - выбор, Esc - назад"))
	return boxStyle.Render(b.String())
}

func (m model) renderAddForm() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("➕ Добавить: " + string(m.itemType)))
	b.WriteString("\n\n")

	for i, inp := range m.inputs {
		label := inp.label + ": "
		if i == m.inputIdx {
			label = selectedStyle.Render("→ ") + label
		} else {
			label = "  " + label
		}
		value := inp.value
		if strings.Contains(strings.ToLower(inp.label), "пароль") ||
			strings.Contains(strings.ToLower(inp.label), "cvv") ||
			strings.Contains(strings.ToLower(inp.label), "секрет") {
			value = strings.Repeat("•", len(value))
		}
		b.WriteString(fmt.Sprintf("%s%s\n", label, value))
	}

	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("❌ " + m.err))
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Enter - следующее поле/сохранить, Esc - назад"))
	return boxStyle.Render(b.String())
}

func (m model) renderView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("📋 Просмотр записи"))
	b.WriteString("\n\n")

	if m.selected != nil {
		it := m.selected
		b.WriteString(fmt.Sprintf("ID: %s\n", it.ID.String()[:8]+"..."))
		b.WriteString(fmt.Sprintf("Тип: %s\n", it.Payload.Type))
		b.WriteString(fmt.Sprintf("Название: %s\n", it.Payload.Title))
		b.WriteString(fmt.Sprintf("Версия: %d\n", it.Version))
		b.WriteString(fmt.Sprintf("Обновлено: %s\n", it.UpdatedAt.Format(time.RFC3339)))

		b.WriteString("\n--- Данные ---\n")
		switch it.Payload.Type {
		case models.ItemCredentials:
			if it.Payload.Credentials != nil {
				b.WriteString(fmt.Sprintf("Логин: %s\n", it.Payload.Credentials.Login))
				b.WriteString(fmt.Sprintf("Пароль: %s\n", it.Payload.Credentials.Password))
			}
		case models.ItemText:
			if it.Payload.Text != nil {
				b.WriteString(fmt.Sprintf("Содержимое: %s\n", it.Payload.Text.Content))
			}
		case models.ItemBankCard:
			if it.Payload.BankCard != nil {
				b.WriteString(fmt.Sprintf("Номер: %s\n", it.Payload.BankCard.Number))
				b.WriteString(fmt.Sprintf("Держатель: %s\n", it.Payload.BankCard.Holder))
				b.WriteString(fmt.Sprintf("Срок: %s\n", it.Payload.BankCard.Expiry))
				b.WriteString(fmt.Sprintf("Банк: %s\n", it.Payload.BankCard.Bank))
			}
		case models.ItemOTP:
			if it.Payload.OTP != nil {
				b.WriteString(fmt.Sprintf("Issuer: %s\n", it.Payload.OTP.Issuer))
				b.WriteString(fmt.Sprintf("Account: %s\n", it.Payload.OTP.Account))
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("d - удалить | s - синхронизировать | Esc - назад"))
	return boxStyle.Render(b.String())
}

func (m model) renderDeleteConfirm() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("🗑️ Подтверждение удаления"))
	b.WriteString("\n\n")

	if m.selected != nil {
		b.WriteString(fmt.Sprintf("Удалить запись \"%s\"?\n\n", m.selected.Payload.Title))
	}

	options := []string{"Да, удалить", "Отмена"}
	for i, opt := range options {
		cursor := "  "
		if i == m.cursor {
			cursor = selectedStyle.Render("→ ")
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, opt))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓ - навигация, Enter - выбор"))
	return boxStyle.Render(b.String())
}

func iconForType(t models.ItemType) string {
	switch t {
	case models.ItemCredentials:
		return "🔐"
	case models.ItemText:
		return "📄"
	case models.ItemBinary:
		return "📁"
	case models.ItemBankCard:
		return "💳"
	case models.ItemOTP:
		return "⏱️"
	default:
		return "📦"
	}
}
