package backup

type ActionType string

const (
	PUSH   = "push"
	REMOVE = "remove"
)

type RemoteAction struct {
	Type ActionType
	File File
}
