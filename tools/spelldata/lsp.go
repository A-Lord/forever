package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/textproto"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	codeParseError     = -32700
	codeMethodNotFound = -32601
	codeInternalError  = -32603

	messageTypeLog = 4
)

type rpcMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type textDocumentItem struct {
	URI  string `json:"uri"`
	Text string `json:"text"`
}

type textDocumentPosition struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Position struct {
		Line      int `json:"line"`
		Character int `json:"character"`
	} `json:"position"`
}

type lspServer struct {
	out      io.Writer
	outMu    sync.Mutex
	ws       *workspace
	trace    bool
	shutdown bool
}

func serveLSP(in io.Reader, out io.Writer) (shutdownRequested bool, err error) {
	server := &lspServer{out: out, ws: newWorkspace(), trace: true}
	reader := textproto.NewReader(bufio.NewReader(in))

	for {
		body, err := readFrame(reader)
		if errors.Is(err, io.EOF) {
			return server.shutdown, nil
		}
		if err != nil {
			return server.shutdown, err
		}

		var msg rpcMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			server.respond(json.RawMessage("null"), nil, &rpcError{Code: codeParseError, Message: err.Error()})
			continue
		}
		if msg.Method == "exit" {
			return server.shutdown, nil
		}
		server.handle(msg)
	}
}

func readFrame(reader *textproto.Reader) ([]byte, error) {
	header, err := reader.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}
	length, err := strconv.Atoi(header.Get("Content-Length"))
	if err != nil {
		return nil, fmt.Errorf("a frame without a Content-Length: %v", header)
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(reader.R, body); err != nil {
		return nil, err
	}
	return body, nil
}

func (s *lspServer) handle(msg rpcMessage) {
	request := len(msg.ID) > 0
	defer func() {
		if r := recover(); r != nil {
			s.log(fmt.Sprintf("✗ %s panicked: %v", msg.Method, r))
			if request {
				s.respond(msg.ID, nil, &rpcError{Code: codeInternalError, Message: fmt.Sprint(r)})
			}
		}
	}()

	switch msg.Method {
	case "initialize":
		var params struct {
			InitializationOptions struct {
				Trace string `json:"trace"`
			} `json:"initializationOptions"`
		}
		_ = json.Unmarshal(msg.Params, &params)
		s.trace = params.InitializationOptions.Trace != "off"
		s.respond(msg.ID, map[string]any{
			"capabilities": map[string]any{
				"hoverProvider": true,
				"textDocumentSync": map[string]any{
					"openClose": true,
					"change":    1,
					"save":      map[string]any{"includeText": false},
				},
			},
			"serverInfo": map[string]any{"name": "wowsims-spelldata"},
		}, nil)

	case "shutdown":
		s.shutdown = true
		s.respond(msg.ID, nil, nil)

	case "textDocument/didOpen":
		var params struct {
			TextDocument textDocumentItem `json:"textDocument"`
		}
		if json.Unmarshal(msg.Params, &params) == nil {
			s.ws.setBuffer(params.TextDocument.URI, params.TextDocument.Text)
		}

	case "textDocument/didChange":
		var params struct {
			TextDocument   textDocumentItem `json:"textDocument"`
			ContentChanges []struct {
				Text string `json:"text"`
			} `json:"contentChanges"`
		}
		if json.Unmarshal(msg.Params, &params) == nil && len(params.ContentChanges) > 0 {
			s.ws.setBuffer(params.TextDocument.URI, params.ContentChanges[len(params.ContentChanges)-1].Text)
		}

	case "textDocument/didSave":
		var params struct {
			TextDocument textDocumentItem `json:"textDocument"`
		}
		if json.Unmarshal(msg.Params, &params) == nil {
			s.ws.invalidate(params.TextDocument.URI)
		}

	case "textDocument/didClose":
		var params struct {
			TextDocument textDocumentItem `json:"textDocument"`
		}
		if json.Unmarshal(msg.Params, &params) == nil {
			s.ws.dropBuffer(params.TextDocument.URI)
		}

	case "textDocument/hover":
		var params textDocumentPosition
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			s.respond(msg.ID, nil, &rpcError{Code: codeParseError, Message: err.Error()})
			return
		}
		s.respond(msg.ID, s.hover(params), nil)

	default:
		if request {
			s.respond(msg.ID, nil, &rpcError{Code: codeMethodNotFound, Message: "method not found: " + msg.Method})
		}
	}
}

func (s *lspServer) hover(params textDocumentPosition) any {
	uri := params.TextDocument.URI
	text, open := s.ws.buffer(uri)
	if !open {
		data, err := os.ReadFile(uriPath(uri))
		if err != nil {
			s.log("✗ " + err.Error())
			return nil
		}
		text = string(data)
	}

	markdown, trace, ok := s.ws.hover(text, params.Position.Line, params.Position.Character, uri)
	s.log(strings.Join(trace, "\n"))
	if !ok {
		return nil
	}
	return map[string]any{"contents": map[string]any{"kind": "markdown", "value": markdown}}
}

func (s *lspServer) log(message string) {
	if !s.trace || message == "" {
		return
	}
	s.notify("window/logMessage", map[string]any{"type": messageTypeLog, "message": message})
}

func (s *lspServer) respond(id json.RawMessage, result any, rpcErr *rpcError) {
	msg := rpcMessage{JSONRPC: "2.0", ID: id, Error: rpcErr}
	if rpcErr == nil {
		encoded, err := json.Marshal(result)
		if err != nil {
			msg.Error = &rpcError{Code: codeInternalError, Message: err.Error()}
		} else {
			msg.Result = encoded
		}
	}
	s.write(msg)
}

func (s *lspServer) notify(method string, params any) {
	encoded, err := json.Marshal(params)
	if err != nil {
		return
	}
	s.write(rpcMessage{JSONRPC: "2.0", Method: method, Params: encoded})
}

func (s *lspServer) write(msg rpcMessage) {
	body, err := json.Marshal(msg)
	if err != nil {
		return
	}
	s.outMu.Lock()
	defer s.outMu.Unlock()
	fmt.Fprintf(s.out, "Content-Length: %d\r\n\r\n%s", len(body), body)
}
