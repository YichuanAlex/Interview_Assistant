import AVFoundation
import CoreMedia
import Foundation
import ScreenCaptureKit

final class SystemAudioCapture: NSObject, SCStreamOutput, SCStreamDelegate {
    private let queue = DispatchQueue(label: "interview-assistant.system-audio")
    private let targetFormat = AVAudioFormat(
        commonFormat: .pcmFormatFloat32,
        sampleRate: 16_000,
        channels: 1,
        interleaved: false
    )!
    private var converters: [String: AVAudioConverter] = [:]
    private let stderr = FileHandle.standardError
    private let stdout = FileHandle.standardOutput

    func start() async throws {
        let content = try await SCShareableContent.excludingDesktopWindows(
            false,
            onScreenWindowsOnly: true
        )
        guard let display = content.displays.first else {
            throw NSError(
                domain: "SystemAudioCapture",
                code: 1,
                userInfo: [NSLocalizedDescriptionKey: "No display is available for ScreenCaptureKit."]
            )
        }

        let filter = SCContentFilter(display: display, excludingWindows: [])
        let config = SCStreamConfiguration()
        config.width = 2
        config.height = 2
        config.minimumFrameInterval = CMTime(value: 1, timescale: 5)
        config.queueDepth = 3
        config.capturesAudio = true
        config.excludesCurrentProcessAudio = true
        config.sampleRate = 48_000
        config.channelCount = 2

        let stream = SCStream(filter: filter, configuration: config, delegate: self)
        try stream.addStreamOutput(self, type: .audio, sampleHandlerQueue: queue)
        try await stream.startCapture()
        log("system audio capture started")
        while true {
            try await Task.sleep(nanoseconds: 1_000_000_000)
        }
    }

    func stream(
        _ stream: SCStream,
        didOutputSampleBuffer sampleBuffer: CMSampleBuffer,
        of outputType: SCStreamOutputType
    ) {
        guard outputType == .audio, sampleBuffer.isValid else {
            return
        }
        guard let formatDescription = CMSampleBufferGetFormatDescription(sampleBuffer) else {
            return
        }
        let sourceFormat = AVAudioFormat(cmAudioFormatDescription: formatDescription)

        let frameCount = AVAudioFrameCount(CMSampleBufferGetNumSamples(sampleBuffer))
        guard frameCount > 0,
              let sourceBuffer = AVAudioPCMBuffer(
                pcmFormat: sourceFormat,
                frameCapacity: frameCount
              ) else {
            return
        }
        sourceBuffer.frameLength = frameCount

        let copyStatus = CMSampleBufferCopyPCMDataIntoAudioBufferList(
            sampleBuffer,
            at: 0,
            frameCount: Int32(frameCount),
            into: sourceBuffer.mutableAudioBufferList
        )
        guard copyStatus == noErr else {
            log("failed to copy PCM data: \(copyStatus)")
            return
        }

        guard let outputBuffer = convertToTargetFormat(sourceBuffer, sourceFormat: sourceFormat),
              let channel = outputBuffer.floatChannelData?[0] else {
            return
        }

        let byteCount = Int(outputBuffer.frameLength) * MemoryLayout<Float>.size
        let data = Data(bytes: channel, count: byteCount)
        stdout.write(data)
    }

    func stream(_ stream: SCStream, didStopWithError error: Error) {
        log("system audio capture stopped: \(error.localizedDescription)")
        exit(2)
    }

    private func convertToTargetFormat(
        _ sourceBuffer: AVAudioPCMBuffer,
        sourceFormat: AVAudioFormat
    ) -> AVAudioPCMBuffer? {
        let key = "\(sourceFormat.sampleRate)-\(sourceFormat.channelCount)-\(sourceFormat.commonFormat.rawValue)-\(sourceFormat.isInterleaved)"
        let converter: AVAudioConverter
        if let existing = converters[key] {
            converter = existing
        } else if let created = AVAudioConverter(from: sourceFormat, to: targetFormat) {
            converters[key] = created
            converter = created
        } else {
            log("failed to create audio converter")
            return nil
        }

        let ratio = targetFormat.sampleRate / sourceFormat.sampleRate
        let capacity = max(1, AVAudioFrameCount(Double(sourceBuffer.frameLength) * ratio) + 32)
        guard let outputBuffer = AVAudioPCMBuffer(
            pcmFormat: targetFormat,
            frameCapacity: capacity
        ) else {
            return nil
        }

        var didProvideInput = false
        var error: NSError?
        let status = converter.convert(to: outputBuffer, error: &error) { _, outStatus in
            if didProvideInput {
                outStatus.pointee = .noDataNow
                return nil
            }
            didProvideInput = true
            outStatus.pointee = .haveData
            return sourceBuffer
        }

        if status == .error {
            log("audio conversion failed: \(error?.localizedDescription ?? "unknown error")")
            return nil
        }
        return outputBuffer
    }

    private func log(_ message: String) {
        if let data = "[system-audio] \(message)\n".data(using: .utf8) {
            stderr.write(data)
        }
    }
}

let capture = SystemAudioCapture()
do {
    try await capture.start()
} catch {
    let message = "[system-audio] \(error.localizedDescription)\n"
    FileHandle.standardError.write(Data(message.utf8))
    exit(1)
}
