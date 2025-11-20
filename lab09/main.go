package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/yeka/zip"
)

const (
	Digits  = "0123456789"
	Lower   = "abcdefghijklmnopqrstuvwxyz"
	Upper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Special = "!@#$%^&*()-_=+,.?/"
	All     = Digits + Lower + Upper + Special
)

func checkPassword(zipBytes []byte, password string) bool {
	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return false
	}

	for _, f := range zipReader.File {
		if f.FileInfo().IsDir() {
			continue
		}

		if f.IsEncrypted() {
			f.SetPassword(password)
			rc, err := f.Open()
			if err != nil {
				return false
			}

			buf := make([]byte, 1)
			_, err = rc.Read(buf)
			rc.Close()

			return err == nil
		}
	}
	return false
}

func worker(zipBytes []byte, jobs <-chan string, found *bool, wg *sync.WaitGroup, mu *sync.Mutex, resultChan chan string) {
	defer wg.Done()
	for pwd := range jobs {
		mu.Lock()
		if *found {
			mu.Unlock()
			return
		}
		mu.Unlock()

		if checkPassword(zipBytes, pwd) {
			mu.Lock()
			*found = true
			mu.Unlock()
			resultChan <- pwd
			return
		}
	}
}

func generatePasswords(charset string, length int, jobs chan<- string, found *bool, mu *sync.Mutex) {
	var generate func(current string)
	generate = func(current string) {
		mu.Lock()
		if *found {
			mu.Unlock()
			return
		}
		mu.Unlock()

		if len(current) == length {
			jobs <- current
			return
		}

		for _, char := range charset {
			generate(current + string(char))
		}
	}
	generate("")
}

func main() {
	mode := flag.String("mode", "brute", "brute або dict")
	file := flag.String("file", "", "Шлях до zip файлу")
	dictFile := flag.String("dict", "", "Словник")
	chars := flag.String("chars", "digits", "digits, lower, mixed, all")
	length := flag.Int("len", 4, "Довжина пароля")
	workers := flag.Int("workers", 12, "Кількість потоків")

	flag.Parse()

	if *file == "" {
		log.Fatal("Вкажіть файл: -file test.zip")
	}

	zipData, err := os.ReadFile(*file)
	if err != nil {
		log.Fatalf("Не вдалося прочитати файл: %v", err)
	}

	var charset string
	switch *chars {
	case "digits":
		charset = Digits
	case "lower":
		charset = Lower
	case "mixed":
		charset = Lower + Digits
	case "all":
		charset = All
	default:
		charset = Digits
	}

	jobs := make(chan string, 1000)
	resultChan := make(chan string)
	var wg sync.WaitGroup
	var mu sync.Mutex
	found := false

	startTime := time.Now()

	fmt.Printf("Атака на %s (RAM mode) | Довжина: %d | Набір: %s\n", *file, *length, *chars)

	for w := 1; w <= *workers; w++ {
		wg.Add(1)
		go worker(zipData, jobs, &found, &wg, &mu, resultChan)
	}

	go func() {
		switch *mode {
		case "brute":
			generatePasswords(charset, *length, jobs, &found, &mu)
		case "dict":
			f, err := os.Open(*dictFile)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				jobs <- scanner.Text()
			}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	password, ok := <-resultChan
	duration := time.Since(startTime)

	fmt.Println("------------------------------------------------")
	if ok {
		fmt.Printf("[УСПІХ] ПАРОЛЬ ЗНАЙДЕНО: %s\n", password)
	} else {
		fmt.Println("[НЕВДАЧА] Пароль не знайдено.")
	}
	fmt.Printf("Час виконання: %s\n", duration)
	fmt.Println("------------------------------------------------")
}
