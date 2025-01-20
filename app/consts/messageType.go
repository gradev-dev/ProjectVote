package consts

const (
	ServerMessageTypeUpdate      = "update"
	ServerMessageTypeJoinedRoom  = "joinedRoom"
	ServerMessageTypeRoomCreated = "roomCreated"
	ServerMessageTypeRedirect    = "redirect"
	ServerMessageTypeSummary     = "summaryInfo"

	ClientMessageTypeCreate  = "create"
	ClientMessageTypeJoin    = "join"
	ClientMessageTypeVote    = "vote"
	ClientMessageTypeReveal  = "reveal"
	ClientMessageTypeReset   = "reset"
	ClientMessageTypeTask    = "task"
	ClientMessageTypeSummary = "summary"
)
