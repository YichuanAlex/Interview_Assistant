#!/usr/bin/env python3
"""
实时语音转文字（faster-whisper）

针对 macOS（Apple Silicon/Intel）本地运行优化：
- 模型：large-v3-turbo（中文/多语言识别精度高，速度接近实时）
- 设备：CPU
- 量化：int8（速度最快，精度损失小）
- beam_size=1（进一步降低延迟）

支持：
- 麦克风输入（默认 MacBook Pro 麦克风）
- 系统内录音频（需配合 BlackHole / OrayVirtualAudioDevice 等虚拟声卡）

用法：
    python3 realtime_transcribe.py
    python3 realtime_transcribe.py --device 2 --language zh --model large-v3-turbo
"""

import argparse
import logging
import platform
import re
import sys
import subprocess
import threading
import time
from pathlib import Path

import numpy as np
import sounddevice as sd

from faster_whisper import WhisperModel
from faster_whisper.vad import VadOptions, get_speech_timestamps


SAMPLE_RATE = 16000
CHUNK_SIZE = 1024  # 64ms @ 16kHz
MIN_SPEECH_DURATION_MS = 120
MIN_SILENCE_DURATION_MS = 260
SPEECH_PAD_MS = 320
MAX_BUFFER_SECONDS = 15
MIN_TRANSCRIBE_RMS = 0.0006
MIN_TRANSCRIBE_PEAK = 0.004
MAX_NO_SPEECH_PROB = 0.95
MIN_AVG_LOGPROB = -1.4
INITIAL_PROMPT = (
    "以下是中文技术面试对话，可能涉及机器人数据闭环、多模态大模型、"
    "RAG、数据治理、模型训练、系统架构、推荐系统、数据库和工程项目。"
    "请忠实、完整、准确地转写面试官的问题。"
)


def setup_logging(debug: bool):
    level = logging.DEBUG if debug else logging.INFO
    logging.basicConfig(
        level=level,
        format="%(asctime)s [%(levelname)s] %(message)s",
        datefmt="%H:%M:%S",
    )


def list_audio_devices():
    devices = sd.query_devices()
    default_input, default_output = sd.default.device
    print("\n可用音频设备：")
    for i, dev in enumerate(devices):
        markers = []
        if i == default_input:
            markers.append("默认输入")
        if i == default_output:
            markers.append("默认输出")
        marker = f" [{' / '.join(markers)}]" if markers else ""
        print(
            f"  {i}: {dev['name']} "
            f"(in={dev['max_input_channels']}, out={dev['max_output_channels']}){marker}"
        )
    print()
    return devices


def load_model(model_size: str, cpu_threads: int, compute_type: str = "int8"):
    model_path = Path(model_size).expanduser()
    if not model_path.is_absolute() and (Path(__file__).resolve().parent / model_path).exists():
        model_size = str(Path(__file__).resolve().parent / model_path)

    logging.info(
        "正在加载模型: %s (device=cpu, compute_type=%s, cpu_threads=%d)",
        model_size,
        compute_type,
        cpu_threads,
    )
    start = time.time()
    model = WhisperModel(
        model_size,
        device="cpu",
        compute_type=compute_type,
        cpu_threads=cpu_threads,
    )
    logging.info("模型加载完成，耗时 %.2fs", time.time() - start)
    return model


def transcribe_chunk(model: WhisperModel, audio: np.ndarray, language: str):
    """转录一段音频，返回文本与耗时。"""
    start = time.time()
    segments, _ = model.transcribe(
        audio,
        language=language,
        beam_size=1,
        best_of=1,
        patience=1,
        condition_on_previous_text=False,
        initial_prompt=INITIAL_PROMPT,
        without_timestamps=True,
        vad_filter=False,
        suppress_blank=True,
        temperature=0.0,
        compression_ratio_threshold=2.4,
        log_prob_threshold=MIN_AVG_LOGPROB,
        no_speech_threshold=MAX_NO_SPEECH_PROB,
    )
    texts = []
    for seg in segments:
        text = seg.text.strip()
        if not text:
            continue
        if seg.avg_logprob < MIN_AVG_LOGPROB:
            logging.debug("丢弃低置信片段: %s avg_logprob=%.2f", text, seg.avg_logprob)
            continue
        if seg.no_speech_prob > MAX_NO_SPEECH_PROB:
            logging.debug(
                "丢弃疑似静音片段: %s no_speech_prob=%.2f",
                text,
                seg.no_speech_prob,
            )
            continue
        texts.append(text)
    elapsed = time.time() - start
    return texts, elapsed


