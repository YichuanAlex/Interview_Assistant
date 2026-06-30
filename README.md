# Interview Assistant / 面试辅助助手

## 1. Overview / 项目概览

**中文**

Interview Assistant 是一个基于 Wails + Vue 3 + Go 的桌面端 AI 面试与做题辅助工具。它可以在面试或笔试过程中完成截图 OCR、题目解析、双音源实时转录、面试官问题识别、个性化回答提示生成，并支持隐身悬浮窗口、置顶、防抢焦、鼠标穿透等桌面能力。

本项目当前重点面向两类场景：

- 截图解题：截取题目、OCR 识别、发送给大模型生成解法。
- 实时面试辅助：自动捕获面试官系统音频和面试者麦克风音频，转成文字，并结合本地个性化材料生成简短可复述的面试提示。

**English**

Interview Assistant is a desktop AI assistant built with Wails, Vue 3 and Go. It supports screenshot OCR, problem solving, dual-source live speech transcription, interviewer question detection, personalized interview hint generation, and a floating stealth window with always-on-top, focus protection and click-through behavior.

The project currently focuses on two workflows:

- Screenshot solving: capture a question, run OCR, and send it to an LLM for solutions.
- Live interview assistance: capture interviewer audio from system audio and candidate audio from the microphone, transcribe both streams, and generate short personalized answer hints from local preparation materials.

## 2. Key Features / 核心功能

**中文**

- 截图解题：`Cmd + 1` 截图，自动 OCR，支持继续追问。
- 多模型支持：支持 OpenAI 兼容接口，例如 OpenAI、DeepSeek、Gemini OpenAI-compatible endpoint 等。
- DeepSeek 文本模式：DeepSeek 不支持图片输入时，自动走 OCR 文本。
- 双音源实时转录：面试官默认走 macOS ScreenCaptureKit 系统音频，面试者默认走系统默认输入设备。
- 个性化面试提示：本地读取 `material/mine` 和 `material/preparation`，用轻量检索选出相关材料，再注入远端 LLM 提示。
- 低延迟材料检索：不部署本地 embedding 模型，不启动向量数据库，只用本地 Markdown 切块和关键词检索，适合实时面试。
- 幽灵窗口：无边框、置顶、防抢焦、可鼠标穿透。
- macOS 原生 OCR：使用 Vision OCR。
- Windows OCR：可调用项目内置 WeChat OCR 相关资源。

**English**

- Screenshot solving: press `Cmd + 1` to capture a question, run OCR, and continue follow-up questions.
- Multi-model support: works with OpenAI-compatible APIs such as OpenAI, DeepSeek, Gemini OpenAI-compatible endpoints and others.
- DeepSeek text mode: when image input is unsupported, OCR text is sent instead.
- Dual-source transcription: interviewer audio uses macOS ScreenCaptureKit system audio by default, while candidate audio uses the system default input device.
- Personalized interview hints: reads local `material/mine` and `material/preparation`, retrieves relevant snippets, and injects them into the remote LLM prompt.
- Low-latency local retrieval: no local embedding model or vector database is required. Markdown files are chunked and searched locally with lightweight keyword scoring.
- Stealth floating window: borderless, always-on-top, focus-safe and click-through capable.
- Native macOS OCR: powered by Apple Vision.
- Windows OCR: can use the bundled WeChat OCR resources.

## 3. Project Structure / 项目结构

