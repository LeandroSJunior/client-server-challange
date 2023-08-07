package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer res.Body.Close()

	rb, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
		return
	}

	var cotacao Cotacao
	err = json.Unmarshal(rb, &cotacao)
	if err != nil {
		log.Fatal(err)
		return
	}

	file, err := os.OpenFile("cotacao.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer file.Close()

	_, err = file.WriteString("Dólar: " + cotacao.USDBRL.Bid + "\n")
	if err != nil {
		log.Fatal(err)
		return
	}

}
