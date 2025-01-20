package app

import (
	"Planning_poker/app/consts"
	"Planning_poker/app/models"
	"Planning_poker/app/utils"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"time"
)

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var roomsMu sync.Mutex
var rooms = make(map[string]*models.Room)

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket Upgrade error:", err)
		return
	}
	defer conn.Close()

	var room *models.Room
	go keepAlive(conn, 30*time.Second)

	for {
		var msg models.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WebSocket closed: %v", err)
			} else {
				log.Printf("Error reading JSON: %v", err)
			}

			if room != nil {
				var uId string
				room.Mu.Lock()
				if userId, exists := room.Clients[conn]; exists {
					uId = userId
					delete(room.Participants, userId)
				}
				delete(room.Clients, conn)
				room.Mu.Unlock()

				if uId == room.Creator {
					room.Broadcast(leaveResponse(room.ID))
				}

				room.Broadcast(updateResponse(false, room))
			}
			break
		}

		switch msg.Type {
		case consts.ClientMessageTypeCreate:
			creatorUUID := uuid.New().String()
			room = createRoom(msg.RoomName, msg.Password, creatorUUID, msg.RoomMethod)
			room.Mu.Lock()
			room.Participants[creatorUUID] = models.Participant{Name: msg.UserSessionName, Vote: "0"}
			room.Clients[conn] = creatorUUID
			room.Mu.Unlock()

			response := models.ServerMessage{
				Type: consts.ServerMessageTypeRoomCreated,
				Room: models.ServerMessageRoom{
					ID: room.ID,
				},
				User: models.ServerMessageUser{
					ID:   creatorUUID,
					Name: msg.UserSessionName,
				},
			}
			conn.WriteJSON(response)

		case consts.ClientMessageTypeJoin:
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

			if msg.UserSessionUUID == "" {
				msg.UserSessionUUID = uuid.New().String()
			}

			room.Mu.Lock()
			if existingParticipant, exists := room.Participants[msg.UserSessionUUID]; exists {
				room.Participants[msg.UserSessionUUID] = models.Participant{
					Name: existingParticipant.Name,
					Vote: existingParticipant.Vote,
				}
			} else {
				if msg.UserSessionName != "" {
					room.Participants[msg.UserSessionUUID] = models.Participant{
						Name: msg.UserSessionName,
						Vote: "0",
					}
				}

			}

			room.Clients[conn] = msg.UserSessionUUID
			room.Mu.Unlock()

			isOwner := msg.UserSessionUUID == room.Creator
			response := models.ServerMessage{
				Type: consts.ServerMessageTypeJoinedRoom,
				User: models.ServerMessageUser{
					ID:        msg.UserSessionUUID,
					Name:      msg.UserSessionName,
					RoomOwner: isOwner,
				},
				Room: models.ServerMessageRoom{
					ID:     room.ID,
					Reveal: room.Reveal,
				},
			}

			conn.WriteJSON(response)

			room.Broadcast(updateResponse(false, room))

		case consts.ClientMessageTypeVote:
			if room == nil {
				conn.WriteJSON(map[string]string{"error": "Not in a room"})
				break
			}

			room.Mu.Lock()
			if participant, exists := room.Participants[msg.UserSessionUUID]; exists {
				if msg.Vote == "" {
					msg.Vote = "coffee"
				}
				participant.Vote = msg.Vote
				room.Participants[msg.UserSessionUUID] = participant
			}
			room.Mu.Unlock()
			room.Broadcast(updateResponse(false, room))

		case consts.ClientMessageTypeReveal:
			if room == nil || msg.UserSessionUUID != room.Creator {
				conn.WriteJSON(map[string]string{"error": "Only the room creator can reveal votes"})
				break
			}
			room.Mu.Lock()
			room.Reveal = true
			hId := room.CurrentVote + 1
			var participantList []models.Participant
			for _, participant := range room.Participants {
				participantList = append(participantList, participant)
			}

			elapsedTime := int64(time.Since(room.StartTime).Seconds())
			room.RoomHistory = append(room.RoomHistory, models.RoomHistory{
				ID:             hId,
				Task:           room.CurrentTask,
				Participants:   participantList,
				RoomMethod:     room.RoomMethod,
				Average:        utils.CalculateVotingAverage(room.Participants),
				AverageFib:     utils.CalculateFibonacciVotingAverage(room.Participants),
				AverageTshirts: utils.CalculateTshirtsVotingAverage(room.Participants),
				Time:           utils.GetElapsedTime(elapsedTime),
			})

			room.CurrentVote = hId
			room.Mu.Unlock()

			room.Broadcast(updateResponse(false, room))

		case consts.ClientMessageTypeReset:
			if room == nil || msg.UserSessionUUID != room.Creator {
				conn.WriteJSON(map[string]string{"error": "Only the room creator can reset"})
				break
			}

			room.Mu.Lock()

			for k := range room.Participants {
				room.Participants[k] = models.Participant{Name: room.Participants[k].Name, Vote: "0"}
			}
			room.Reveal = false
			room.CurrentTask = ""
			room.Mu.Unlock()

			room.Broadcast(updateResponse(true, room))

		case consts.ClientMessageTypeTask:
			if room == nil || msg.UserSessionUUID != room.Creator {
				conn.WriteJSON(map[string]string{"error": "Only the room creator can reset"})
				break
			}

			room.Mu.Lock()
			room.LastTask = room.CurrentTask
			room.StartTime = time.Now()
			room.CurrentTask = msg.TaskName
			room.Mu.Unlock()

			room.Broadcast(updateResponse(false, room))

		case "check":
			roomsMu.Lock()
			room = rooms[msg.RoomID]
			roomsMu.Unlock()

			if room == nil {
				conn.WriteJSON(map[string]string{"error": "Room not found"})
				break
			}

			hasPassword := room.Password != ""
			response := models.ServerMessage{
				Type:        "info",
				HasPassword: hasPassword,
			}

			conn.WriteJSON(response)

		case consts.ClientMessageTypeSummary:
			roomsMu.Lock()
			room = rooms[msg.RoomID]
			roomsMu.Unlock()

			room.Broadcast(summaryResponse())
		}
	}
}

