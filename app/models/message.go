package models

type Message struct {
	Type            string `json:"type"`         // Typ wiadomości (e.g., "create", "join", "vote", "reveal", "reset")
	RoomID          string `json:"room_id"`      // ID pokoju
	RoomName        string `json:"room_name"`    // Nazwa pokoju
	UserSessionUUID string `json:"session_uuid"` // UUID uczestnika
	UserSessionName string `json:"user_name"`    // Nazwa uczestnika
	Vote            string `json:"vote"`         // Głos
	Reveal          bool   `json:"reveal"`       // Czy odkryć głosy
	Password        string `json:"password"`     // Hasło pokoju
	TaskName        string `json:"task_name"`    // Nazwa zadania
	RoomMethod      string `json:"room_method"`  // Metoda głosowania
}

type ErrorMessage struct {
	Message string `json:"error"`
}

type ServerMessage struct {
	MessageId   string              `json:"message_id"`
	Type        string              `json:"type"`
	Room        ServerMessageRoom   `json:"room"`
	User        ServerMessageUser   `json:"user"`
	Voting      ServerMessageVoting `json:"voting"`
	Url         string              `json:"url"`
	HasPassword bool                `json:"has_password"`
}

type ServerMessageRoom struct {
	ID            string                 `json:"id"`
	Participants  map[string]Participant `json:"participants"`
	Reveal        bool                   `json:"reveal"`
	Reset         bool                   `json:"reset"`
	CurrentTask   string                 `json:"current_task"`
	LastTask      string                 `json:"last_task"`
	RoomOwnerUUID string                 `json:"-"`
}

type ServerMessageUser struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	RoomOwner bool   `json:"room_owner"`
}

type ServerMessageVoting struct {
	Fibonacci int     `json:"fibonacci"`
	Average   float64 `json:"average"`
	Tshirt    string  `json:"tshirt"`
}
