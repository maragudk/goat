package model

import (
	"encoding/json"
	"fmt"

	goomodel "maragu.dev/goo/model"
)

type ID = goomodel.ID

const (
	MySpeakerID = ID("s_26a91be1873f385bb0631ad868bf7c85")
)

type Time = goomodel.Time

type ModelType string

const (
	ModelTypeBrain     = ModelType("brain")
	ModelTypeLlamaCPP  = ModelType("llamacpp")
	ModelTypeOpenAI    = ModelType("openai")
	ModelTypeAnthropic = ModelType("anthropic")
)

type Model struct {
	ID      ID
	Created Time
	Updated Time
	Name    string
	Type    ModelType
	Config  string
}

func (m Model) URL() string {
	config := unmarshalConfig(m.Config)

	switch m.Type {
	case ModelTypeLlamaCPP:
		return fmt.Sprintf("http://%v/v1", config["address"])
	case ModelTypeOpenAI:
		return ""
	default:
		panic("unsupported model type")
	}
}

func (m Model) Token() string {
	config := unmarshalConfig(m.Config)

	switch m.Type {
	case ModelTypeLlamaCPP:
		return ""
	case ModelTypeOpenAI:
		return config["token"]
	default:
		panic("unsupported model type")
	}
}

type Speaker struct {
	ID      ID
	Created Time
	Updated Time
	ModelID ID `db:"modelID"`
	Name    string
	System  string
}

type Conversation struct {
	ID      ID
	Created Time
	Updated Time
	Topic   string
}

type Turn struct {
	ID             ID
	Created        Time
	Updated        Time
	ConversationID ID `db:"conversationID"`
	SpeakerID      ID `db:"speakerID"`
	Content        string
}

func unmarshalConfig(s string) map[string]string {
	config := map[string]string{}
	if err := json.Unmarshal([]byte(s), &config); err != nil {
		panic(err)
	}
	return config
}
