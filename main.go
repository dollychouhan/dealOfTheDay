package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"sync"
	"time"
)

type Deal struct {
	Id      string    `json:"id"`
	Item    int       `json:"item"`
	Price   float64   `json:"price"`
	EndTime time.Time `json:"endTime"`
	Claimed int       `json:"claimed"`
	UserId  map[string]bool
}

type UpdateDealData struct {
	Items   int       `json:"items"`
	EndTime time.Time `json:"endTime"`
}

var (
	deals = make(map[string]*Deal)
	mu    sync.Mutex
)

func GenerateId() string {
	return time.Now().Format("123456789")
}

// ClaimDeal is used to claim the deal if exists
func ClaimDeal(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	vars := mux.Vars(r)
	id := vars["id"]

	userId := r.URL.Query().Get("userId")

	if deal, exists := deals[id]; exists {
		if time.Now().After(deal.EndTime) {
			http.Error(w, "Deal ended", http.StatusGone)
			return
		}

		if deal.Claimed >= deal.Item {
			http.Error(w, "Deal is out of range", http.StatusGone)
			return
		}

		if deal.UserId[userId] {
			http.Error(w, "User already claimed the deal", http.StatusGone)
			return
		}

		deal.Claimed++
		deal.UserId[userId] = true
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(deal); err != nil {
			log.Println("Error while encoding the data to claim the deal: ", err)
			return
		}
	} else {
		http.Error(w, "Deal not found", http.StatusNotFound)
		return
	}
}

// UpdateDeal is used to update the deals
func UpdateDeal(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	vars := mux.Vars(r)
	id := vars["id"]

	updateDealData := UpdateDealData{}

	if err := json.NewDecoder(r.Body).Decode(&updateDealData); err != nil {
		log.Println("Error while decoding the data to update the deal: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if deal, exists := deals[id]; exists {
		deal.Item = updateDealData.Items
		deal.EndTime = updateDealData.EndTime

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(deal); err != nil {
			log.Println("Error while encoding the data to update the deal: ", err)
			return
		}
	} else {
		http.Error(w, "Deal not found while updating the deal", http.StatusNotFound)
		return
	}
}

// EndDeal function is used to end the deal
func EndDeal(w http.ResponseWriter, r *http.Request) {

	mu.Lock()
	defer mu.Unlock()

	vars := mux.Vars(r)
	id := vars["id"]

	if deal, exists := deals[id]; exists {
		delete(deals, id)
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(deal); err != nil {
			log.Println("Error while encoding the data in End deal:", err)
			return
		}
	} else {
		http.Error(w, "Deal not found", http.StatusNotFound)
		return
	}
}

// CreateDeal function is used to create the deal
func CreateDeal(w http.ResponseWriter, r *http.Request) {

	mu.Lock()

	defer mu.Unlock()

	var deal Deal
	if err := json.NewDecoder(r.Body).Decode(&deal); err != nil {
		log.Println("Error while decoding the data: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	deal.Id = GenerateId()
	deal.UserId = make(map[string]bool)
	deals[deal.Id] = &deal

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(deal); err != nil {
		log.Println("Error while encoding the data: ", err)
		return
	}
}

func main() {

	router := mux.NewRouter()

	router.HandleFunc("/createDeal", CreateDeal).Methods("POST")
	router.HandleFunc("/endDeal/{id}", EndDeal).Methods("POST")
	router.HandleFunc("/updateDeal/{id}", UpdateDeal).Methods("PUT")
	router.HandleFunc("/claimDeal/{id}", ClaimDeal).Methods("POST")

	log.Println("Server is running at the port 8082")
	err := http.ListenAndServe(":8082", router)
	if err != nil {
		log.Println("Error while running the server")
		return
	}
}
