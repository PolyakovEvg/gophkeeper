// Package tui содержит тесты терминального интерфейса.
package tui

import (
	"testing"

	"github.com/PolyakovEvg/gophkeeper/internal/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelViewAndUpdate(t *testing.T) {
	t.Parallel()

	m := model{screen: screenList, items: []models.LocalItem{{
		ID: uuid.New(),
		Payload: models.ItemPayload{
			Type:     models.ItemText,
			Title:    "note",
			Metadata: map[string]string{"k": "v"},
		},
	}}}

	view := m.View()
	assert.Contains(t, view, "Ваши записи")
	assert.Contains(t, view, "note")

	// j - вниз по списку (1 элемент → курсор не двигается)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = updated.(model)
	assert.Equal(t, 0, m.cursor)

	// k - вверх
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = updated.(model)
	assert.Equal(t, 0, m.cursor)

	// пустой список
	empty := model{screen: screenList}
	assert.Contains(t, empty.View(), "Нет записей")

	// q на экране списка → выход
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = updated.(model)
	assert.True(t, m.quitting)
	assert.NotNil(t, cmd)
	assert.Equal(t, "", m.View())

	_ = m.Init()
}

func TestModelNavigation(t *testing.T) {
	t.Parallel()

	m := model{screen: screenList, items: []models.LocalItem{
		{ID: uuid.New(), Payload: models.ItemPayload{Type: models.ItemText, Title: "a"}},
		{ID: uuid.New(), Payload: models.ItemPayload{Type: models.ItemText, Title: "b"}},
	}}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(model)
	assert.Equal(t, 1, m.cursor)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(model)
	assert.Equal(t, 0, m.cursor)
}

func TestMainScreenNavigation(t *testing.T) {
	t.Parallel()

	m := model{screen: screenMain}

	// вниз до последнего пункта (3 пункта: 0,1,2)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(model)
	assert.Equal(t, 1, m.cursor)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(model)
	assert.Equal(t, 2, m.cursor)

	// дальше двигаться нельзя
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(model)
	assert.Equal(t, 2, m.cursor)

	// вверх
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(model)
	assert.Equal(t, 1, m.cursor)
}

func TestMainEnterToLogin(t *testing.T) {
	t.Parallel()

	m := model{screen: screenMain, cursor: 0}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	assert.Equal(t, screenLogin, m.screen)
	assert.Len(t, m.inputs, 2)
	assert.Equal(t, 0, m.inputIdx)
}

func TestMainEnterToRegister(t *testing.T) {
	t.Parallel()

	m := model{screen: screenMain, cursor: 1}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	assert.Equal(t, screenRegister, m.screen)
}

func TestMainEnterQuit(t *testing.T) {
	t.Parallel()

	m := model{screen: screenMain, cursor: 2}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	assert.True(t, m.quitting)
	assert.NotNil(t, cmd)
}

func TestEscapeFromLogin(t *testing.T) {
	t.Parallel()

	m := model{screen: screenLogin, inputs: []inputField{{label: "Логин"}}}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	assert.Equal(t, screenMain, m.screen)
	assert.Nil(t, m.inputs)
}

func TestEscapeFromList(t *testing.T) {
	t.Parallel()

	m := model{screen: screenList}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	assert.True(t, m.quitting)
	assert.NotNil(t, cmd)
}

func TestQuitFromLogin(t *testing.T) {
	t.Parallel()

	// q на экране логина не выходит
	m := model{screen: screenLogin}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = updated.(model)
	assert.False(t, m.quitting)
}

func TestAddTypeNavigation(t *testing.T) {
	t.Parallel()

	m := model{screen: screenAddType}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(model)
	assert.Equal(t, 1, m.cursor)

	// вниз до максимума (4)
	for i := 0; i < 5; i++ {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updated.(model)
	}
	assert.Equal(t, 4, m.cursor)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	assert.Equal(t, screenList, m.screen)
}

func TestAddTypeEnter(t *testing.T) {
	t.Parallel()

	cases := []struct {
		cursor int
		want   models.ItemType
	}{
		{0, models.ItemCredentials},
		{1, models.ItemText},
		{2, models.ItemBinary},
		{3, models.ItemBankCard},
		{4, models.ItemOTP},
	}
	for _, tc := range cases {
		m := model{screen: screenAddType, cursor: tc.cursor}
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = updated.(model)
		assert.Equal(t, screenAddForm, m.screen)
		assert.Equal(t, tc.want, m.itemType)
		assert.NotEmpty(t, m.inputs)
	}
}

func TestAddFormNavigation(t *testing.T) {
	t.Parallel()

	m := model{
		screen:   screenAddForm,
		inputs:   getInputsForType(models.ItemCredentials),
		inputIdx: 0,
	}
	// enter - переход между полями
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	assert.Equal(t, 1, m.inputIdx)

	// up - назад
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(model)
	assert.Equal(t, 0, m.inputIdx)

	// esc - назад к выбору типа
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	assert.Equal(t, screenAddType, m.screen)
}

