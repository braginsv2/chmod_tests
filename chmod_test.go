package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"github.com/stretchr/testify/assert"
)

func runChmod(args ...string) error {
	cmd := exec.Command("chmod", args...)
	return cmd.Run()
}

func getPermissions(filename string) (string, error) {
	output, err := exec.Command("stat", "-c", "%A", filename).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func resetPermissions(filename string, mode string) error {
	return runChmod(mode, filename)
}

func TestChmod(t *testing.T) {
	testFile := "testfile"
	_, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("Не удалось создать тестовый файл: %v", err)
	}
	defer os.Remove(testFile)

	// Устанавливаем начальные права 755 перед каждым тестом
	if err := resetPermissions(testFile, "755"); err != nil {
		t.Fatalf("Не удалось установить начальные права: %v", err)
	}

	tests := []struct {
		name         string
		args         []string
		expectError  bool
		expectedMode string
	}{
		{"SetExecutable", []string{"755", testFile}, false, "rwxr-xr-x"},
		{"SetReadOnly", []string{"444", testFile}, false, "r--r--r--"},
		{"SetNoPermissions", []string{"000", testFile}, false, "---------"},
		{"AddExecute", []string{"+x", testFile}, false, "rwxr-xr-x"},
		{"RemoveRead", []string{"-r", testFile}, false, "-wx--x--x"},
		{"InvalidMode", []string{"999", testFile}, true, ""},
		{"NonExistentFile", []string{"755", "nonexistentfile"}, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Сброс прав перед каждым тестом
			if err := resetPermissions(testFile, "755"); err != nil {
				t.Fatalf("Не удалось сбросить права перед тестом %s: %v", tt.name, err)
			}

			// Исполняем команду chmod
			if err := runChmod(tt.args...); (err != nil) != tt.expectError {
				t.Errorf("runChmod() error = %v, expectError %v", err, tt.expectError)
				t.Logf("Параметры для chmod: %v", tt.args)
			} else if !tt.expectError {
				// Получаем текущие права доступа
				if mode, err := getPermissions(testFile); err == nil {
					if len(mode) > 0 && mode[0] == '-' {
						mode = mode[1:] // игнорируем символ типа файла 
					}
					assert.Equal(t, tt.expectedMode, mode, "Не совпадают права доступа для: "+tt.name)
				} else {
					t.Errorf("Ошибка при получении прав доступа: %v", err)
				}
			}
		})
	}
}
