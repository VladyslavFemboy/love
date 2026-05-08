// load_test.go
package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	url := "https://luvion.site"
	rps := 10
	duration := 10

	if len(os.Args) > 1 {
		url = os.Args[1]
	}
	if len(os.Args) > 2 {
		rps, _ = strconv.Atoi(os.Args[2])
	}
	if len(os.Args) > 3 {
		duration, _ = strconv.Atoi(os.Args[3])
	}

	interval := time.Second / time.Duration(rps)

	client := &http.Client{}

	var sent int
	var errors int

	fmt.Printf("🚀 Start\nURL: %s\nRPS: %d\nDuration: %ds\n\n", url, rps, duration)

	ticker := time.NewTicker(interval)
	stop := time.After(time.Duration(duration) * time.Second)

	// лог каждую секунду
	go func() {
		for {
			time.Sleep(1 * time.Second)
			fmt.Printf("📊 sent: %d | errors: %d\n", sent, errors)
		}
	}()

	for {
		select {
		case <-ticker.C:
			go func() {
				req, _ := http.NewRequest("GET", url, nil)
				req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")

				resp, err := client.Do(req)
				if err != nil {
					errors++
					return
				}
				resp.Body.Close()
				sent++
			}()

		case <-stop:
			fmt.Println("\n✅ Done")
			fmt.Printf("Total: %d | Errors: %d\n", sent, errors)
			return
		}
	}
}