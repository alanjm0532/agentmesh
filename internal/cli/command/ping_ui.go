package command

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/alanjm0532/agentmesh/internal/cli/client"
	"github.com/alanjm0532/agentmesh/pkg/protocol"
	tea "github.com/charmbracelet/bubbletea"
)

var pingFrames = [...]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type pingResultMsg struct {
	response protocol.PingResponse
	err      error
}

type pingTickMsg time.Time

type pingModel struct {
	ctx      context.Context
	cancel   context.CancelFunc
	api      *client.Client
	server   string
	frame    int
	done     bool
	response protocol.PingResponse
	err      error
}

func runPingUI(ctx context.Context, api *client.Client, server string, input io.Reader, output io.Writer) (protocol.PingResponse, error) {
	requestCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	model := pingModel{ctx: requestCtx, cancel: cancel, api: api, server: server}
	program := tea.NewProgram(model, tea.WithContext(ctx), tea.WithInput(input), tea.WithOutput(output))
	final, err := program.Run()
	if err != nil {
		return protocol.PingResponse{}, fmt.Errorf("运行 ping 交互界面失败: %w", err)
	}
	result := final.(pingModel)
	return result.response, result.err
}

func (m pingModel) Init() tea.Cmd {
	return tea.Batch(func() tea.Msg {
		response, err := m.api.Ping(m.ctx)
		return pingResultMsg{response: response, err: err}
	}, tickPing())
}

func (m pingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case pingResultMsg:
		m.response, m.err, m.done = msg.response, msg.err, true
		return m, tea.Quit
	case pingTickMsg:
		m.frame = (m.frame + 1) % len(pingFrames)
		return m, tickPing()
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			// 同时取消 HTTP 请求，避免退出界面后请求继续等待。
			m.cancel()
			m.err, m.done = fmt.Errorf("ping 已取消: %w", context.Canceled), true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m pingModel) View() string {
	if m.done {
		return ""
	}
	return fmt.Sprintf("%s 正在 ping %s（q / Ctrl+C 取消）\n", pingFrames[m.frame], m.server)
}

func tickPing() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return pingTickMsg(t)
	})
}