```text
Interview_Assistant/
├── app/                         # Wails-bound Go application APIs
├── pkg/
│   ├── interview/               # Interview coach and hint generation
│   ├── knowledge/               # Local personalized material retrieval
│   ├── transcription/           # Dual-source transcription process manager
│   ├── llm/                     # OpenAI-compatible LLM adapters
│   ├── ocr/                     # macOS Vision OCR and Windows OCR bridge
│   ├── screen/                  # Screenshot service
│   ├── shortcut/                # Global hotkeys
│   └── platform/                # macOS and Windows native integrations
├── BUZZ/faster-whisper/
│   ├── realtime_transcribe.py   # Realtime faster-whisper transcription
│   ├── system_audio_capture.swift # macOS system audio capture helper
│   └── models/                  # Local faster-whisper models
├── frontend/                    # Vue 3 frontend
├── material/
│   ├── mine/                    # Candidate CV, projects, personal profile
│   └── preparation/             # Interview preparation notes
├── build/bin/Interview_Assistant.app
├── Start_Interview_Assistant.command
└── wails.json
```

## 4. Requirements / 环境要求

**中文**

推荐环境：

- macOS 12+，推荐 Apple Silicon。
- Go 1.24+。
- Node.js 22+。
- Python 3.11。
- Wails CLI v2.12。
- Xcode Command Line Tools，用于 `swiftc` 编译系统音频采集 helper。
- macOS 权限：麦克风、屏幕录制。

Python 依赖见 [requirements.txt](requirements.txt)，核心包括：

- `faster-whisper`
- `sounddevice`
- `numpy`
- `onnxruntime`

**English**

Recommended environment:

- macOS 12+, Apple Silicon recommended.
- Go 1.24+.
- Node.js 22+.
- Python 3.11.
- Wails CLI v2.12.
- Xcode Command Line Tools, required for compiling the Swift system-audio helper with `swiftc`.
- macOS permissions: Microphone and Screen Recording.

Python dependencies are listed in [requirements.txt](requirements.txt). The main packages are:

- `faster-whisper`
- `sounddevice`
- `numpy`
- `onnxruntime`

## 5. Setup From Scratch / 新机器从零配置

### 5.1 Clone or Copy the Project / 获取项目

**中文**

把项目放到本地，例如：

```bash
cd ~/Downloads
git clone <your-repo-url> Interview_Assistant
cd Interview_Assistant
```

如果你不是通过 git 获取，而是复制文件夹，也请确保 `BUZZ/faster-whisper/models`、`material`、`frontend`、`pkg` 等目录完整存在。

**English**

Place the project locally, for example:

```bash
cd ~/Downloads
git clone <your-repo-url> Interview_Assistant
cd Interview_Assistant
```

If you copy the folder manually instead of cloning, make sure `BUZZ/faster-whisper/models`, `material`, `frontend`, `pkg` and the other project directories are present.

### 5.2 Install Go / 安装 Go

**中文**

安装 Go 后确认：

```bash
go version
```

**English**

Install Go and verify:

```bash
go version
```

### 5.3 Install Node.js / 安装 Node.js

**中文**

推荐 Node.js 22：

```bash
node -v
npm -v
cd frontend
npm install
cd ..
```

**English**

Node.js 22 is recommended:

```bash
node -v
npm -v
cd frontend
npm install
cd ..
```

### 5.4 Install Wails / 安装 Wails

**中文**

可以全局安装：

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
wails doctor
```

如果 `wails` 不在 PATH 中，也可以直接用：

```bash
go run github.com/wailsapp/wails/v2/cmd/wails@v2.12.0 build
```

**English**

You can install Wails globally:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
wails doctor
```

If `wails` is not available in PATH, use:

```bash
go run github.com/wailsapp/wails/v2/cmd/wails@v2.12.0 build
```

### 5.5 Install Python Environment / 配置 Python 环境

**中文**

推荐使用 conda：

```bash
conda create -n interview python=3.11 -y
conda activate interview
pip install -r requirements.txt
```

如果使用项目当前启动脚本，脚本默认会尝试激活：

```text
/opt/homebrew/Caskroom/miniconda/base/envs/interview/bin/python3
```

新机器路径不同的话，请修改 [Start_Interview_Assistant.command](Start_Interview_Assistant.command) 里的 `CONDA_BASE`，或者保证 `python3` 在 PATH 中可用。