func TestInputTyping(t *testing.T) {
	t.Parallel()

	m := model{
		screen:   screenLogin,
		inputs:   []inputField{{label: "Логин", value: ""}},
		inputIdx: 0,
	}
	// "x" не привязан к действию → попадает в handleInput
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	m = updated.(model)
	assert.Equal(t, "x", m.inputs[0].value)

	// backspace
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = updated.(model)
	assert.Equal(t, "", m.inputs[0].value)
}

func TestViewScreenRender(t *testing.T) {
	t.Parallel()

	item := &models.LocalItem{
		ID:      uuid.New(),
		Version: 3,
		Payload: models.ItemPayload{
			Type:  models.ItemBankCard,
			Title: "My Card",
			BankCard: &models.BankCardData{
				Number: "4111111111111111", Holder: "IVAN", Expiry: "12/30", Bank: "Bank",
			},
		},
	}
	m := model{screen: screenView, selected: item}
	view := m.View()
	assert.Contains(t, view, "My Card")
	assert.Contains(t, view, "4111111111111111")
	assert.Contains(t, view, "IVAN")
}

func TestDeleteConfirmRender(t *testing.T) {
	t.Parallel()

	m := model{
		screen:   screenDeleteConfirm,
		selected: &models.LocalItem{Payload: models.ItemPayload{Title: "ToDelete"}},
		cursor:   0,
	}
	view := m.View()
	assert.Contains(t, view, "ToDelete")
	assert.Contains(t, view, "Да, удалить")
}

func TestIconForType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		typ  models.ItemType
		want string
	}{
		{models.ItemCredentials, "🔐"},
		{models.ItemText, "📄"},
		{models.ItemBinary, "📁"},
		{models.ItemBankCard, "💳"},
		{models.ItemOTP, "⏱️"},
		{models.ItemType("unknown"), "📦"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, iconForType(tc.typ))
	}
}

func TestBuildPayload(t *testing.T) {
	t.Parallel()

	// credentials
	m := model{itemType: models.ItemCredentials, inputs: []inputField{
		{label: "Название", value: "t"},
		{label: "Логин", value: "l"},
		{label: "Пароль", value: "p"},
	}}
	p := m.buildPayload()
	require.NotNil(t, p)
	assert.Equal(t, "t", p.Title)
	assert.Equal(t, "l", p.Credentials.Login)
	assert.Equal(t, "p", p.Credentials.Password)

	// text
	m = model{itemType: models.ItemText, inputs: []inputField{
		{label: "Название", value: "t"},
		{label: "Текст", value: "content"},
	}}
	p = m.buildPayload()
	require.NotNil(t, p)
	assert.Equal(t, "content", p.Text.Content)

	// bank card
	m = model{itemType: models.ItemBankCard, inputs: []inputField{
		{label: "Название", value: "t"},
		{label: "Номер", value: "123"},
		{label: "Держатель", value: "H"},
		{label: "Срок", value: "12/30"},
		{label: "CVV", value: "123"},
		{label: "Банк", value: "B"},
	}}
	p = m.buildPayload()
	require.NotNil(t, p)
	assert.Equal(t, "123", p.BankCard.Number)

	// otp
	m = model{itemType: models.ItemOTP, inputs: []inputField{
		{label: "Название", value: "t"},
		{label: "Секрет", value: "S"},
		{label: "Issuer", value: "I"},
		{label: "Account", value: "A"},
	}}
	p = m.buildPayload()
	require.NotNil(t, p)
	assert.Equal(t, "S", p.OTP.Secret)
	assert.Equal(t, 30, p.OTP.Period)

	// пустые inputs
	m = model{itemType: models.ItemText, inputs: nil}
	assert.Nil(t, m.buildPayload())
}

func TestGetInputsForType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		typ  models.ItemType
		want int
	}{
		{models.ItemCredentials, 3},
		{models.ItemText, 2},
		{models.ItemBinary, 2},
		{models.ItemBankCard, 6},
		{models.ItemOTP, 4},
	}
	for _, tc := range cases {
		assert.Len(t, getInputsForType(tc.typ), tc.want)
	}
}

func TestHandleEnterOnEmptyList(t *testing.T) {
	t.Parallel()

	m := model{screen: screenList, items: nil}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	assert.Equal(t, screenList, m.screen)
	assert.Nil(t, m.selected)
}

func TestHandleEnterOnListSelectsItem(t *testing.T) {
	t.Parallel()

	items := []models.LocalItem{
		{ID: uuid.New(), Payload: models.ItemPayload{Type: models.ItemText, Title: "x"}},
	}
	m := model{screen: screenList, items: items, cursor: 0}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	assert.Equal(t, screenView, m.screen)
	require.NotNil(t, m.selected)
	assert.Equal(t, "x", m.selected.Payload.Title)
}

