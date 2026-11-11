package model

type HelloInput struct {
	Name string `json:"name" jsonschema:"要打招呼的名字"`
}

type HelloOutput struct {
	Message string `json:"message" jsonschema:"问候语"`
}
