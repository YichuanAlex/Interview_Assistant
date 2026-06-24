#!/usr/bin/env swift
import Foundation
import Vision
import AppKit

@available(macOS 10.15, *)
func recognizeText(in imagePath: String) -> String {
    guard let image = NSImage(contentsOfFile: imagePath),
          let cgImage = image.cgImage(forProposedRect: nil, context: nil, hints: nil) else {
        return ""
    }

    let request = VNRecognizeTextRequest()
    request.recognitionLanguages = ["zh-Hans", "en-US"]
    request.usesLanguageCorrection = true

    let handler = VNImageRequestHandler(cgImage: cgImage, options: [:])
    do {
        try handler.perform([request])
    } catch {
        return ""
    }

    guard let observations = request.results as? [VNRecognizedTextObservation] else {
        return ""
    }

    let lines = observations.compactMap { obs -> String? in
        guard let top = obs.topCandidates(1).first else { return nil }
        return top.string
    }

    return lines.joined(separator: "\n")
}

if CommandLine.arguments.count < 2 {
    exit(1)
}

let imagePath = CommandLine.arguments[1]
if #available(macOS 10.15, *) {
    let text = recognizeText(in: imagePath)
    print(text)
} else {
    print("")
}
