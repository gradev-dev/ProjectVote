package models

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type Room struct {
	ID           string                     // Unikalny ID pokoju (UUID)
	Name         string                     // Nazwa pokoju
	Password     string                     // Hasło pokoju (opcjonalne)
	CurrentTask  string                     // Aktualny Task
	LastTask     string                     // Poprzedni Task
	Participants map[string]Participant     // Uczestnicy
	Clients      map[*websocket.Conn]string // Klienci WebSocket przypisani do uczestników
	Reveal       bool                       // Flaga odkrycia głosów
	Creator      string                     // Twórca pokoju (tylko on może odkrywać/resetować głosy)
	CurrentVote  int                        // Aktualna literacja głosowania
	RoomMethod   string                     // Metoda głosowania
	StartTime    time.Time                  // Czas rozpoczęcia
	RoomHistory  []RoomHistory              // Historia głosowania
	Mu           sync.Mutex                 // Mutex dla synchronizacji
}

type RoomHistory struct {
	ID             int           `json:"ID,omitempty"`           // ID historii
	Task           string        `json:"task,omitempty"`         // Aktualny Task
	Participants   []Participant `json:"participants,omitempty"` // Uczestnicy
	RoomMethod     string        `json:"roomMethod,omitempty"`
	Average        float64       `json:"average,omitempty"`
	AverageFib     int           `json:"fibonacci,omitempty"`
	AverageTshirts string        `json:"tshirt,omitempty"`
	JiraTaskUrl    string        `json:"jiraTaskUrl,omitempty"`
	Time           string        `json:"time"` // Czas trwania w sekundach
}

type Timer struct {
}

func (r *Room) Broadcast(message ServerMessage) {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	message.MessageId = uuid.NewString()
	for client, userID := range r.Clients {
		clientMessage := message
		clientMessage.User.RoomOwner = (userID == message.Room.RoomOwnerUUID)

		data, err := json.Marshal(clientMessage)
		if err != nil {
			log.Println("Error marshalling message:", err)
			continue
		}

		if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Error sending message to client: %v, error: %v", userID, err)
			client.Close()
			delete(r.Clients, client)
		} else {
			log.Printf("Message sent to client: %v", userID)
		}
	}
}
