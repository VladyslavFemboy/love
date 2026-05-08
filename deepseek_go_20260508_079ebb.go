package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Response struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	TestID   string `json:"test_id,omitempty"`
	Site     string `json:"site,omitempty"`
	Requests int    `json:"requests,omitempty"`
	Duration int    `json:"duration,omitempty"`
}

var (
	activeTests = make(map[string]*exec.Cmd)
	mu          sync.RWMutex
)

// Обработчик для /api/my.site/5000/3600/
func handler(w http.ResponseWriter, r *http.Request) {
	// Настройки CORS для Discord бота
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	
	path := strings.Trim(r.URL.Path, "/")
	
	// Остановка всех тестов
	if path == "stopalltest" {
		stopAllTests()
		json.NewEncoder(w).Encode(Response{
			Status:  "success",
			Message: "Все тесты остановлены",
		})
		return
	}
	
	// Парсим параметры: my.site/5000/3600
	parts := strings.Split(path, "/")
	if len(parts) != 3 {
		json.NewEncoder(w).Encode(Response{
			Status:  "error",
			Message: "Формат: /site/requests/seconds",
		})
		return
	}
	
	site := parts[0]
	requests, err := strconv.Atoi(parts[1])
	if err != nil {
		json.NewEncoder(w).Encode(Response{
			Status:  "error",
			Message: "Неверное количество запросов",
		})
		return
	}
	
	duration, err := strconv.Atoi(parts[2])
	if err != nil {
		json.NewEncoder(w).Encode(Response{
			Status:  "error",
			Message: "Неверное время",
		})
		return
	}
	
	// Добавляем https если нет
	if !strings.HasPrefix(site, "http") {
		site = "https://" + site
	}
	
	// Запускаем тест
	testID := runTest(site, requests, duration)
	
	json.NewEncoder(w).Encode(Response{
		Status:   "success",
		Message:  "Тест запущен",
		TestID:   testID,
		Site:     site,
		Requests: requests,
		Duration: duration,
	})
}

func runTest(site string, requests, duration int) string {
	testID := fmt.Sprintf("%d", time.Now().UnixNano())
	
	// Команда для запуска вашего chatgpt.go
	cmd := exec.Command("go", "run", "chatgpt.go", site, 
		fmt.Sprintf("%d", requests), 
		fmt.Sprintf("%d", duration))
	
	// Логи в файл
	logFile, _ := os.Create(fmt.Sprintf("test_%s.log", testID))
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	
	cmd.Start()
	
	mu.Lock()
	activeTests[testID] = cmd
	mu.Unlock()
	
	// Чистим после завершения
	go func() {
		cmd.Wait()
		mu.Lock()
		delete(activeTests, testID)
		mu.Unlock()
		logFile.Close()
		log.Printf("✅ Тест %s завершен", testID)
	}()
	
	log.Printf("🚀 Запущен тест %s: %s (RPS: %d, Время: %dс)", testID[:8], site, requests, duration)
	return testID
}

func stopAllTests() {
	mu.Lock()
	defer mu.Unlock()
	
	for id, cmd := range activeTests {
		if cmd.Process != nil {
			cmd.Process.Kill()
			log.Printf("🛑 Остановлен тест %s", id)
		}
	}
	activeTests = make(map[string]*exec.Cmd)
}

func main() {
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	
	http.HandleFunc("/", handler)
	http.HandleFunc("/api/", handler)
	
	log.Printf("✅ API сервер запущен на порту %s", port)
	log.Printf("📝 Пример запроса: GET /my.site/5000/3600/")
	log.Printf("🛑 Остановка всех: GET /stopalltest")
	
	log.Fatal(http.ListenAndServe(":"+port, nil))
}