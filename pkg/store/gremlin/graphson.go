package gremlin

import (
	"encoding/json"
)

type GenericValue struct {
	Type  string          `json:"@type"`
	Value json.RawMessage `json:"@value"`
}

type VertexPropertyValue struct {
	ID    GenericValue    `json:"id"`
	Label string          `json:"label"`
	Value json.RawMessage `json:"value"`
}

type VertexProperty struct {
	Type  string              `json:"@type"`
	Value VertexPropertyValue `json:"@value"`
}

type VertexValue struct {
	ID         json.RawMessage             `json:"id"`
	Label      string                      `json:"label"`
	Properties map[string][]VertexProperty `json:"properties"`
}

type EdgePropertyValue struct {
	Label string          `json:"key"`
	Value json.RawMessage `json:"value"`
}

type EdgeProperty struct {
	Type  string            `json:"@type"`
	Value EdgePropertyValue `json:"@value"`
}

type EdgeValue struct {
	ID         json.RawMessage         `json:"id"`
	Label      string                  `json:"label"`
	InVLabel   string                  `json:"inVLabel"`
	OutVLabel  string                  `json:"outVLabel"`
	InV        json.RawMessage         `json:"inV"`
	OutV       json.RawMessage         `json:"outV"`
	Properties map[string]EdgeProperty `json:"properties"`
}
