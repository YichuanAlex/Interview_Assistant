package transcription

import (
	"errors"
	"fmt"
	"strings"

	"Interview_Assistant/pkg/logger"
)

const (
	interviewerAccuracyModel = "./models/large-v3-turbo"
	intervieweeFastModel     = "./models/small"
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
// interviewerDeviceName: 面试官音频设备名称，留空时自动绑定系统音频/系统输出
// intervieweeDeviceName: 面试者音频设备名称，留空时自动绑定系统默认输入
func (d *DualTranscription) Start(interviewerDeviceName, intervieweeDeviceName, model, language string) error {
	if d.IsRunning() {
		return fmt.Errorf("转录已经在运行")
	}

	if model == "" || model == "auto" {
		model = interviewerAccuracyModel
	}
	if language == "" {
		language = "zh"
	}

	interviewerModel := model
	intervieweeModel := model
	if strings.Contains(model, "large-v3-turbo") {
		// 面试官问题的识别精度优先；面试者音频只做上下文，保留小模型降低双进程 CPU 压力。
		intervieweeModel = intervieweeFastModel
	}

	logger.Printf("启动双音源转录：面试官设备=%s，面试者设备=%s，面试官模型=%s，面试者模型=%s\n",
		interviewerDeviceName, intervieweeDeviceName, interviewerModel, intervieweeModel)

	// 同时启动两个子进程
	var errs []string
	if err := d.interviewer.Start(-1, interviewerDeviceName, interviewerModel, language); err != nil {
		errs = append(errs, "面试官音源: "+err.Error())
	}
	if err := d.interviewee.Start(-1, intervieweeDeviceName, intervieweeModel, language); err != nil {
		errs = append(errs, "面试者音源: "+err.Error())
	}

	if len(errs) > 0 {
		// 启动失败时停止已启动的进程
		_ = d.Stop()
		return errors.New(strings.Join(errs, "; "))
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
