package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/s-piazzano/cosmoria/internal/adminauth"
	"github.com/s-piazzano/cosmoria/internal/audit"
	"github.com/s-piazzano/cosmoria/internal/auth"
	"github.com/s-piazzano/cosmoria/internal/collections"
	"github.com/s-piazzano/cosmoria/internal/core"
	"github.com/s-piazzano/cosmoria/internal/rbac"
	"github.com/s-piazzano/cosmoria/internal/records"
	"github.com/s-piazzano/cosmoria/internal/storage"
	"github.com/s-piazzano/cosmoria/internal/tenant"
)

type state int

const (
	stateNew state = iota
	stateInitialized
	stateReady
)

type Server struct {
	pool     *pgxpool.Pool
	cfg      *core.Config
	state    state
	services *Services
}

func NewServer(pool *pgxpool.Pool, cfg *core.Config) *Server {
	collSvc := collections.NewService(pool)
	s3Client := storage.NewS3Client(cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket, cfg.S3Region, cfg.S3UseSSL)
	return &Server{
		pool:  pool,
		cfg:   cfg,
		state: stateNew,
		services: &Services{
			AdminAuth:   adminauth.NewService(pool, cfg),
			Auth:        auth.NewService(pool, cfg),
			Tenant:      tenant.NewService(pool),
			Collections: collSvc,
			RBAC:        rbac.NewService(pool),
			Records:     records.NewService(pool, collSvc),
			Storage:     storage.NewService(pool, s3Client),
			Audit:       audit.NewService(pool),
		},
	}
}

func (s *Server) Run() error {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		resp := s.handleLine(line)
		if resp != nil {
			data, err := json.Marshal(resp)
			if err != nil {
				return fmt.Errorf("mcp: marshal response: %w", err)
			}
			fmt.Fprintln(os.Stdout, string(data))
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("mcp: stdin: %w", err)
	}
	return nil
}

func (s *Server) handleLine(line string) *Response {
	var req Request
	if err := json.Unmarshal([]byte(line), &req); err != nil {
		return &Response{
			JSONRPC: "2.0",
			Error:   &ErrorObj{Code: -32700, Message: "Parse error"},
		}
	}

	ctx := context.Background()

	if req.ID == nil {
		s.handleNotification(&req)
		return nil
	}

	return s.handleRequest(ctx, &req)
}

func (s *Server) handleNotification(req *Request) {
	if req.Method == "notifications/initialized" && s.state == stateInitialized {
		s.state = stateReady
	}
}

func (s *Server) handleRequest(ctx context.Context, req *Request) *Response {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(ctx, req)
	case "ping":
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage("{}"),
		}
	default:
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &ErrorObj{Code: -32601, Message: fmt.Sprintf("Method not found: %s", req.Method)},
		}
	}
}

func (s *Server) handleInitialize(req *Request) *Response {
	if s.state != stateNew {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &ErrorObj{Code: -32000, Message: "Already initialized"},
		}
	}

	s.state = stateInitialized

	result, _ := json.Marshal(map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]any{
			"tools": map[string]bool{
				"listChanged": false,
			},
		},
		"serverInfo": map[string]string{
			"name":    "cosmoria",
			"version": "0.1.0",
		},
	})

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func (s *Server) handleToolsList(req *Request) *Response {
	if s.state != stateReady {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &ErrorObj{Code: -32000, Message: "Not ready"},
		}
	}

	result, _ := json.Marshal(map[string]any{
		"tools": tools(),
	})

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func (s *Server) handleToolsCall(ctx context.Context, req *Request) *Response {
	if s.state != stateReady {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &ErrorObj{Code: -32000, Message: "Not ready"},
		}
	}

	var params CallToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &ErrorObj{Code: -32602, Message: "Invalid params"},
		}
	}

	for _, tool := range tools() {
		if tool.Name == params.Name {
			text, err := tool.Handler(ctx, params.Arguments, s.services)
			return toolResponse(req.ID, text, err)
		}
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Error:   &ErrorObj{Code: -32602, Message: fmt.Sprintf("Unknown tool: %s", params.Name)},
	}
}

func toolResponse(id *int, text string, err error) *Response {
	content := []ContentBlock{{Type: "text", Text: text}}
	result := map[string]any{
		"content": content,
	}
	if err != nil {
		result["isError"] = true
		result["content"] = []ContentBlock{{Type: "text", Text: err.Error()}}
	}

	data, _ := json.Marshal(result)
	return &Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  data,
	}
}
