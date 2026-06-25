package app

import (
	"Interview_Assistant/pkg/config"
	"Interview_Assistant/pkg/llm"
	"Interview_Assistant/pkg/logger"
	"Interview_Assistant/pkg/ocr"
	"Interview_Assistant/pkg/resume"
	"Interview_Assistant/pkg/interview"
	"Interview_Assistant/pkg/screen"
	"Interview_Assistant/pkg/shortcut"
	"Interview_Assistant/pkg/solution"
	"Interview_Assistant/pkg/state"
	"Interview_Assistant/pkg/task"
	"Interview_Assistant/pkg/transcription"
	"context"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context

	configManager *config.ConfigManager
	stateManager  *state.StateManager
	taskManager   *task.TaskCoordinator

	llmService           *llm.Service
	resumeService        *resume.Service
	shortcutService      *shortcut.Service
	screenService        *screen.Service
	ocrService           *ocr.Service
	solver               *solution.Solver
	dualTranscription    *transcription.DualTranscription
	coach                *interview.Coach
}

func NewApp() *App {
	configManager := config.NewConfigManager()

	return &App{
		configManager: configManager,
		stateManager:  state.NewStateManager(),
		taskManager:   task.NewTaskCoordinator(),
		screenService: screen.NewService(),
		ocrService:    ocr.NewService(),
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	if err := a.configManager.Load(); err != nil {
		logger.Printf("加载配置失败: %v", err)
	}

	cfg := a.configManager.Get()
	if cfg.WindowWidth > 0 && cfg.WindowHeight > 0 {
		runtime.WindowSetSize(ctx, cfg.WindowWidth, cfg.WindowHeight)
		logger.Printf("应用保存的窗口尺寸: %dx%d", cfg.WindowWidth, cfg.WindowHeight)
	}

	a.stateManager.Startup(ctx, a.EmitEvent)
	a.screenService.Startup(ctx)

	a.llmService = llm.NewService(a.configManager.Get(), a.configManager)
	a.solver = solution.NewSolver(a.llmService.GetProvider())
	a.resumeService = resume.NewService(a.configManager.Get(), a.configManager)
	a.dualTranscription = transcription.NewDualTranscription(a.EmitEvent)
	a.coach = interview.NewCoach(a.llmService.GetProvider(), a.EmitEvent)

	// 应用启动后自动开始双音源实时转录
	go func() {
		time.Sleep(2 * time.Second)
		if a.dualTranscription == nil {
			return
		}
		// 面试官：系统音频/会议软件（优先 OrayVirtualAudioDevice，备选 BlackHole）
		// 面试者：麦克风（优先 MacBook Pro 麦克风）
		err := a.dualTranscription.Start("OrayVirtualAudioDevice", "MacBook Pro", "./models/small", "zh")
		if err != nil {
			logger.Printf("自动启动双音源转录失败: %v\n", err)
		}
	}()

	a.shortcutService = shortcut.NewService(a, a.configManager.Get().Shortcuts, func(callback func(map[string]shortcut.KeyBinding)) {
		a.configManager.Subscribe(func(newConfig config.Config, oldConfig config.Config) {
			callback(newConfig.Shortcuts)
		})
	})
	a.shortcutService.Start()

	a.configManager.Subscribe(a.onConfigChanged)

	// 监听实时转录文本，添加到面试提示上下文
	runtime.EventsOn(ctx, "transcription", func(optionalData ...interface{}) {
		if len(optionalData) < 2 {
			return
		}
		text, ok := optionalData[1].(string)
		if !ok || text == "" {
			return
		}
		role := ""
		if len(optionalData) >= 3 {
			role, _ = optionalData[2].(string)
		}
		if a.coach != nil {
			a.coach.AddTranscript(text, role)
		}
	})

	a.stateManager.UpdateInitStatus(state.StatusReady)
}

func (a *App) onConfigChanged(newConfig config.Config, oldConfig config.Config) {
	if a.solver != nil {
		a.solver.SetProvider(a.llmService.GetProvider())
	}
	if a.coach != nil {
		a.coach.SetProvider(a.llmService.GetProvider())
	}

	if !newConfig.KeepContext && a.solver != nil {
		a.solver.ClearHistory()
	}

	logger.Println("配置已更新并应用")
}

func (a *App) OnShutdown(ctx context.Context) {
	if a.shortcutService != nil {
		a.shortcutService.Stop()
	}
	if err := a.configManager.Save(); err != nil {
		logger.Printf("保存配置失败: %v", err)
	}
}

func (a *App) EmitEvent(eventName string, data ...interface{}) {
	runtime.EventsEmit(a.ctx, eventName, data...)
}

func (a *App) Show() {
	runtime.WindowShow(a.ctx)
}

// StartTranscription 启动双音源实时语音转录
func (a *App) StartTranscription(interviewerDeviceName string, intervieweeDeviceName string, model string, language string) string {
	if a.dualTranscription == nil {
		return "转录服务未初始化"
	}
	if err := a.dualTranscription.Start(interviewerDeviceName, intervieweeDeviceName, model, language); err != nil {
		return err.Error()
	}
	return ""
}

// StopTranscription 停止双音源实时语音转录
func (a *App) StopTranscription() string {
	if a.dualTranscription == nil {
		return "转录服务未初始化"
	}
	if err := a.dualTranscription.Stop(); err != nil {
		return err.Error()
	}
	return ""
}

// IsTranscribing 返回是否正在转录
func (a *App) IsTranscribing() bool {
	if a.dualTranscription == nil {
		return false
	}
	return a.dualTranscription.IsRunning()
}

// GenerateInterviewHint 根据当前转录上下文生成面试提示
func (a *App) GenerateInterviewHint() string {
	if a.coach == nil {
		return "面试提示服务未初始化"
	}
	hint, err := a.coach.GenerateHint(context.Background())
	if err != nil {
		return "生成提示失败: " + err.Error()
	}
	return hint
}

// ClearInterviewContext 清空面试提词上下文
func (a *App) ClearInterviewContext() string {
	if a.coach == nil {
		return "面试提示服务未初始化"
	}
	a.coach.ClearHistory()
	return ""
}

// MinimiseWindow 最小化窗口
func (a *App) MinimiseWindow() {
	runtime.WindowMinimise(a.ctx)
}

// ToggleMaximiseWindow 切换窗口最大化状态
func (a *App) ToggleMaximiseWindow() {
	runtime.WindowToggleMaximise(a.ctx)
}
