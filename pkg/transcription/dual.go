package transcription

import (
	"fmt"
	"strings"

	"Interview_Assistant/pkg/logger"
)

// DualTranscription 同时管理面试官（系统音频）和面试者（麦克风）两个转录源
type DualTranscription struct {
	interviewer *TranscriptionService
	interviewee *TranscriptionService
}

// NewDualTranscription 创建双音源转录服务
func NewDualTranscription(emitFunc func(eventName string, data ...interface{})) *DualTranscription {
	return &DualTranscription{
		interviewer: NewTranscriptionService(emitFunc, "interviewer"),
		interviewee: NewTranscriptionService(emitFunc, "interviewee"),
	}
}

// IsRunning 返回是否正在转录
func (d *DualTranscription) IsRunning() bool {
	return d.interviewer.IsRunning() || d.interviewee.IsRunning()
}

// Start 启动双音源转录
// interviewerDeviceName: 面试官音频设备名称（系统音频/会议软件，如 OrayVirtualAudioDevice）
// intervieweeDeviceName: 面试者音频设备名称（麦克风，如 MacBook Pro 麦克风）
func (d *DualTranscription) Start(interviewerDeviceName, intervieweeDeviceName, model, language string) error {
	if d.IsRunning() {
		return fmt.Errorf("转录已经在运行")
	}

	if model == "" {
		model = "./models/small"
	}
	if language == "" {
		language = "zh"
	}

	logger.Printf("启动双音源转录：面试官设备=%s，面试者设备=%s\n", interviewerDeviceName, intervieweeDeviceName)

	// 同时启动两个子进程
	var errs []string
	if err := d.interviewer.Start(0, interviewerDeviceName, model, language); err != nil {
		errs = append(errs, "面试官音源: "+err.Error())
	}
	if err := d.interviewee.Start(0, intervieweeDeviceName, model, language); err != nil {
		errs = append(errs, "面试者音源: "+err.Error())
	}

	if len(errs) > 0 {
		// 启动失败时停止已启动的进程
		_ = d.Stop()
		return fmt.Errorf(strings.Join(errs, "; "))
	}

	return nil
}

// Stop 停止双音源转录
func (d *DualTranscription) Stop() error {
	_ = d.interviewer.Stop()
	_ = d.interviewee.Stop()
	return nil
}

// InterviewerService 返回面试官音源服务
func (d *DualTranscription) InterviewerService() *TranscriptionService {
	return d.interviewer
}

// IntervieweeService 返回面试者音源服务
func (d *DualTranscription) IntervieweeService() *TranscriptionService {
	return d.interviewee
}
