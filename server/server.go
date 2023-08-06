package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type EconomiaResponseBody struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {
	http.HandleFunc("/cotacao", cotacaoHandler)
	http.ListenAndServe(":8080", nil)
}

func cotacaoHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	rb, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var parsed EconomiaResponseBody
	err = json.Unmarshal(rb, &parsed)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	db, err := sql.Open("sqlite3", "./bid.db")
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	err = insert(db, &parsed)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(parsed)
}

func insert(db *sql.DB, erb *EconomiaResponseBody) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()

	stmt, err := db.Prepare("insert into cotacao(id , code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, createDate) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx,
		erb.USDBRL.Code+erb.USDBRL.Codein,
		erb.USDBRL.Code,
		erb.USDBRL.Codein,
		erb.USDBRL.Name,
		erb.USDBRL.High,
		erb.USDBRL.Low,
		erb.USDBRL.VarBid,
		erb.USDBRL.PctChange,
		erb.USDBRL.Bid,
		erb.USDBRL.Ask,
		erb.USDBRL.Timestamp,
		erb.USDBRL.CreateDate,
	)
	if err != nil {
		return err
	}

	return nil
}