**English**

Conda is recommended:

```bash
conda create -n interview python=3.11 -y
conda activate interview
pip install -r requirements.txt
```

The current start script tries to activate:

```text
/opt/homebrew/Caskroom/miniconda/base/envs/interview/bin/python3
```

On a new machine, update `CONDA_BASE` in [Start_Interview_Assistant.command](Start_Interview_Assistant.command), or make sure `python3` is available in PATH.

### 5.6 Prepare faster-whisper Models / 准备 faster-whisper 模型

**中文**

默认转录模型路径是：

```text
BUZZ/faster-whisper/models/small
```

推荐同时准备：

```text
BUZZ/faster-whisper/models/small
BUZZ/faster-whisper/models/large-v3-turbo
```

`small` 延迟低，适合实时；`large-v3-turbo` 准确率更高，但 CPU 占用和延迟更高。也可以直接传 Hugging Face 模型名，但面试场景建议提前下载到本地，避免现场联网下载。

**English**

The default transcription model path is:

```text
BUZZ/faster-whisper/models/small
```

Recommended local models:

```text
BUZZ/faster-whisper/models/small
BUZZ/faster-whisper/models/large-v3-turbo
```

`small` is faster and better for realtime use. `large-v3-turbo` is more accurate but uses more CPU and has higher latency. Hugging Face model names can also be used, but local models are recommended for interviews.

### 5.7 macOS Permissions / macOS 权限

**中文**

首次运行前或首次触发功能时，请在系统设置中允许：

- 麦克风：用于采集面试者声音。
- 屏幕录制：用于截图 OCR，也用于 ScreenCaptureKit 系统音频采集。

路径通常为：

```text
系统设置 -> 隐私与安全性 -> 麦克风
系统设置 -> 隐私与安全性 -> 屏幕录制
```

如果授权后仍无效，重启应用。某些 macOS 权限变更需要完全退出应用后重新打开。

**English**

Before first use, or when prompted, allow:

- Microphone: captures the candidate's voice.
- Screen Recording: required for screenshots and ScreenCaptureKit system-audio capture.

Usually found at:

```text
System Settings -> Privacy & Security -> Microphone
System Settings -> Privacy & Security -> Screen Recording
```

If permission changes do not take effect immediately, quit and restart the app.

## 6. LLM Configuration / 大模型配置

**中文**

打开应用设置面板，填写：

- API Key
- Base URL
- Model

常见配置：

```text
DeepSeek:
Base URL = https://api.deepseek.com/v1
Model    = deepseek-chat 或你的 DeepSeek 模型名

OpenAI:
Base URL = https://api.openai.com/v1
Model    = gpt-4o / gpt-4.1 / 其他模型

Gemini OpenAI-compatible:
Base URL = https://generativelanguage.googleapis.com/v1beta/openai
Model    = gemini-2.0-flash 或兼容模型
```

配置会保存到：

```text
~/Library/Application Support/Interview_Assistant/config
```

**English**

Open the settings panel and configure:

- API Key
- Base URL
- Model

Common examples:

```text
DeepSeek:
Base URL = https://api.deepseek.com/v1
Model    = deepseek-chat or your DeepSeek model name

OpenAI:
Base URL = https://api.openai.com/v1
Model    = gpt-4o / gpt-4.1 / another model

Gemini OpenAI-compatible:
Base URL = https://generativelanguage.googleapis.com/v1beta/openai
Model    = gemini-2.0-flash or a compatible model
```

Settings are saved to:

```text
~/Library/Application Support/Interview_Assistant/config
```

## 7. Running the App / 启动方式

### Production App / 生产包启动

**中文**

构建后运行：

```bash
./Start_Interview_Assistant.command
```

该脚本会：

- 进入项目目录。
- 激活 conda `interview` 环境。
- 清除 macOS quarantine 属性。
- 运行 `build/bin/Interview_Assistant.app/Contents/MacOS/Interview_Assistant`。