func TestHandleDeleteFromView(t *testing.T) {
	t.Parallel()

	item := &models.LocalItem{ID: uuid.New(), Payload: models.ItemPayload{Title: "d"}}
	m := model{screen: screenView, selected: item}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m = updated.(model)
	assert.Equal(t, screenDeleteConfirm, m.screen)
}

func TestHandleAddFromList(t *testing.T) {
	t.Parallel()

	m := model{screen: screenList}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = updated.(model)
	assert.Equal(t, screenAddType, m.screen)
}

func TestHandleAddFromMainNoOp(t *testing.T) {
	t.Parallel()

	m := model{screen: screenMain}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = updated.(model)
	assert.Equal(t, screenMain, m.screen)
}

func TestHandleLogoutFromMainNoOp(t *testing.T) {
	t.Parallel()

	m := model{screen: screenMain, login: "user"}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("L")})
	m = updated.(model)
	assert.Equal(t, screenMain, m.screen)
	assert.Equal(t, "user", m.login)
}

func TestHandleDeleteConfirmCancel(t *testing.T) {
	t.Parallel()

	m := model{screen: screenDeleteConfirm, cursor: 1, selected: &models.LocalItem{}}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	assert.Equal(t, screenList, m.screen)
	assert.Nil(t, m.selected)
}

func TestHandleSyncFromMainNoOp(t *testing.T) {
	t.Parallel()

	m := model{screen: screenMain}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m = updated.(model)
	assert.Equal(t, screenMain, m.screen)
}

func TestHandleBackspaceNoInputs(t *testing.T) {
	t.Parallel()

	m := model{screen: screenLogin, inputs: nil}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = updated.(model)
	assert.Nil(t, m.inputs)
}

func TestHandleInputOnListNoOp(t *testing.T) {
	t.Parallel()

	m := model{screen: screenList}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("z")})
	m = updated.(model)
	assert.Equal(t, screenList, m.screen)
}

func TestHandleInputShift(t *testing.T) {
	t.Parallel()

	m := model{
		screen:   screenLogin,
		inputs:   []inputField{{label: "Логин", value: ""}},
		inputIdx: 0,
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("shift+a")})
	m = updated.(model)
	assert.Equal(t, "A", m.inputs[0].value)
}

func TestViewScreenNilSelected(t *testing.T) {
	t.Parallel()

	m := model{screen: screenView, selected: nil}
	view := m.View()
	assert.Contains(t, view, "Просмотр записи")
}

func TestViewScreenAllTypes(t *testing.T) {
	t.Parallel()

	// credentials
	m := model{screen: screenView, selected: &models.LocalItem{
		Payload: models.ItemPayload{Type: models.ItemCredentials, Title: "c", Credentials: &models.CredentialsData{Login: "l", Password: "p"}},
	}}
	assert.Contains(t, m.View(), "l")

	// text
	m = model{screen: screenView, selected: &models.LocalItem{
		Payload: models.ItemPayload{Type: models.ItemText, Title: "t", Text: &models.TextData{Content: "txt"}},
	}}
	assert.Contains(t, m.View(), "txt")

	// otp
	m = model{screen: screenView, selected: &models.LocalItem{
		Payload: models.ItemPayload{Type: models.ItemOTP, Title: "o", OTP: &models.OTPData{Issuer: "I", Account: "A"}},
	}}
	assert.Contains(t, m.View(), "I")

	// nil sub-payload
	m = model{screen: screenView, selected: &models.LocalItem{
		Payload: models.ItemPayload{Type: models.ItemCredentials, Title: "c"},
	}}
	assert.Contains(t, m.View(), "Просмотр записи")
}

func TestQuittingView(t *testing.T) {
	t.Parallel()

	m := model{quitting: true}
	assert.Empty(t, m.View())
}

func TestUnknownScreenView(t *testing.T) {
	t.Parallel()

	m := model{screen: screen(999)}
	assert.Empty(t, m.View())
}

func TestUpdateNonKeyMsg(t *testing.T) {
	t.Parallel()

	m := model{screen: screenMain}
	updated, cmd := m.Update("not-a-key-msg")
	m = updated.(model)
	assert.Equal(t, screenMain, m.screen)
	assert.Nil(t, cmd)
}

func TestRenderMain(t *testing.T) {
	t.Parallel()

	m := model{screen: screenMain, cursor: 1}
	view := m.View()
	assert.Contains(t, view, "GophKeeper")
	assert.Contains(t, view, "Войти")
	assert.Contains(t, view, "Регистрация")
	assert.Contains(t, view, "Выход")
}