def emit_result(text: str, json_output: bool = False, role: str = ""):
    """输出识别结果。json_output 为 True 时输出 JSON 行，便于 Go 解析。"""
    if json_output:
        import json
        payload = {"timestamp": time.strftime("%H:%M:%S"), "text": text}
        if role:
            payload["role"] = role
        print(json.dumps(payload, ensure_ascii=False), flush=True)
    else:
        prefix = f"[{role}] " if role else ""
        print(f"[{time.strftime('%H:%M:%S')}] {prefix}{text}")
        sys.stdout.flush()


def create_vad_options():
    return VadOptions(
        threshold=0.5,
        neg_threshold=0.35,
        min_speech_duration_ms=MIN_SPEECH_DURATION_MS,
        min_silence_duration_ms=MIN_SILENCE_DURATION_MS,
        speech_pad_ms=SPEECH_PAD_MS,
        max_speech_duration_s=MAX_BUFFER_SECONDS,
    )


def normalize_device_name(name: str) -> str:
    return re.sub(r"[\s\\-_/()（）\\[\\].,]+", "", name.lower())


def is_macos() -> bool:
    return platform.system().lower() == "darwin"


def system_audio_source_path() -> Path:
    return Path(__file__).resolve().with_name("system_audio_capture.swift")


def system_audio_supported() -> bool:
    return is_macos() and system_audio_source_path().exists()


def build_system_audio_helper() -> str:
    source = system_audio_source_path()
    if not source.exists():
        raise RuntimeError(f"未找到系统音频采集脚本: {source}")

    cache_dir = Path.home() / ".cache" / "interview-assistant"
    cache_dir.mkdir(parents=True, exist_ok=True)
    helper = cache_dir / "system_audio_capture"

    needs_build = not helper.exists()
    if not needs_build:
        needs_build = source.stat().st_mtime > helper.stat().st_mtime

    if needs_build:
        logging.info("正在编译系统音频采集助手: %s", helper)
        result = subprocess.run(
            ["swiftc", str(source), "-o", str(helper)],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        )
        if result.returncode != 0:
            raise RuntimeError(
                "编译系统音频采集助手失败: "
                + (result.stderr.strip() or result.stdout.strip())
            )
    return str(helper)


