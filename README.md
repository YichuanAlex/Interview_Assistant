# Interview Assistant

一款基于 Wails + Vue 3 的跨桌面端 AI 做题助手，支持多平台会议录屏不可视。截图 → OCR 识别 → 发送给大模型 → 查看解析，适合算法题、笔试题、代码报错等场景。

## 功能特性

- **截图解题**：快捷键截取屏幕区域，自动识别图中文字并发送给 AI。
- **多模型支持**：支持 OpenAI、DeepSeek、Anthropic、Google Gemini 等兼容 OpenAI API 的模型。
- **DeepSeek 文本模式**：DeepSeek API 暂不支持图片输入，应用会自动 OCR 截图中的文字并以文本形式发送。
- **追问模式**：在 AI 回答后，可在输入框粘贴报错或补充说明继续追问，保留当前对话上下文。
- **Markdown 渲染**：AI 返回的代码、公式、列表等内容自动渲染，支持复制代码。
- **幽灵窗口**：无边框、置顶、防抢焦、可鼠标穿透，不遮挡题目页面。
- **跨平台**：macOS 使用原生 Vision OCR；Windows 可接入 Wechat_OCR。

## 环境要求

- **Go** 1.25+
- **Node.js** 22+
- **Wails CLI** v2.12+
- **macOS** 12+ 或 **Windows** 10/11

安装 Wails CLI：

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## 快速开始

### 1. 启动开发模式

双击项目根目录下的启动脚本：

```bash
./Start_Interview_Assistant.command
```

或手动执行：

```bash
cd Interview_Assistant
wails dev
```

### 2. 运行发布版

构建完成后，直接运行：

```bash
./build/bin/Interview_Assistant.app
```

或拖拽到 `/Applications` 目录后从启动台打开。

### 3. 生产构建

```bash
wails build -ldflags "-s -w" -tags prod
```

构建产物位于 `build/bin/Interview_Assistant.app`。

## 配置方法

首次启动后，点击右上角设置按钮进入配置面板：

| 配置项   | 说明             | 示例                             |
| -------- | ---------------- | -------------------------------- |
| API Key  | 你的模型 API Key | `sk-...`                       |
| Base URL | 模型服务商地址   | `https://api.deepseek.com/v1`  |
| Model    | 模型名称         | `deepseek-v4-pro` / `gpt-4o` |

### 常用模型配置

**DeepSeek**

- Base URL：`https://api.deepseek.com/v1`
- Model：`deepseek-v4-pro` 或 `deepseek-v4-flash`

**OpenAI**

- Base URL：`https://api.openai.com/v1`
- Model：`gpt-4o`

**Google Gemini（OpenAI 兼容端点）**

- Base URL：`https://generativelanguage.googleapis.com/v1beta/openai`
- Model：`gemini-2.0-flash`

配置会自动保存到系统配置目录，下次启动无需重新填写。

## 使用方法

### 快捷键

| 功能             | macOS                     | 说明                     |
| ---------------- | ------------------------- | ------------------------ |
| 截图             | `Cmd + 1`               | 截取屏幕区域并自动 OCR   |
| 发送             | `Cmd + J`               | 发送当前题目给 AI        |
| 显示/隐藏窗口    | `Cmd + 2`               | 切换窗口可见性           |
| 鼠标穿透         | `Cmd + 3`               | 让鼠标点击穿透到后方窗口 |
| 删除最后一张截图 | `Cmd + D`               | 清空当前截图缓冲         |
| 移动窗口         | `Cmd + Option + 方向键` | 微调窗口位置             |

### 顶部按钮

- 红 / 黄 / 绿三键：关闭应用、最小化、全屏切换。
- 设置齿轮：打开配置面板。

### 使用流程

1. 打开题目页面，按 `Cmd + 1` 截图。
2. 等待底部输入框自动填入 OCR 识别结果。
3. 检查并编辑识别结果，确认无误后按 `Cmd + J` 发送。
4. AI 返回三段式内容：题目解析、完整 Python 代码、完整 C++ 代码。
5. 如果代码运行报错，把报错粘贴到底部输入框，按 `Enter` 或点击发送继续追问。

## 平台适配

### macOS

- 使用系统原生 **Vision** 框架进行 OCR，无需额外依赖。
- 截图需要授权：首次按 `Cmd + 1` 时会提示申请屏幕录制权限，请前往「系统设置 → 隐私与安全性 → 屏幕录制」开启。

### Windows

- 项目内置 `Wechat_OCR/` 目录，可调用微信 OCR 引擎识别截图文字。
- 需要安装 Python，并确保 `python` 命令可用。
- 将题目截图保存后，OCR 服务会调用 `Wechat_OCR/ocr_wrapper.py` 进行识别。

## 项目结构

```
Interview_Assistant/
├── app/                  # Go 后端业务逻辑
├── frontend/             # Vue 3 前端
├── pkg/                  # 公共包
│   ├── config/           # 配置管理
│   ├── llm/              # 大模型适配器
│   ├── ocr/              # OCR 服务（macOS Vision / Windows Wechat_OCR）
│   ├── platform/         # 平台相关 API
│   ├── screen/           # 截图服务
│   ├── shortcut/         # 全局快捷键
│   ├── solution/         # 解题流程
│   └── state/            # 应用状态管理
├── scripts/              # 辅助脚本（如 macOS OCR Swift 脚本）
├── Wechat_OCR/           # Windows OCR 引擎
├── build/                # 构建产物与资源
├── Start_Interview_Assistant.command  # 开发模式启动脚本
├── main.go               # 程序入口
├── wails.json            # Wails 配置
└── go.mod                # Go 依赖
```

## 注意事项

- DeepSeek 当前 API 不支持图片输入，因此会走 OCR 文本模式；视觉模型（如 GPT-4o、Gemini）可直接看图。
- 若 macOS 提示「无法打开」，请在终端执行：
  ```bash
  xattr -cr build/bin/Interview_Assistant.app
  ```
- 开发模式下终端会保留并显示运行日志；关闭前端界面后终端会自动退出。

## License

[LICENSE](./LICENSE)
