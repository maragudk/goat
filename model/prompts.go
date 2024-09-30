package model

import (
	"fmt"
)

const (
	globalPrompt = `You are an LLM assistant called %v participating in a converation with multiple speakers.
Each speaker has its message prefixed with a name you can use to refer to the speaker.
Do not prefix your answers with a name.

Example:

Avery: Hi everyone!
Hi Avery.

Messages may contain mentions prefixed with a @ sign. You don't have to do this yourself.

Example:

Someone saying hello to Avery with "Hello @avery".
Someone saying hi to John with "Hi @john".

`
)

func CreateGlobalPrompt(speakerName string) string {
	return fmt.Sprintf(globalPrompt, speakerName)
}