func keepAlive(conn *websocket.Conn, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println("Ping failed:", err)
				return
			}
		}
	}
}

func createRoom(name, password, creator, roomMethod string) *models.Room {
	room := &models.Room{
		ID:           uuid.New().String(),
		Name:         name,
		Password:     password,
		Participants: make(map[string]models.Participant),
		Clients:      make(map[*websocket.Conn]string),
		Reveal:       false,
		Creator:      creator,
		RoomMethod:   roomMethod,
	}

	roomsMu.Lock()
	rooms[room.ID] = room
	roomsMu.Unlock()
	return room
}

func GetExistsRoom(roomUUID string) (bool, *models.Room) {
	roomsMu.Lock()
	room, exists := rooms[roomUUID]
	roomsMu.Unlock()

	return exists, room
}

func CleanEmptyRooms() {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	for roomID, room := range rooms {
		if len(room.Participants) == 0 {
			fmt.Printf("Usuwam pokój: %s\n", roomID)
			delete(rooms, roomID)
		}
	}
}

func updateResponse(reset bool, room *models.Room) models.ServerMessage {
	return models.ServerMessage{
		Type: consts.ServerMessageTypeUpdate,
		Room: models.ServerMessageRoom{
			Participants:  room.Participants,
			Reveal:        room.Reveal,
			Reset:         reset,
			CurrentTask:   room.CurrentTask,
			LastTask:      room.LastTask,
			RoomOwnerUUID: room.Creator,
		},
		Voting: models.ServerMessageVoting{
			Fibonacci: utils.CalculateFibonacciVotingAverage(room.Participants),
			Average:   utils.CalculateVotingAverage(room.Participants),
			Tshirt:    utils.CalculateTshirtsVotingAverage(room.Participants),
		},
	}
}

func summaryResponse() models.ServerMessage {
	return models.ServerMessage{
		Type: consts.ServerMessageTypeSummary,
	}
}

func leaveResponse(roomID string) models.ServerMessage {
	return models.ServerMessage{
		Type: consts.ServerMessageTypeRedirect,
		Url:  "/summary/" + roomID,
	}
}