func TestRenderLogin(t *testing.T) {
	t.Parallel()

	m := model{
		screen: screenLogin,
		inputs: []inputField{
			{label: "Логин", value: "user"},
			{label: "Пароль", value: "secret"},
		},
		inputIdx: 1,
		err:      "bad creds",
	}
	view := m.View()
	assert.Contains(t, view, "Вход")
	assert.Contains(t, view, "user")
	assert.Contains(t, view, "••••••")
	assert.Contains(t, view, "bad creds")
}

func TestRenderRegister(t *testing.T) {
	t.Parallel()

	m := model{
		screen: screenRegister,
		inputs: []inputField{
			{label: "Логин", value: "u"},
			{label: "Пароль", value: "p"},
		},
		inputIdx: 0,
		success:  "done",
		err:      "",
	}
	view := m.View()
	assert.Contains(t, view, "Регистрация")
	assert.Contains(t, view, "done")
}

func TestRenderAddType(t *testing.T) {
	t.Parallel()

	m := model{screen: screenAddType, cursor: 2}
	view := m.View()
	assert.Contains(t, view, "Добавить запись")
	assert.Contains(t, view, "Текст")
	assert.Contains(t, view, "Файл")
}

func TestRenderAddForm(t *testing.T) {
	t.Parallel()

	m := model{
		screen:   screenAddForm,
		itemType: models.ItemCredentials,
		inputs: []inputField{
			{label: "Название", value: "t"},
			{label: "Логин", value: "l"},
			{label: "Пароль", value: "p"},
		},
		inputIdx: 2,
		err:      "some error",
	}
	view := m.View()
	assert.Contains(t, view, "Добавить")
	assert.Contains(t, view, "t")
	assert.Contains(t, view, "•")
	assert.Contains(t, view, "some error")
}

func TestRenderListWithLoginAndErrors(t *testing.T) {
	t.Parallel()

	m := model{
		screen:  screenList,
		login:   "alice",
		items:   []models.LocalItem{{Payload: models.ItemPayload{Type: models.ItemText, Title: "x"}}},
		err:     "sync failed",
		success: "ok",
	}
	view := m.View()
	assert.Contains(t, view, "alice")
	assert.Contains(t, view, "sync failed")
	assert.Contains(t, view, "ok")
}

func TestHandleLoginEnterAdvance(t *testing.T) {
	t.Parallel()

	m := model{
		screen:   screenLogin,
		inputs:   []inputField{{label: "Логин"}, {label: "Пароль"}},
		inputIdx: 0,
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	assert.Equal(t, 1, m.inputIdx)
	assert.Equal(t, screenLogin, m.screen)
}

func TestHandleRegisterEnterAdvance(t *testing.T) {
	t.Parallel()

	m := model{
		screen:   screenRegister,
		inputs:   []inputField{{label: "Логин"}, {label: "Пароль"}},
		inputIdx: 0,
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	assert.Equal(t, 1, m.inputIdx)
	assert.Equal(t, screenRegister, m.screen)
}

func TestHandleAddFormEnterAdvance(t *testing.T) {
	t.Parallel()

	m := model{
		screen:   screenAddForm,
		itemType: models.ItemText,
		inputs:   getInputsForType(models.ItemText),
		inputIdx: 0,
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	assert.Equal(t, 1, m.inputIdx)
	assert.Equal(t, screenAddForm, m.screen)
}

func TestHandleDownInAddFormAtMax(t *testing.T) {
	t.Parallel()

	m := model{
		screen:   screenAddForm,
		inputs:   getInputsForType(models.ItemText),
		inputIdx: 1,
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(model)
	assert.Equal(t, 1, m.inputIdx)
}

func TestHandleUpInAddFormAtMin(t *testing.T) {
	t.Parallel()

	m := model{
		screen:   screenAddForm,
		inputs:   getInputsForType(models.ItemText),
		inputIdx: 0,
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(model)
	assert.Equal(t, 0, m.inputIdx)
}

func TestHandleEscapeFromAddForm(t *testing.T) {
	t.Parallel()

	m := model{screen: screenAddForm}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	assert.Equal(t, screenAddType, m.screen)
}

func TestHandleEscapeFromView(t *testing.T) {
	t.Parallel()

	m := model{screen: screenView}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	assert.Equal(t, screenList, m.screen)
}

func TestHandleEscapeFromDeleteConfirm(t *testing.T) {
	t.Parallel()

	m := model{screen: screenDeleteConfirm}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	assert.Equal(t, screenList, m.screen)
}

func TestHandleQuitFromList(t *testing.T) {
	t.Parallel()

	m := model{screen: screenList}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = updated.(model)
	assert.True(t, m.quitting)
	assert.NotNil(t, cmd)
}

func TestHandleEnterUnknownScreen(t *testing.T) {
	t.Parallel()

	m := model{screen: screen(999)}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	assert.Equal(t, screen(999), m.screen)
}
