package llm

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync"

	"github.com/openai/openai-go"
)

var (
	ErrInvalidContextIndex = fmt.Errorf("invalid context index")
)

type TalkContextInfo struct {
	Index   int
	Title   string
	Current bool
}

type TalkContext struct {
	Messages []openai.ChatCompletionMessageParamUnion `json:"messages"`
}

// IsEmpty checks if the TalkContext has no messages.
func (t *TalkContext) IsEmpty() bool {
	return len(t.Messages) == 0
}

// TalkContextManager manages multiple TalkContexts.
type TalkContextManager struct {
	mu       sync.RWMutex
	contexts []*TalkContext
	current  int

	restoreOnce sync.Once
}

// addMsg adds a message to the current TalkContext, creating it if necessary.
func (m *TalkContextManager) addMsg(msg openai.ChatCompletionMessageParamUnion) *TalkContext {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.current == 0 && len(m.contexts) == 0 {
		m.contexts = append(m.contexts, &TalkContext{})
		m.current = len(m.contexts) - 1
	}
	m.contexts[m.current].Messages = append(m.contexts[m.current].Messages, msg)
	return m.contexts[m.current]
}

// setMsgs sets the messages of the current TalkContext, creating it if necessary.
func (m *TalkContextManager) setMsgs(msgs []openai.ChatCompletionMessageParamUnion) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.current == 0 && len(m.contexts) == 0 {
		m.contexts = append(m.contexts, &TalkContext{})
		m.current = len(m.contexts) - 1
	}
	m.contexts[m.current].Messages = msgs
}

// Clear clears the messages of the current context.
func (m *TalkContextManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.contexts) == 0 {
		return
	}
	m.contexts[m.current].Messages = []openai.ChatCompletionMessageParamUnion{}
}

// New creates a new TalkContext and sets it as the current context.
func (m *TalkContextManager) New() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.contexts = append(m.contexts, &TalkContext{})
	m.current = len(m.contexts) - 1
}

// Delete removes a context by index and adjusts the current context if necessary.
func (m *TalkContextManager) Delete(index int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if index < 0 || index >= len(m.contexts) {
		return ErrInvalidContextIndex
	}
	m.contexts = append(m.contexts[:index], m.contexts[index+1:]...)
	if m.current >= index {
		m.current = m.current - 1
	}
	return nil
}

// Current returns the current TalkContext, creating a new one if none exists.
func (m *TalkContextManager) Current() *TalkContext {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.current == 0 && len(m.contexts) == 0 {
		m.contexts = append(m.contexts, &TalkContext{})
		m.current = len(m.contexts) - 1
	}
	return m.contexts[m.current]
}

// List returns a list of TalkContextInfo for all contexts, including the current one.
func (m *TalkContextManager) List() []*TalkContextInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var ctxs []*TalkContextInfo
	for i, ctx := range m.contexts {
		info := TalkContextInfo{
			Index: i,
			Title: "New context",
		}
		if len(ctx.Messages) > 0 {
			info.Title = reflect.TypeOf(ctx.Messages[0].GetContent().AsAny()).String()
			if str, ok := ctx.Messages[0].GetContent().AsAny().(*string); ok {
				info.Title = *str
			}
		}
		if i == m.current {
			info.Current = true
		}
		ctxs = append(ctxs, &info)
	}
	return ctxs
}

// SwitchTo switches the current context to the one at the specified index.
func (m *TalkContextManager) SwitchTo(index int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if index < 0 || index >= len(m.contexts) {
		return ErrInvalidContextIndex
	}
	m.current = index
	return nil
}

// Save contexts to file
func (m *TalkContextManager) Save(store string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	f, err := os.Create(store)
	if err != nil {
		return fmt.Errorf("create llm contexts store: %w", err)
	}
	if err := json.NewEncoder(f).Encode(map[string]any{"current": m.current, "contexts": m.contexts}); err != nil {
		return fmt.Errorf("encode llm contexts store: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close llm contexts store: %w", err)
	}
	return nil
}

// Load contexts from file
func (m *TalkContextManager) Load(store string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	f, err := os.Open(store)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil // No file to load, just return
		}
		return fmt.Errorf("open llm contexts store: %w", err)
	}
	defer f.Close()
	type contexts struct {
		Current  int            `json:"current"`
		Contexts []*TalkContext `json:"contexts"`
	}
	var ctxs contexts
	if err := json.NewDecoder(f).Decode(&ctxs); err != nil {
		return fmt.Errorf("decode llm contexts store: %w", err)
	}
	m.contexts = ctxs.Contexts
	m.current = ctxs.Current
	return nil
}

// LoadOnce loads the contexts from the store only once, using sync.Once to ensure it is only done once.
func (m *TalkContextManager) LoadOnce(store string) (err error) {
	m.restoreOnce.Do(func() {
		if err = m.Load(store); err != nil {
			return
		}
	})
	return
}
