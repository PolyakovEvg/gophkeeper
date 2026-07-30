// Package tui предоставляет терминальный интерфейс для GophKeeper.
package tui

import (
	"strings"
	"unicode/utf8"

	models "github.com/PolyakovEvg/gophkeeper/internal/models"

	tea "github.com/charmbracelet/bubbletea"
)

// Init - инициализация модели
func (m model) Init() tea.Cmd { return nil }

// Update - обработка событий
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m.handleEscape()
		case "q":
			return m.handleQuit()
		case "up", "k":
			return m.handleUp()
		case "down", "j":
			return m.handleDown()
		case "enter", " ":
			return m.handleEnter()
		case "a":
			return m.handleAdd()
		case "s":
			return m.handleSync()
		case "d":
			return m.handleDelete()
		case "L":
			return m.handleLogout()
		case "backspace":
			return m.handleBackspace()
		default:
			return m.handleInput(msg.String())
		}
	}
	return m, nil
}

func (m model) handleEscape() (tea.Model, tea.Cmd) {
	if m.screen != screenMain && m.screen != screenList {
		switch m.screen {
		case screenLogin, screenRegister:
			m.screen = screenMain
		case screenUnlock:
			m.screen = screenMain
		case screenAddType, screenView, screenDeleteConfirm:
			m.screen = screenList
		case screenAddForm:
			m.screen = screenAddType
		}
		m.err = ""
		m.success = ""
		m.inputs = nil
		return m, nil
	}
	m.quitting = true
	return m, tea.Quit
}

