package cli_test

import (
	"bytes"
	"encoding/base32"
	"log/slog"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PolyakovEvg/gophkeeper/internal/auth"
	"github.com/PolyakovEvg/gophkeeper/internal/client/cli"
	"github.com/PolyakovEvg/gophkeeper/internal/server/handlers"
	"github.com/PolyakovEvg/gophkeeper/internal/server/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLIVersionAndFlow(t *testing.T) {
	srvStore := storage.OpenTest(t)

	router := handlers.NewRouter(srvStore, auth.NewService("s"), slog.Default())
	server := httptest.NewServer(router.Routes())
	t.Cleanup(server.Close)

	dataDir := t.TempDir()
	info := cli.VersionInfo{Version: "1.2.3", BuildDate: "2026-01-01"}

	var out, errBuf bytes.Buffer
	code := cli.Run([]string{"version"}, &out, &errBuf, info)
	assert.Equal(t, 0, code)
	assert.Contains(t, out.String(), "1.2.3")

	out.Reset()
	code = cli.Run([]string{"help"}, &out, &errBuf, info)
	assert.Equal(t, 0, code)

	out.Reset()
	errBuf.Reset()
	code = cli.Run([]string{
		"register", "-server", server.URL, "-data", dataDir,
		"-login", "dave", "-password", "secret",
	}, &out, &errBuf, info)
	require.Equal(t, 0, code, errBuf.String())

	out.Reset()
	errBuf.Reset()
	code = cli.Run([]string{
		"add", "text", "-server", server.URL, "-data", dataDir,
		"-title", "hello", "-content", "world", "-master-password", "secret",
	}, &out, &errBuf, info)
	require.Equal(t, 0, code, errBuf.String())

	out.Reset()
	errBuf.Reset()
	code = cli.Run([]string{"list", "-server", server.URL, "-data", dataDir, "-master-password", "secret"}, &out, &errBuf, info)
	require.Equal(t, 0, code, errBuf.String())
	assert.Contains(t, out.String(), "hello")

	out.Reset()
	errBuf.Reset()
	code = cli.Run([]string{"sync", "-server", server.URL, "-data", dataDir, "-master-password", "secret"}, &out, &errBuf, info)
	require.Equal(t, 0, code, errBuf.String())

	out.Reset()
	errBuf.Reset()
	code = cli.Run([]string{"sync", "-binary", "-server", server.URL, "-data", dataDir, "-master-password", "secret"}, &out, &errBuf, info)
	require.Equal(t, 0, code, errBuf.String())
}

func TestCLICommands(t *testing.T) {
	srvStore := storage.OpenTest(t)

	server := httptest.NewServer(handlers.NewRouter(srvStore, auth.NewService("s"), slog.Default()).Routes())
	t.Cleanup(server.Close)

	dataDir := t.TempDir()
	info := cli.VersionInfo{Version: "test", BuildDate: "now"}
	base := []string{"-server", server.URL, "-data", dataDir}

	var out, errBuf bytes.Buffer
	run := func(args ...string) int {
		out.Reset()
		errBuf.Reset()
		return cli.Run(append(args[:1], append(base, args[1:]...)...), &out, &errBuf, info)
	}

	require.Equal(t, 0, run("register", "-login", "u1", "-password", "p1"), errBuf.String())
	require.Equal(t, 0, run("login", "-login", "u1", "-password", "p1"), errBuf.String())

	require.Equal(t, 0, run("add", "credentials", "-title", "gh", "-login", "l", "-password", "p", "-meta", "site=github.com"), errBuf.String())
	require.Equal(t, 0, run("add", "card", "-title", "visa", "-number", "4111", "-holder", "A", "-expiry", "12/30", "-cvv", "123"), errBuf.String())
	// ggignore - demo TOTP seed ("Hello!"), not a real secret
	otpSecret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte("Hello!"))
	require.Equal(t, 0, run("add", "otp", "-title", "otp", "-secret", otpSecret), errBuf.String())

	binFile := filepath.Join(t.TempDir(), "blob.bin")
	require.NoError(t, os.WriteFile(binFile, []byte{1, 2, 3}, 0o600))
	require.Equal(t, 0, run("add", "binary", "-title", "bin", "-file", binFile), errBuf.String())

	out.Reset()
	errBuf.Reset()
	code := cli.Run(append([]string{"list", "-master-password", "p1"}, base...), &out, &errBuf, info)
	require.Equal(t, 0, code, errBuf.String())
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	require.NotEmpty(t, lines)
	id := strings.Fields(lines[0])[0]

	out.Reset()
	errBuf.Reset()
	code = cli.Run(append([]string{"get", id, "-master-password", "p1"}, base...), &out, &errBuf, info)
	require.Equal(t, 0, code, errBuf.String())
	assert.Contains(t, out.String(), "title")

	out.Reset()
	errBuf.Reset()
	_ = cli.Run(append([]string{"list", "-master-password", "p1"}, base...), &out, &errBuf, info)
	var otpID string
	for _, line := range strings.Split(out.String(), "\n") {
		if strings.Contains(line, "\totp\t") {
			otpID = strings.Fields(line)[0]
			break
		}
	}
	require.NotEmpty(t, otpID)
	out.Reset()
	errBuf.Reset()
	code = cli.Run(append([]string{"otp", otpID, "-master-password", "p1"}, base...), &out, &errBuf, info)
	require.Equal(t, 0, code, errBuf.String())
	assert.Len(t, strings.TrimSpace(out.String()), 6)

	out.Reset()
	errBuf.Reset()
	code = cli.Run(append([]string{"delete", id, "-master-password", "p1"}, base...), &out, &errBuf, info)
	require.Equal(t, 0, code, errBuf.String())

	assert.Equal(t, 1, cli.Run([]string{"unknown"}, &out, &errBuf, info))
	assert.Equal(t, 1, cli.Run(nil, &out, &errBuf, info))
}

