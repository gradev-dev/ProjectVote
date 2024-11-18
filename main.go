package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Message struct {
	Type     string `json:"type"`     // Typ wiadomości (e.g., "create", "join", "vote", "reveal", "reset")
	RoomID   string `json:"roomId"`   // ID pokoju
	RoomName string `json:"roomName"` // Nazwa pokoju
	Name     string `json:"name"`     // Imię uczestnika
	Vote     int    `json:"vote"`     // Głos (wartość Fibonacci)
	Reveal   bool   `json:"reveal"`   // Czy odkryć głosy
	Password string `json:"password"` // Hasło pokoju
}

type Participant struct {
	Name string `json:"name"`
	Vote int    `json:"vote"`
}

type Room struct {
	ID           string                     // Unikalny ID pokoju (UUID)
	Name         string                     // Nazwa pokoju
	Password     string                     // Hasło pokoju (opcjonalne)
	Participants map[string]Participant     // Uczestnicy
	Clients      map[*websocket.Conn]string // Klienci WebSocket przypisani do uczestników
	Reveal       bool                       // Flaga odkrycia głosów
	Creator      string                     // Twórca pokoju (tylko on może odkrywać/resetować głosy)
	mu           sync.Mutex                 // Mutex dla synchronizacji
}

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var roomsMu sync.Mutex
var rooms = make(map[string]*Room) // Mapa pokoi

// Tworzy nowy pokój
func createRoom(name, password, creator string) *Room {
	room := &Room{
		ID:           uuid.New().String(),
		Name:         name,
		Password:     password,
		Participants: make(map[string]Participant),
		Clients:      make(map[*websocket.Conn]string),
		Reveal:       false,
		Creator:      creator,
	}

	roomsMu.Lock()        // Zablokuj dostęp do mapy
	rooms[room.ID] = room // Dodaj pokój do mapy
	roomsMu.Unlock()      // Odblokuj dostęp do mapy

	log.Printf("Room created: ID=%s, Name=%s, Creator=%s", room.ID, room.Name, room.Creator)
	return room
}

// Broadcast do wszystkich w pokoju
func (r *Room) Broadcast(message interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := json.Marshal(message)
	if err != nil {
		log.Println("Error marshalling message:", err)
		return
	}

	for client := range r.Clients {
		err := client.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Printf("Error sending message to client: %v", err)
			client.Close()
			delete(r.Clients, client)
		} else {
			log.Printf("Message sent to client: %v", r.Clients[client])
		}
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket Upgrade error:", err)
		return
	}
	defer conn.Close()

	var room *Room
	var userName string

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WebSocket closed: %v", err)
			} else {
				log.Printf("Error reading JSON: %v", err)
			}
			// Usuń klienta i zaktualizuj pokój
			if room != nil {
				room.mu.Lock()
				if userName != "" {
					delete(room.Participants, userName)
					log.Printf("User %s removed from room %s", userName, room.ID)
				}
				delete(room.Clients, conn)
				room.mu.Unlock()

				// Aktualizuj stan pokoju
				broadcastMsg := map[string]interface{}{
					"type":         "update",
					"participants": room.Participants,
					"reveal":       room.Reveal,
				}
				room.Broadcast(broadcastMsg)
			}
			break
		}

		switch msg.Type {
		case "create":
			// Tworzenie pokoju
			room = createRoom(msg.RoomName, msg.Password, msg.Name)
			room.mu.Lock()
			room.Participants[msg.Name] = Participant{Name: msg.Name, Vote: 0}
			room.Clients[conn] = msg.Name
			room.mu.Unlock()

			userName = msg.Name // Zapisz nazwę użytkownika
			response := map[string]interface{}{
				"type":   "roomCreated",
				"roomId": room.ID,
			}
			conn.WriteJSON(response)

		case "join":
			// Dołączanie do pokoju
			roomsMu.Lock()
			room = rooms[msg.RoomID]
			roomsMu.Unlock()

			if room == nil {
				conn.WriteJSON(map[string]string{"error": "Room not found"})
				break
			}
			if room.Password != "" && room.Password != msg.Password {
				conn.WriteJSON(map[string]string{"error": "Invalid password"})
				break
			}

			isOwner := (msg.Name == room.Creator)

			room.mu.Lock()
			room.Participants[msg.Name] = Participant{Name: msg.Name, Vote: 0}
			room.Clients[conn] = msg.Name
			room.mu.Unlock()

			userName = msg.Name // Zapisz nazwę użytkownika
			conn.WriteJSON(map[string]interface{}{
				"type":     "joinedRoom",
				"roomId":   room.ID,
				"roomName": room.Name, // Dodaj nazwę pokoju
				"isOwner":  isOwner,
			})

			// Aktualizuj listę uczestników dla wszystkich
			broadcastMsg := map[string]interface{}{
				"type":         "update",
				"participants": room.Participants,
				"reveal":       room.Reveal,
			}
			room.Broadcast(broadcastMsg)

		case "vote":
			// Głosowanie
			if room == nil {
				conn.WriteJSON(map[string]string{"error": "Not in a room"})
				break
			}
			room.mu.Lock()
			if participant, exists := room.Participants[msg.Name]; exists {
				participant.Vote = msg.Vote
				room.Participants[msg.Name] = participant
			}
			room.mu.Unlock()

			// Aktualizacja dla wszystkich w pokoju
			broadcastMsg := map[string]interface{}{
				"type":         "update",
				"participants": room.Participants,
				"reveal":       room.Reveal,
			}
			room.Broadcast(broadcastMsg)

		case "reveal":
			// Odkrywanie głosów - tylko twórca
			if room == nil || msg.Name != room.Creator {
				conn.WriteJSON(map[string]string{"error": "Only the room creator can reveal votes"})
				break
			}
			room.mu.Lock()
			room.Reveal = true
			room.mu.Unlock()

			broadcastMsg := map[string]interface{}{
				"type":         "update",
				"participants": room.Participants,
				"reveal":       room.Reveal,
			}
			room.Broadcast(broadcastMsg)

		case "reset":
			// Resetowanie pokoju - tylko twórca
			if room == nil || msg.Name != room.Creator {
				conn.WriteJSON(map[string]string{"error": "Only the room creator can reset"})
				break
			}
			room.mu.Lock()
			for k := range room.Participants {
				room.Participants[k] = Participant{Name: room.Participants[k].Name, Vote: 0}
			}
			room.Reveal = false
			room.mu.Unlock()

			broadcastMsg := map[string]interface{}{
				"type":         "update",
				"participants": room.Participants,
				"reveal":       room.Reveal,
			}
			room.Broadcast(broadcastMsg)
		}
	}
}

func main() {
	// Tworzenie routera
	r := gin.Default()

	// Ładowanie szablonów HTML
	r.LoadHTMLGlob("templates/*")

	// Obsługa plików statycznych
	r.Static("/css", "./static/css")
	r.Static("/js", "./static/js")

	// Strona główna (tworzenie pokoju)
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "create.html", nil)
	})

	// Strona dołączania do pokoju
	r.GET("/join/:roomId", func(c *gin.Context) {
		roomID := c.Param("roomId")
		c.HTML(http.StatusOK, "join.html", gin.H{
			"RoomID": roomID,
		})
	})

	// Strona głosowania
	r.GET("/voting/:roomId", func(c *gin.Context) {
		roomID := c.Param("roomId")
		c.HTML(http.StatusOK, "voting.html", gin.H{
			"RoomID": roomID,
		})
	})

	// Obsługa WebSocket
	r.GET("/ws", func(c *gin.Context) {
		handleWebSocket(c.Writer, c.Request)
	})

	// Start serwera
	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
