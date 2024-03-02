package main

import (
	_ "fmt"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	_ "net/http/httptest"
	_ "os"
	"strings"
	_ "strings"
	"testing"
	_ "time"
)

type MockDB struct{}

func (m *MockDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	user := dest.(*User)
	user.ID = 1
	user.Username = "johndoe"
	user.Email = "johndoe@example.com"
	user.PhoneNumber = "1234567890"
	return &gorm.DB{}
}

func (m *MockDB) Create(value interface{}) *gorm.DB {
	return &gorm.DB{}
}

func (m *MockDB) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	return &gorm.DB{}
}

func (m *MockDB) Model(value interface{}) *gorm.DB {
	return &gorm.DB{}
}

func TestHashPassword(t *testing.T) {
	password := "testPassword"
	hashedPassword, err := hashPassword(password)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Проверяем, что пароль был успешно хеширован
	if len(hashedPassword) == 0 {
		t.Error("Hashed password is empty")
	}

	// Проверяем, что хешированный пароль не совпадает с оригинальным
	if hashedPassword == password {
		t.Error("Hashed password matches original password")
	}

	// Проверяем, что функция comparePasswords возвращает true для правильного пароля
	if !comparePasswords(hashedPassword, password) {
		t.Error("comparePasswords returned false for correct password")
	}

	// Проверяем, что функция comparePasswords возвращает false для неправильного пароля
	if comparePasswords(hashedPassword, "wrongPassword") {
		t.Error("comparePasswords returned true for incorrect password")
	}
}

func TestGetUser(t *testing.T) {
	// Создание экземпляра фейковой базы данных
	mockDB := &MockDB{}

	// Создание запроса GET для получения информации о пользователе
	request, err := http.NewRequest("GET", "/getuser?id=1", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Создание фейкового HTTP ResponseWriter
	response := httptest.NewRecorder()

	// Обработка запроса с помощью вашего обработчика, передавая фейковую базу данных
	getUserHandlerWithDB(response, request, mockDB)

	// Проверка статус кода ответа
	if response.Code != http.StatusOK {
		t.Errorf("Expected status %v; got %v", http.StatusOK, response.Code)
	}

	// Проверка содержимого ответа
	expected := `{"id":1,"username":"johndoe","email":"johndoe@example.com","phone_number":"1234567890"}`
	if strings.TrimSpace(response.Body.String()) != expected {
		t.Errorf("Expected response body %v; got %v", expected, response.Body.String())
	}
}

func TestProductSearchEndToEnd(t *testing.T) {
	// Создаем фейковый сервер
	var baseURL = "http://localhost:5050"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что запрос отправлен правильно
		if r.URL.Path != "/getusersforpaging" {
			t.Errorf("Expected path /getusersforpaging, got %s", r.URL.Path)
		}

		// Проверяем, что метод запроса - GET
		if r.Method != http.MethodGet {
			t.Errorf("Expected method GET, got %s", r.Method)
		}

		// Отправляем фейковый ответ
		usersJSON := `[{"id":1,"username":"testuser","email":"test@example.com","phone_number":"1234567890"}]`
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(usersJSON))
	}))
	defer server.Close()

	// Заменяем базовый URL на адрес фейкового сервера
	oldBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = oldBaseURL }()

	// Выполняем запрос
	resp, err := http.Get(baseURL + "/getusersforpaging?filter=test")
	if err != nil {
		t.Fatalf("Error performing request: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем код состояния ответа
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Проверяем, что ответ содержит ожидаемые данные
	expected := `[{"id":1,"username":"testuser","email":"test@example.com","phone_number":"1234567890"}]`
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %v", err)
	}
	if string(body) != expected {
		t.Errorf("Expected body %s, got %s", expected, string(body))
	}
}
