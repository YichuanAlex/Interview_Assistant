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
import sys
import threading
import time
from collections import deque

import numpy as np
import sounddevice as sd

from faster_whisper import WhisperModel
from faster_whisper.vad import VadOptions, get_speech_timestamps


SAMPLE_RATE = 16000
CHUNK_SIZE = 1024  # 64ms @ 16kHz
MIN_SPEECH_DURATION_MS = 250
MIN_SILENCE_DURATION_MS = 350
SPEECH_PAD_MS = 200
MAX_BUFFER_SECONDS = 15


def setup_logging(debug: bool):
    level = logging.DEBUG if debug else logging.INFO
    logging.basicConfig(
        level=level,
        format="%(asctime)s [%(levelname)s] %(message)s",
        datefmt="%H:%M:%S",
    )


def list_audio_devices():
    devices = sd.query_devices()
    print("\n可用音频输入设备：")
    for i, dev in enumerate(devices):
        if dev["max_input_channels"] > 0:
            print(f"  {i}: {dev['name']} (channels={dev['max_input_channels']})")
    print()
    return devices


def load_model(model_size: str, cpu_threads: int, compute_type: str = "int8"):
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
        condition_on_previous_text=True,
        without_timestamps=True,
        vad_filter=False,
        suppress_blank=True,
        temperature=0.0,
    )
    texts = [seg.text.strip() for seg in segments if seg.text.strip()]
    elapsed = time.time() - start
    return texts, elapsed


def emit_result(text: str, json_output: bool = False):
    """输出识别结果。json_output 为 True 时输出 JSON 行，便于 Go 解析。"""
    if json_output:
        import json
        print(json.dumps({"timestamp": time.strftime("%H:%M:%S"), "text": text}, ensure_ascii=False), flush=True)
    else:
        print(f"[{time.strftime('%H:%M:%S')}] {text}")
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


class RealtimeTranscriber:
    def __init__(
        self,
        model: WhisperModel,
        input_device: int,
        language: str,
        max_buffer_seconds: int = MAX_BUFFER_SECONDS,
        json_output: bool = False,
    ):
        self.model = model
        self.input_device = input_device
        self.language = language
        self.max_buffer_samples = int(max_buffer_seconds * SAMPLE_RATE)
        self.buffer = np.zeros(0, dtype=np.float32)
        self.lock = threading.Lock()
        self.running = True
        self.vad_options = create_vad_options()
        self.last_text = ""
        self.json_output = json_output

    def audio_callback(self, indata, frames, time_info, status):
        if status:
            logging.warning("音频流状态: %s", status)
        mono = indata[:, 0].astype(np.float32).copy()
        with self.lock:
            self.buffer = np.concatenate([self.buffer, mono])
            if len(self.buffer) > self.max_buffer_samples:
                self.buffer = self.buffer[-self.max_buffer_samples :]

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
                is_last = chunk["end"] >= len(region) - int(SAMPLE_RATE * 0.4)
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

                texts, elapsed = transcribe_chunk(
                    self.model, speech_audio, self.language
                )
                if texts:
                    line = " ".join(texts)
                    if line != self.last_text:
                        emit_result(line, self.json_output)
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
            f"\n开始实时转录：device={self.input_device}, language={self.language}"
        )
        print("请说话（按 Ctrl+C 停止）...\n")

        stream = sd.InputStream(
            samplerate=SAMPLE_RATE,
            channels=1,
            dtype=np.float32,
            blocksize=CHUNK_SIZE,
            device=self.input_device,
            callback=self.audio_callback,
        )

        processor = threading.Thread(target=self.process_loop, daemon=True)

        try:
            with stream:
                processor.start()
                while self.running:
                    time.sleep(0.1)
        except KeyboardInterrupt:
            print("\n停止转录。")
        finally:
            self.running = False
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
    parser.add_argument("--debug", action="store_true", help="开启调试日志")
    parser.add_argument(
        "--json-output",
        action="store_true",
        help="以 JSON 行格式输出识别结果，便于被其他程序解析",
    )
    return parser.parse_args()


def select_default_input_device():
    devices = sd.query_devices()
    for i, dev in enumerate(devices):
        name = dev["name"].lower()
        if dev["max_input_channels"] > 0 and (
            "microphone" in name or "麦克风" in name
        ):
            return i
    return sd.default.device[0]


def main():
    args = parse_args()
    setup_logging(args.debug)

    devices = list_audio_devices()

    if args.list_devices:
        return

    input_device = args.device
    if input_device is None:
        input_device = select_default_input_device()
        print(f"自动选择输入设备: {input_device} - {devices[input_device]['name']}")
    else:
        print(f"使用输入设备: {input_device} - {devices[input_device]['name']}")

    if args.language == "auto":
        language = None
    else:
        language = args.language

    model = load_model(args.model, args.cpu_threads, args.compute_type)
    transcriber = RealtimeTranscriber(
        model=model,
        input_device=input_device,
        language=language,
        json_output=args.json_output,
    )
    transcriber.run()


if __name__ == "__main__":
    main()
