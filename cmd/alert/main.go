package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/hpcloud/tail"
	_ "github.com/joho/godotenv/autoload"
)


func getInt64Env(name string) (int64, error) {
	valStr := os.Getenv(name)

	val, err := strconv.ParseInt(valStr, 10, 64)

	if err != nil {
		return 0, err
	}

	return val, nil
}

type config struct {
	LogFilePath string
	BotApiKey   string
	ChatId 		int64
}

func (c *config) Load() {
	c.LogFilePath = os.Getenv("LOG_FILE")
	if c.LogFilePath == "" {
		panic("WHERE'S THE GODDAMN LOG FILE")
	}
	c.BotApiKey = os.Getenv("TG_BOT_API_KEY")
	if c.BotApiKey == "" {
		panic("WHERE'S THE GODDAMN TG_BOT_API_KEY")
	}
	chatId, err := getInt64Env("CHAT_ID")
	if err != nil {
		panic("WHERE'S THE GODDAMN CHAT_ID")
	}
	c.ChatId = chatId
}

type LogEntry struct {
	Type  string `json:"type" validate:"required"`
	Level string `json:"level" validate:"required"`
}

func main() {
	cfg := config{}
	cfg.Load()

	// Please....
	threshold := 10
	window    := 10
	mu := sync.Mutex{}
	errs := map[string]int{}
	validator := validator.New()
	tgBot, err := NewTelegramNotifier(cfg.BotApiKey, cfg.ChatId)

	if err != nil {
		panic(err)
	}

	track := func(errType string) {
		mu.Lock()
		defer mu.Unlock()

		count, ok := errs[errType]

		if !ok {
			count = 1
		} else {
			count = count + 1
		}

		if count > threshold {
			if err := tgBot.SendAlert(errType, count, fmt.Sprintf("%d seconds", window)); err != nil {
				fmt.Println("SendAlert  function got goofed up:", err)
			}
		}

		errs[errType] = count

	}

	t, err := tail.TailFile(cfg.LogFilePath, tail.Config{
		Follow:   true,
		ReOpen:   true,
		Location: &tail.SeekInfo{Offset: 0, Whence: 2},
	})

	if err != nil {
		panic(err)
	}	

	go func() {
		
		// please....
		ticker := time.NewTicker(time.Second * time.Duration(window))

		defer ticker.Stop()

		for range ticker.C {
			mu.Lock()
			for k := range errs {
				delete(errs, k)			
			}
			mu.Unlock()
		}
	}()


	fmt.Println("We startin' baby...")
	for line := range t.Lines {
		var entry LogEntry
		if err := json.Unmarshal([]byte(line.Text), &entry); err != nil {
			continue
		}

		if err := validator.Struct(&entry); err != nil {
			continue
		}

		if entry.Level == "error" {
			track(entry.Type)
		}
	}


}