func TestCLIErrors(t *testing.T) {
	srvStore := storage.OpenTest(t)

	server := httptest.NewServer(handlers.NewRouter(srvStore, auth.NewService("s"), slog.Default()).Routes())
	t.Cleanup(server.Close)

	dataDir := t.TempDir()
	info := cli.VersionInfo{Version: "test", BuildDate: "now"}
	base := []string{"-server", server.URL, "-data", dataDir}

	var out, errBuf bytes.Buffer
	run := func(args ...string) int {
		out.Reset()
		errBuf.Reset()
		full := append([]string{args[0]}, base...)
		full = append(full, args[1:]...)
		return cli.Run(full, &out, &errBuf, info)
	}

	// register без логина/пароля
	assert.Equal(t, 1, run("register", "-login", "", "-password", ""))
	assert.Contains(t, errBuf.String(), "required")

	// login без логина/пароля
	assert.Equal(t, 1, run("login", "-login", "", "-password", ""))
	assert.Contains(t, errBuf.String(), "required")

	// add без аргументов
	assert.Equal(t, 1, run("add"))
	assert.Contains(t, errBuf.String(), "usage")

	// add без сессии (RestoreSession падает)
	assert.Equal(t, 1, run("add", "text", "-title", "x", "-content", "y"))
	assert.Contains(t, errBuf.String(), "not logged in")

	// list без сессии (без master-password)
	assert.Equal(t, 1, run("list", "-master-password", "wrongpassword"))
	assert.Contains(t, errBuf.String(), "not logged in")

	// get без аргументов
	assert.Equal(t, 1, run("get", "-master-password", "test"))
	assert.Contains(t, errBuf.String(), "usage")

	// get без сессии
	assert.Equal(t, 1, run("get", "00000000-0000-0000-0000-000000000000", "-master-password", "wrongpassword"))
	assert.Contains(t, errBuf.String(), "not logged in")

	// регистрируемся для проверки невалидного id и unknown type
	require.Equal(t, 0, run("register", "-login", "u2", "-password", "p2"), errBuf.String())

	// add неизвестного типа (теперь после регистрации)
	assert.Equal(t, 1, run("add", "bogus", "-title", "x"))
	assert.Contains(t, errBuf.String(), "unknown item type")

	// get с невалидным id (требуется master-password)
	assert.Equal(t, 1, run("get", "not-a-uuid", "-master-password", "p2"))
	assert.Contains(t, errBuf.String(), "invalid id")

	// delete без аргументов
	assert.Equal(t, 1, run("delete", "-master-password", "p2"))
	assert.Contains(t, errBuf.String(), "usage")

	// delete с невалидным id
	assert.Equal(t, 1, run("delete", "not-a-uuid", "-master-password", "p2"))
	assert.Contains(t, errBuf.String(), "invalid id")

	// otp без аргументов
	assert.Equal(t, 1, run("otp", "-master-password", "p2"))
	assert.Contains(t, errBuf.String(), "usage")

	// otp с невалидным id
	assert.Equal(t, 1, run("otp", "not-a-uuid", "-master-password", "p2"))
	assert.Contains(t, errBuf.String(), "invalid id")

	// add binary с несуществующим файлом
	assert.Equal(t, 1, run("add", "binary", "-title", "x", "-file", "/nonexistent/file"))
	assert.Contains(t, errBuf.String(), "no such file")

	// add без title
	assert.Equal(t, 1, run("add", "text", "-content", "y"))
	assert.Contains(t, errBuf.String(), "title is required")
}

func TestCLIFlagParsing(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	info := cli.VersionInfo{Version: "test", BuildDate: "now"}

	var out, errBuf bytes.Buffer
	code := cli.Run([]string{"version", "-server=http://localhost:8080", "-data=" + dataDir}, &out, &errBuf, info)
	assert.Equal(t, 0, code)
	assert.Contains(t, out.String(), "test")
}

func TestEnvOr(t *testing.T) {
	t.Setenv("GOPHKEEPER_TEST_ENV", "value")
	dataDir := t.TempDir()
	info := cli.VersionInfo{Version: "test", BuildDate: "now"}
	var out, errBuf bytes.Buffer
	code := cli.Run([]string{"version", "-data", dataDir}, &out, &errBuf, info)
	assert.Equal(t, 0, code)
}
