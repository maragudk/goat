package model

// Error is for errors in the business domain. See the constants below.
type Error string

const (
	ErrorConversationNotFound = Error("CONVERSATION_NOT_FOUND")
	ErrorModelNotFound        = Error("MODEL_NOT_FOUND")
	ErrorSpeakerNotFound      = Error("SPEAKER_NOT_FOUND")
)

func (e Error) Error() string {
	return string(e)
}