class SystemAudioReader:
    def __init__(self, on_audio):
        self.on_audio = on_audio
        self.proc = None
        self.threads = []
        self.running = False

    def start(self):
        helper = build_system_audio_helper()
        self.proc = subprocess.Popen(
            [helper],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            bufsize=0,
        )
        self.running = True
        self.threads = [
            threading.Thread(target=self._read_stdout, daemon=True),
            threading.Thread(target=self._read_stderr, daemon=True),
        ]
        for thread in self.threads:
            thread.start()

    def stop(self):
        self.running = False
        if self.proc and self.proc.poll() is None:
            self.proc.terminate()
            try:
                self.proc.wait(timeout=2)
            except subprocess.TimeoutExpired:
                self.proc.kill()
        for thread in self.threads:
            thread.join(timeout=1)

    def _read_stdout(self):
        assert self.proc is not None and self.proc.stdout is not None
        pending = b""
        while self.running:
            data = pending + self.proc.stdout.read(CHUNK_SIZE * 4)
            if not data:
                break
            usable_len = (len(data) // 4) * 4
            pending = data[usable_len:]
            if usable_len == 0:
                continue
            data = data[:usable_len]
            audio = np.frombuffer(data, dtype=np.float32).copy()
            if audio.size:
                self.on_audio(audio)

    def _read_stderr(self):
        assert self.proc is not None and self.proc.stderr is not None
        for raw in self.proc.stderr:
            line = raw.decode("utf-8", errors="replace").strip()
            if line:
                logging.info(line)


class RealtimeTranscriber:
    def __init__(
        self,
        model: WhisperModel,
        input_device: int | None,
        language: str,
        source: str = "input",
        channels: int = 1,
        source_description: str = "",
        max_buffer_seconds: int = MAX_BUFFER_SECONDS,
        json_output: bool = False,
        role: str = "",
    ):
        self.model = model
        self.input_device = input_device
        self.language = language
        self.source = source
        self.channels = channels
        self.source_description = source_description
        self.max_buffer_samples = int(max_buffer_seconds * SAMPLE_RATE)
        self.buffer = np.zeros(0, dtype=np.float32)
        self.lock = threading.Lock()
        self.running = True
        self.vad_options = create_vad_options()
        self.last_text = ""
        self.json_output = json_output
        self.role = role
        self.last_audio_log = time.time()

    def append_audio(self, mono: np.ndarray):
        with self.lock:
            self.buffer = np.concatenate([self.buffer, mono])
            if len(self.buffer) > self.max_buffer_samples:
                self.buffer = self.buffer[-self.max_buffer_samples :]

    def audio_callback(self, indata, frames, time_info, status):
        if status:
            logging.warning("音频流状态: %s", status)
        audio = indata.astype(np.float32, copy=False)
        if audio.ndim == 1:
            mono = audio.copy()
        else:
            mono = audio.mean(axis=1).copy()
        self.append_audio(mono)

    def process_loop(self):
        processed_pos = 0
        while self.running:
            time.sleep(0.15)
            with self.lock:
                audio = self.buffer.copy()

            if len(audio) - processed_pos < SAMPLE_RATE * 0.5:
                continue

            region = audio[processed_pos:]
            if len(region) < SAMPLE_RATE * 0.3:
                continue

            speech_chunks = get_speech_timestamps(
                region,
                self.vad_options,
                sampling_rate=SAMPLE_RATE,
            )

            if not speech_chunks:
                processed_pos = max(0, len(audio) - int(SAMPLE_RATE * 0.5))
                continue

            # 只转录已经说完的语音段（后面出现了足够静音）
            ended_chunks = []
            for chunk in speech_chunks:
                is_last = chunk["end"] >= len(region) - int(SAMPLE_RATE * 0.25)
                if not is_last:
                    ended_chunks.append(chunk)

            if not ended_chunks:
                continue

            last_end = 0
            for chunk in ended_chunks:
                start = chunk["start"]
                end = chunk["end"]
                last_end = max(last_end, end)
                speech_audio = region[start:end]
                if len(speech_audio) < SAMPLE_RATE * 0.2:
                    continue
                rms = float(np.sqrt(np.mean(np.square(speech_audio))))
                peak = float(np.max(np.abs(speech_audio)))
                if rms < MIN_TRANSCRIBE_RMS and peak < MIN_TRANSCRIBE_PEAK:
                    logging.debug(
                        "跳过低能量片段: rms=%.5f peak=%.5f duration=%.2fs",
                        rms,
                        peak,
                        len(speech_audio) / SAMPLE_RATE,
                    )
                    continue

                texts, elapsed = transcribe_chunk(
                    self.model, speech_audio, self.language
                )
                if texts:
                    line = " ".join(texts)
                    if line != self.last_text:
                        emit_result(line, self.json_output, self.role)
                        self.last_text = line
                logging.debug(
                    "片段 %.2fs-%.2fs 转录耗时 %.2fs",
                    start / SAMPLE_RATE,
                    end / SAMPLE_RATE,
                    elapsed,
                )

            processed_pos += last_end
            with self.lock:
                if processed_pos > len(self.buffer) * 0.8:
                    self.buffer = self.buffer[processed_pos:]
                    processed_pos = 0

    def run(self):
        print(
            f"\n开始实时转录：source={self.source_description or self.source}, "
            f"language={self.language}"
        )
        print("请说话（按 Ctrl+C 停止）...\n")

        processor = threading.Thread(target=self.process_loop, daemon=True)
        processor_started = False

        try:
            if self.source == "system-audio":
                reader = SystemAudioReader(self.append_audio)
                reader.start()
                processor.start()
                processor_started = True
                try:
                    while self.running:
                        if reader.proc and reader.proc.poll() is not None:
                            raise RuntimeError(
                                f"系统音频采集助手已退出，exit={reader.proc.returncode}"
                            )
                        time.sleep(0.1)
                finally:
                    reader.stop()
            else:
                stream = sd.InputStream(
                    samplerate=SAMPLE_RATE,
                    channels=self.channels,
                    dtype=np.float32,
                    blocksize=CHUNK_SIZE,
                    device=self.input_device,
                    callback=self.audio_callback,
                )
                with stream:
                    processor.start()
                    processor_started = True
                    while self.running:
                        time.sleep(0.1)
        except KeyboardInterrupt:
            print("\n停止转录。")
        finally:
            self.running = False
            if processor_started:
                processor.join(timeout=2)


def parse_args():
    parser = argparse.ArgumentParser(
        description="macOS 本地实时语音转文字（faster-whisper）"
    )
    parser.add_argument(
        "--model",
        type=str,
        default="./models/large-v3-turbo",
        help="模型路径或名称，默认 ./models/large-v3-turbo（中文推荐）。可选 ./models/small/base 等",
    )
    parser.add_argument(
        "--device",
        type=int,
        default=None,
        help="音频输入设备编号（默认自动选择 MacBook Pro 麦克风）",
    )
    parser.add_argument(
        "--device-name",
        type=str,
        default=None,
        help="按名称匹配音频输入设备（优先于 --device）",
    )
    parser.add_argument(
        "--source",
        type=str,
        default="auto",
        choices=["auto", "input", "output", "system-audio"],
        help="音频来源：auto 按角色自动选择；input 录输入设备；output 录默认输出的可回采设备；system-audio 使用 macOS 系统音频采集",
    )
    parser.add_argument(
        "--language",
        type=str,
        default="zh",
        help="识别语言代码，默认 zh（中文）。设为 auto 则自动检测（首次较慢）",
    )
    parser.add_argument(
        "--compute-type",
        type=str,
        default="int8",
        choices=["int8", "float16", "float32"],
        help="计算精度，默认 int8（最快）",
    )
    parser.add_argument(
        "--cpu-threads",
        type=int,
        default=8,
        help="CPU 线程数，默认 8（Apple Silicon 建议性能核心数）",
    )
    parser.add_argument(
        "--list-devices", action="store_true", help="列出音频设备后退出"
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="只解析并打印将使用的音频来源，不加载模型、不开始转录",
    )
    parser.add_argument("--debug", action="store_true", help="开启调试日志")
    parser.add_argument(
        "--json-output",
        action="store_true",
        help="以 JSON 行格式输出识别结果，便于被其他程序解析",
    )
    parser.add_argument(
        "--role",
        type=str,
        default="",
        help="标识该音频流的角色，例如 interviewer 或 interviewee",
    )
    return parser.parse_args()


def select_input_device_by_name(device_name: str):
    devices = sd.query_devices()
    target = normalize_device_name(device_name)
    for i, dev in enumerate(devices):
        name = normalize_device_name(dev["name"])
        if dev["max_input_channels"] > 0 and target in name:
            return i
    return None


def select_default_input_device():
    default_input = sd.default.device[0]
    if default_input is None or default_input < 0:
        raise RuntimeError("系统没有默认输入设备，请先在系统设置中选择麦克风")
    devices = sd.query_devices()
    if devices[default_input]["max_input_channels"] <= 0:
        raise RuntimeError(
            f"系统默认输入设备不可录音: {devices[default_input]['name']}"
        )
    return default_input


def select_input_for_default_output():
    devices = sd.query_devices()
    default_output = sd.default.device[1]
    if default_output is None or default_output < 0:
        raise RuntimeError("系统没有默认输出设备，请先在系统设置中选择扬声器/耳机")

    output_dev = devices[default_output]
    if output_dev["max_input_channels"] > 0:
        return default_output

    output_name = normalize_device_name(output_dev["name"])
    for i, dev in enumerate(devices):
        if dev["max_input_channels"] <= 0:
            continue
        input_name = normalize_device_name(dev["name"])
        if input_name == output_name or output_name in input_name or input_name in output_name:
            return i

    virtual_keywords = [
        "orayvirtualaudiodevice",
        "blackhole",
        "soundflower",
        "loopback",
        "virtualaudio",
        "aggregate",
        "multioutput",
        "多输出",
    ]
    for i, dev in enumerate(devices):
        name = normalize_device_name(dev["name"])
        if dev["max_input_channels"] > 0 and any(k in name for k in virtual_keywords):
            logging.warning(
                "系统默认输出 '%s' 不是可回采设备，暂用虚拟输入 '%s'。"
                "若会议声音未路由到该虚拟设备，面试官音频仍无法被录到。",
                output_dev["name"],
                dev["name"],
            )
            return i

    raise RuntimeError(
        "当前系统默认输出设备不可直接录制: "
        f"{output_dev['name']}。请使用 system-audio 来源，或把会议/系统输出切到 "
        "BlackHole/OrayVirtualAudioDevice/多输出设备。"
    )


def input_device_channels(device_index: int) -> int:
    dev = sd.query_devices()[device_index]
    return max(1, min(int(dev["max_input_channels"]), 2))


def resolve_audio_source(args, devices):
    role = (args.role or "").strip().lower()
    source = args.source
    source_description = ""

    if source == "auto":
        if role == "interviewer" and system_audio_supported():
            return {
                "source": "system-audio",
                "input_device": None,
                "channels": 1,
                "description": "macOS system audio",
            }
        source = "output" if role == "interviewer" else "input"

    if source == "system-audio":
        if not system_audio_supported():
            raise RuntimeError("当前系统不支持 system-audio 来源")
        return {
            "source": "system-audio",
            "input_device": None,
            "channels": 1,
            "description": "macOS system audio",
        }

    input_device = args.device
    if args.device_name:
        device_name = args.device_name.strip()
        lowered = normalize_device_name(device_name)
        if lowered in {"default", "systeminput", "defaultinput", "mic", "microphone"}:
            input_device = select_default_input_device()
        elif lowered in {"systemoutput", "defaultoutput", "speaker", "speakers", "output"}:
            input_device = select_input_for_default_output()
        else:
            matched = select_input_device_by_name(device_name)
            if matched is None:
                if role == "interviewer":
                    raise RuntimeError(
                        f"未找到面试官音源设备 '{device_name}'，不会回退到麦克风。"
                    )
                logging.warning(
                    "未找到名称包含 '%s' 的输入设备，改用系统默认输入", device_name
                )
                matched = select_default_input_device()
            input_device = matched
    elif input_device is None:
        if source == "output":
            input_device = select_input_for_default_output()
        else:
            input_device = select_default_input_device()

    if input_device is None or input_device < 0 or input_device >= len(devices):
        raise RuntimeError(f"无效的音频输入设备编号: {input_device}")
    if devices[input_device]["max_input_channels"] <= 0:
        raise RuntimeError(f"设备不可录音: {input_device} - {devices[input_device]['name']}")

    channels = input_device_channels(input_device)
    source_description = f"{input_device} - {devices[input_device]['name']}"
    return {
        "source": "input",
        "input_device": input_device,
        "channels": channels,
        "description": source_description,
    }


def main():
    args = parse_args()
    setup_logging(args.debug)

    devices = list_audio_devices()

    if args.list_devices:
        return

    try:
        audio_source = resolve_audio_source(args, devices)
    except Exception as exc:
        print(f"音频来源解析失败: {exc}", file=sys.stderr)
        sys.exit(1)

    print(
        "音频来源绑定: "
        f"role={args.role or 'default'}, "
        f"source={audio_source['source']}, "
        f"target={audio_source['description']}"
    )

    if args.dry_run:
        return

    if args.language == "auto":
        language = None
    else:
        language = args.language

    model = load_model(args.model, args.cpu_threads, args.compute_type)
    transcriber = RealtimeTranscriber(
        model=model,
        input_device=audio_source["input_device"],
        language=language,
        source=audio_source["source"],
        channels=audio_source["channels"],
        source_description=audio_source["description"],
        json_output=args.json_output,
        role=args.role,
    )
    transcriber.run()


if __name__ == "__main__":
    main()
