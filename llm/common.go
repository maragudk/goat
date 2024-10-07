package llm

import (
	"github.com/sashabaranov/go-openai"
)

type Model string

func (m Model) String() string {
	return string(m)
}

type MessageRole string

const (
	MessageRoleSystem    = MessageRole(openai.ChatMessageRoleSystem)
	MessageRoleUser      = MessageRole(openai.ChatMessageRoleUser)
	MessageRoleAssistant = MessageRole(openai.ChatMessageRoleAssistant)
	MessageRoleFunction  = MessageRole(openai.ChatMessageRoleFunction)
	MessageRoleTool      = MessageRole(openai.ChatMessageRoleTool)
)

type Message struct {
	Content string
	Name    string
	Role    MessageRole
}