如果项目路径不是 `/Users/didi/Downloads/Interview_Assistant`，请修改启动脚本中的路径。

**English**

After building, run:

```bash
./Start_Interview_Assistant.command
```

The script:

- Enters the project directory.
- Activates the conda `interview` environment.
- Clears macOS quarantine attributes.
- Runs `build/bin/Interview_Assistant.app/Contents/MacOS/Interview_Assistant`.

If your project path is not `/Users/didi/Downloads/Interview_Assistant`, update the paths in the start script.

### Development Mode / 开发模式

**中文**

```bash
wails dev
```

或：

```bash
go run github.com/wailsapp/wails/v2/cmd/wails@v2.12.0 dev
```

**English**

```bash
wails dev
```

Or:

```bash
go run github.com/wailsapp/wails/v2/cmd/wails@v2.12.0 dev
```

### Build / 构建

**中文**

```bash
go run github.com/wailsapp/wails/v2/cmd/wails@v2.12.0 build
```

构建产物：

```text
build/bin/Interview_Assistant.app
```

**English**

```bash
go run github.com/wailsapp/wails/v2/cmd/wails@v2.12.0 build
```

Build output:

```text
build/bin/Interview_Assistant.app
```

## 8. Screenshot Solving / 截图解题

**中文**

默认快捷键：

| 功能 | 快捷键 |
| --- | --- |
| 截图 OCR | `Cmd + 1` |
| 显示/隐藏窗口 | `Cmd + 2` |
| 鼠标穿透 | `Cmd + 3` |
| 发送给模型 | `Cmd + J` |
| 删除最后截图 | `Cmd + D` |
| 移动窗口 | `Cmd + Option + 方向键` |

流程：

1. 打开题目页面。
2. 按 `Cmd + 1` 截图。
3. 等 OCR 文字进入输入框。
4. 检查文字后按 `Cmd + J`。
5. 根据模型返回继续追问。

**English**

Default hotkeys:

| Action | Hotkey |
| --- | --- |
| Screenshot OCR | `Cmd + 1` |
| Show or hide window | `Cmd + 2` |
| Click-through | `Cmd + 3` |
| Send to model | `Cmd + J` |
| Delete last screenshot | `Cmd + D` |
| Move window | `Cmd + Option + Arrow` |

Workflow:

1. Open the question page.
2. Press `Cmd + 1`.
3. Wait for OCR text to fill the input box.
4. Review the text and press `Cmd + J`.
5. Continue follow-up questions if needed.

## 9. Realtime Interview Mode / 实时面试模式

**中文**

应用启动后会自动尝试开始双音源转录：

- 面试官：`interviewer`，默认使用 macOS ScreenCaptureKit 系统音频。
- 面试者：`interviewee`，默认使用系统默认输入设备，也就是你在系统设置里选择的麦克风。

前端实时转录面板会显示两类角色：

- `面试官`：用于触发自动提示。
- `我`：作为上下文参考，不主动触发提示。

当面试官停止说话约 2 秒后，Coach 会判断最新内容是否像一个实质问题。如果只是“你好”“喂”“嗯”等寒暄，不会调用远端 API。若是实质问题，会结合个性化材料生成简短提示。

**English**

The app automatically tries to start dual-source transcription on launch:

- Interviewer: `interviewer`, using macOS ScreenCaptureKit system audio by default.
- Candidate: `interviewee`, using the system default input device.

The live transcript panel shows:

- `面试官`: interviewer audio, used to trigger automatic hints.
- `我`: candidate audio, used as context only.

After the interviewer stops speaking for about 2 seconds, the Coach checks whether the latest utterance looks like a real question. Greetings such as "hello" do not trigger the remote API. Real questions trigger a short personalized hint.

## 10. Personalized Materials / 个性化面试材料

**中文**

本项目会自动读取：

