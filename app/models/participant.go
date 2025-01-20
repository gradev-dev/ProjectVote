package models

type Participant struct {
	Name string `json:"name"` // Nazwa użytkownika
	Vote string `json:"vote"` // Głos użytkownika
}