func (m model) handleQuit() (tea.Model, tea.Cmd) {
	if m.screen == screenMain || m.screen == screenList {
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m model) handleUp() (tea.Model, tea.Cmd) {
	if m.screen == screenAddForm {
		if m.inputIdx > 0 {
			m.inputIdx--
		}
	} else if m.cursor > 0 {
		m.cursor--
	}
	return m, nil
}

func (m model) handleDown() (tea.Model, tea.Cmd) {
	if m.screen == screenAddForm {
		if m.inputIdx < len(m.inputs)-1 {
			m.inputIdx++
		}
	} else {
		switch m.screen {
		case screenMain:
			if m.cursor < 2 {
				m.cursor++
			}
		case screenList:
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case screenAddType:
			if m.cursor < 4 {
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m model) handleAdd() (tea.Model, tea.Cmd) {
	if m.screen == screenList {
		m.screen = screenAddType
		m.cursor = 0
	}
	return m, nil
}

func (m model) handleSync() (tea.Model, tea.Cmd) {
	if m.screen == screenList || m.screen == screenView {
		if err := m.app.Sync(m.ctx, false); err != nil {
			m.err = err.Error()
			return m, nil
		}
		m.success = "Синхронизация завершена"
		m.items, _ = m.app.ListItems()
	}
	return m, nil
}

func (m model) handleDelete() (tea.Model, tea.Cmd) {
	if m.screen == screenView && m.selected != nil {
		m.screen = screenDeleteConfirm
		m.cursor = 0
	}
	return m, nil
}

func (m model) handleLogout() (tea.Model, tea.Cmd) {
	if m.screen == screenList {
		_ = m.app.Logout()
		m.login = ""
		m.items = nil
		m.screen = screenMain
		m.cursor = 0
	}
	return m, nil
}

func (m model) handleBackspace() (tea.Model, tea.Cmd) {
	if m.screen == screenLogin || m.screen == screenRegister || m.screen == screenAddForm || m.screen == screenUnlock {
		if len(m.inputs) > 0 && m.inputIdx < len(m.inputs) {
			runes := []rune(m.inputs[m.inputIdx].value)
			if len(runes) > 0 {
				m.inputs[m.inputIdx].value = string(runes[:len(runes)-1])
			}
		}
	}
	return m, nil
}

func (m model) handleInput(char string) (tea.Model, tea.Cmd) {
	if m.screen == screenLogin || m.screen == screenRegister || m.screen == screenAddForm || m.screen == screenUnlock {
		if len(m.inputs) > 0 && m.inputIdx < len(m.inputs) {
			if strings.HasPrefix(char, "shift+") {
				char = strings.ToUpper(strings.TrimPrefix(char, "shift+"))
			}
			// Учитываем длину в рунах, а не в байтах, иначе многобайтовые символы
			// (например кириллица) отбрасываются как "не одна клавиша".
			if utf8.RuneCountInString(char) == 1 {
				m.inputs[m.inputIdx].value += char
			}
		}
	}
	return m, nil
}

func (m model) handleEnter() (tea.Model, tea.Cmd) {
	m.err = ""
	m.success = ""

	switch m.screen {
	case screenMain:
		return m.handleMainEnter()
	case screenLogin:
		return m.handleLoginEnter()
	case screenRegister:
		return m.handleRegisterEnter()
	case screenUnlock:
		return m.handleUnlockEnter()
	case screenList:
		return m.handleListEnter()
	case screenAddType:
		return m.handleAddTypeEnter()
	case screenAddForm:
		return m.handleAddFormEnter()
	case screenDeleteConfirm:
		return m.handleDeleteConfirmEnter()
	}
	return m, nil
}

func (m model) handleMainEnter() (tea.Model, tea.Cmd) {
	switch m.cursor {
	case 0:
		m.screen = screenLogin
		m.inputs = []inputField{{label: "Логин", value: ""}, {label: "Пароль", value: ""}}
		m.inputIdx = 0
	case 1:
		m.screen = screenRegister
		m.inputs = []inputField{{label: "Логин", value: ""}, {label: "Пароль", value: ""}}
		m.inputIdx = 0
	case 2:
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m model) handleLoginEnter() (tea.Model, tea.Cmd) {
	if m.inputIdx < len(m.inputs)-1 {
		m.inputIdx++
		return m, nil
	}
	if err := m.app.Login(m.ctx, m.inputs[0].value, m.inputs[1].value); err != nil {
		m.err = err.Error()
		return m, nil
	}
	m.login = m.inputs[0].value
	m.success = "Успешный вход!"
	m.screen = screenList
	m.items, _ = m.app.ListItems()
	m.inputs = nil
	return m, nil
}

func (m model) handleRegisterEnter() (tea.Model, tea.Cmd) {
	if m.inputIdx < len(m.inputs)-1 {
		m.inputIdx++
		return m, nil
	}
	if err := m.app.Register(m.ctx, m.inputs[0].value, m.inputs[1].value); err != nil {
		m.err = err.Error()
		return m, nil
	}
	m.success = "Регистрация успешна! Теперь войдите."
	m.screen = screenLogin
	m.inputs = []inputField{{label: "Логин", value: m.inputs[0].value}, {label: "Пароль", value: ""}}
	m.inputIdx = 0
	return m, nil
}

func (m model) handleUnlockEnter() (tea.Model, tea.Cmd) {
	if len(m.inputs) == 0 || m.inputs[0].value == "" {
		m.err = "Введите пароль"
		return m, nil
	}
	password := m.inputs[0].value
	if err := m.app.RestoreSessionWithPassword(password); err != nil {
		m.err = "Неверный пароль"
		m.inputs[0].value = ""
		return m, nil
	}
	m.login = m.app.GetLogin()
	m.success = "Сессия восстановлена!"
	m.screen = screenList
	m.items, _ = m.app.ListItems()
	m.inputs = nil
	return m, nil
}

func (m model) handleListEnter() (tea.Model, tea.Cmd) {
	if len(m.items) > 0 && m.cursor >= 0 && m.cursor < len(m.items) {
		m.selected = &m.items[m.cursor]
		m.screen = screenView
	}
	return m, nil
}

func (m model) handleAddTypeEnter() (tea.Model, tea.Cmd) {
	types := []models.ItemType{models.ItemCredentials, models.ItemText, models.ItemBinary, models.ItemBankCard, models.ItemOTP}
	m.itemType = types[m.cursor]
	m.screen = screenAddForm
	m.inputs = getInputsForType(m.itemType)
	m.inputIdx = 0
	return m, nil
}

func (m model) handleAddFormEnter() (tea.Model, tea.Cmd) {
	if m.inputIdx < len(m.inputs)-1 {
		m.inputIdx++
		return m, nil
	}
	payload := m.buildPayload()
	if payload == nil {
		m.err = "Ошибка создания записи"
		return m, nil
	}
	if _, err := m.app.AddItem(*payload); err != nil {
		m.err = err.Error()
		return m, nil
	}
	m.success = "Запись добавлена!"
	m.screen = screenList
	m.items, _ = m.app.ListItems()
	m.inputs = nil
	return m, nil
}

func (m model) handleDeleteConfirmEnter() (tea.Model, tea.Cmd) {
	if m.cursor == 0 && m.selected != nil {
		if err := m.app.DeleteItem(m.selected.ID); err != nil {
			m.err = err.Error()
			return m, nil
		}
		m.success = "Запись удалена"
		m.items, _ = m.app.ListItems()
	}
	m.screen = screenList
	m.selected = nil
	return m, nil
}

func getInputsForType(t models.ItemType) []inputField {
	switch t {
	case models.ItemCredentials:
		return []inputField{
			{label: "Название", value: ""},
			{label: "Логин", value: ""},
			{label: "Пароль", value: ""},
		}
	case models.ItemText:
		return []inputField{
			{label: "Название", value: ""},
			{label: "Текст", value: ""},
		}
	case models.ItemBinary:
		return []inputField{
			{label: "Название", value: ""},
			{label: "Путь к файлу", value: ""},
		}
	case models.ItemBankCard:
		return []inputField{
			{label: "Название", value: ""},
			{label: "Номер карты", value: ""},
			{label: "Держатель", value: ""},
			{label: "Срок (MM/YY)", value: ""},
			{label: "CVV", value: ""},
			{label: "Банк", value: ""},
		}
	case models.ItemOTP:
		return []inputField{
			{label: "Название", value: ""},
			{label: "Секрет (base32)", value: ""},
			{label: "Issuer", value: ""},
			{label: "Account", value: ""},
		}
	}
	return nil
}

func (m model) buildPayload() *models.ItemPayload {
	if len(m.inputs) == 0 {
		return nil
	}

	payload := &models.ItemPayload{
		Type:  m.itemType,
		Title: m.inputs[0].value,
	}

	switch m.itemType {
	case models.ItemCredentials:
		payload.Credentials = &models.CredentialsData{
			Login:    m.inputs[1].value,
			Password: m.inputs[2].value,
		}
	case models.ItemText:
		payload.Text = &models.TextData{
			Content: m.inputs[1].value,
		}
	case models.ItemBankCard:
		payload.BankCard = &models.BankCardData{
			Number: m.inputs[1].value,
			Holder: m.inputs[2].value,
			Expiry: m.inputs[3].value,
			CVV:    m.inputs[4].value,
			Bank:   m.inputs[5].value,
		}
	case models.ItemOTP:
		payload.OTP = &models.OTPData{
			Secret:  m.inputs[1].value,
			Issuer:  m.inputs[2].value,
			Account: m.inputs[3].value,
			Period:  30,
			Digits:  6,
		}
	}

	return payload
}