```text
material/mine
material/preparation
```

推荐放置内容：

```text
material/mine/
  your_cv.md
  project_summary.md
  personal_pitch.md

material/preparation/
  role_notes.md
  company_notes.md
  technical_topics.md
  mock_answers.md
```

当前实现方式是本地轻量检索：

1. 启动时扫描 Markdown 和 txt 文件。
2. 将文档按段落切成片段。
3. 对最新面试官问题做关键词和中英文 token 匹配。
4. 选出最相关的少量片段。
5. 把这些片段注入远端 LLM prompt。

这样做的优点：

- 本地部署最轻量。
- 不需要向量数据库。
- 不需要本地 embedding 模型。
- 不需要每次把全部材料发给 API。
- 响应速度适合实时面试。

材料写法建议：

- 简历和项目经历放在 `material/mine`。
- 岗位 JD、公司材料、技术准备、模拟回答放在 `material/preparation`。
- 每个项目尽量包含：背景、目标、技术栈、难点、量化结果、可复述话术。
- 文件可中英混写，系统会做中英文混合检索。

**English**

The app automatically reads:

```text
material/mine
material/preparation
```

Recommended layout:

```text
material/mine/
  your_cv.md
  project_summary.md
  personal_pitch.md

material/preparation/
  role_notes.md
  company_notes.md
  technical_topics.md
  mock_answers.md
```

The current implementation uses lightweight local retrieval:

1. Scan Markdown and txt files on startup.
2. Split documents into paragraph-based chunks.
3. Match the latest interviewer question with mixed Chinese and English tokens.
4. Select a small number of relevant chunks.
5. Inject those chunks into the remote LLM prompt.

Benefits:

- Minimal local deployment.
- No vector database.
- No local embedding model.
- No need to send all materials on every request.
- Fast enough for live interviews.

Material writing tips:

- Put CV and personal project experience in `material/mine`.
- Put job descriptions, company notes, technical notes and mock answers in `material/preparation`.
- For each project, include background, goal, stack, difficulty, measurable results and repeatable talking points.
- Chinese and English can be mixed.

## 11. Audio and Transcription Details / 音频与转录细节

**中文**

脚本：

```text
BUZZ/faster-whisper/realtime_transcribe.py
```

macOS 系统音频 helper：

```text
BUZZ/faster-whisper/system_audio_capture.swift
```

检查设备绑定：

```bash
/path/to/python3 BUZZ/faster-whisper/realtime_transcribe.py --role interviewer --json-output --dry-run
/path/to/python3 BUZZ/faster-whisper/realtime_transcribe.py --role interviewee --json-output --dry-run
```

预期：

```text
role=interviewer, source=system-audio
role=interviewee, source=input, target=<your default microphone>
```

真实短测：

```bash
/path/to/python3 BUZZ/faster-whisper/realtime_transcribe.py \
  --role interviewer \
  --model ./models/small \
  --language zh \
  --json-output
```

然后用 macOS `say` 或会议软件播放一句话，stdout 应该只输出 JSON 行。

**English**

Main script:

```text
BUZZ/faster-whisper/realtime_transcribe.py
```

macOS system-audio helper:

```text
BUZZ/faster-whisper/system_audio_capture.swift
```

Check device binding:

```bash
/path/to/python3 BUZZ/faster-whisper/realtime_transcribe.py --role interviewer --json-output --dry-run
/path/to/python3 BUZZ/faster-whisper/realtime_transcribe.py --role interviewee --json-output --dry-run
```

Expected:

```text
role=interviewer, source=system-audio
role=interviewee, source=input, target=<your default microphone>
```

Short live test:

```bash
/path/to/python3 BUZZ/faster-whisper/realtime_transcribe.py \
  --role interviewer \
  --model ./models/small \
  --language zh \
  --json-output
```

Play a sentence with macOS `say` or your meeting app. stdout should contain JSON transcription lines only.

