package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/s-piazzano/cosmoria/internal/adminauth"
	"github.com/s-piazzano/cosmoria/internal/auth"
	"github.com/s-piazzano/cosmoria/internal/collections"
	"github.com/s-piazzano/cosmoria/internal/rbac"
	"github.com/s-piazzano/cosmoria/internal/records"
	"github.com/s-piazzano/cosmoria/internal/tenant"
)

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ErrorObj       `json:"error,omitempty"`
}

type ErrorObj struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
	Handler     func(ctx context.Context, args json.RawMessage, svc *Services) (string, error)
}

type Services struct {
	AdminAuth   *adminauth.Service
	Auth        *auth.Service
	Tenant      *tenant.Service
	Collections *collections.Service
	RBAC        *rbac.Service
	Records     *records.Service
}

type CallToolParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type jsonProp struct {
	name     string
	typ      string
	desc     string
	required bool
}

func objSchema(props ...jsonProp) json.RawMessage {
	m := map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
	propsMap := m["properties"].(map[string]any)
	var req []string
	for _, p := range props {
		propsMap[p.name] = map[string]string{
			"type":        p.typ,
			"description": p.desc,
		}
		if p.required {
			req = append(req, p.name)
		}
	}
	if len(req) > 0 {
		m["required"] = req
	}
	data, _ := json.Marshal(m)
	return data
}

func success(text string) *Response {
	result, _ := json.Marshal(map[string]any{
		"content": []ContentBlock{{Type: "text", Text: text}},
	})
	return &Response{JSONRPC: "2.0", Result: result}
}

func failure(id *int, msg string) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &ErrorObj{Code: -32603, Message: fmt.Sprintf("Internal error: %s", msg)},
	}
}