## 12. Testing / 测试

**中文**

Go 测试：

```bash
go test ./...
```

Python 语法检查：

```bash
/path/to/python3 -m py_compile BUZZ/faster-whisper/realtime_transcribe.py
```

前端构建：

```bash
cd frontend
npm run build
cd ..
```

Wails 生产构建：

```bash
go run github.com/wailsapp/wails/v2/cmd/wails@v2.12.0 build
```

**English**

Go tests:

```bash
go test ./...
```

Python syntax check:

```bash
/path/to/python3 -m py_compile BUZZ/faster-whisper/realtime_transcribe.py
```

Frontend build:

```bash
cd frontend
npm run build
cd ..
```

Wails production build:

```bash
go run github.com/wailsapp/wails/v2/cmd/wails@v2.12.0 build
```

## 13. Troubleshooting / 常见问题

### The interviewer audio is not transcribed / 面试官声音没有转录

**中文**

检查：

1. macOS 屏幕录制权限是否给了应用或终端。
2. `system_audio_capture.swift` 是否存在。
3. `swiftc` 是否可用：

```bash
swift --version
```

4. dry-run 是否显示：

```text
source=system-audio
```

**English**

Check:

1. Screen Recording permission is granted to the app or terminal.
2. `system_audio_capture.swift` exists.
3. `swiftc` is available:

```bash
swift --version
```

4. dry-run shows:

```text
source=system-audio
```

### The candidate and interviewer lines are duplicated / 我与面试官出现重复文本

**中文**

如果扬声器声音被麦克风再次收录，系统会做短时间重复文本过滤，优先保留面试官音源。建议戴耳机，可以进一步降低回声。

**English**

If speaker audio is picked up by the microphone, the app filters near-duplicate lines within a short time window and prefers interviewer audio. Wearing headphones further reduces echo.

### The app prints many transcription logs / 终端转录日志太多

**中文**

JSON 模式下，stdout 只用于转录 JSON，设备列表和状态信息走 stderr。默认关闭 faster-whisper 的 INFO 处理日志。如果需要调试，给脚本加 `--debug`。

**English**

In JSON mode, stdout is reserved for transcription JSON only. Device and status logs go to stderr. faster-whisper INFO logs are suppressed by default. Use `--debug` for debugging.

### DeepSeek does not accept images / DeepSeek 不支持图片

**中文**

这是 DeepSeek API 限制。项目会自动使用 OCR 文本模式。若要直接看图，请切换支持视觉输入的模型。

**English**

This is a DeepSeek API limitation. The app uses OCR text mode automatically. Use a vision-capable model if direct image input is required.

### The start script path is wrong / 启动脚本路径不对

**中文**

修改 [Start_Interview_Assistant.command](Start_Interview_Assistant.command) 中的项目路径和 conda 路径。

**English**

Update the project path and conda path in [Start_Interview_Assistant.command](Start_Interview_Assistant.command).

## 14. Recommended Interview Workflow / 推荐面试使用流程

**中文**

1. 面试前把简历、项目、岗位 JD、公司技术材料放进 `material/mine` 和 `material/preparation`。
2. 启动应用，确认右侧进入“转录中”。
3. 用 dry-run 或真实会议测试确认面试官音源是 `system-audio`。
4. 面试时让窗口保持置顶但不抢焦。
5. 面试官问完问题后，等自动提示生成。
6. 回答时优先使用提示中的个人项目和量化结果。

**English**

1. Before the interview, place your CV, projects, job description and company notes under `material/mine` and `material/preparation`.
2. Start the app and confirm transcription is running.
3. Use dry-run or a meeting test to confirm interviewer audio is `system-audio`.
4. Keep the window always-on-top but focus-safe.
5. Wait for automatic hints after the interviewer finishes a question.
6. Answer with your own projects and measurable results from the hint.

## 15. License / 许可证

See [LICENSE](LICENSE).