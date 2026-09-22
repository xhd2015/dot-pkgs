// RecognizeText.swift — Apple Vision OCR helper for go-pkgs/ocr.
// Invoked as: vision-ocr <image-path>
// stdout: recognized lines (one per line); stderr: diagnostics
import AppKit
import Foundation
import Vision

guard CommandLine.arguments.count >= 2 else {
    fputs("usage: vision-ocr <image-path>\n", stderr)
    exit(2)
}

let path = CommandLine.arguments[1]
guard FileManager.default.fileExists(atPath: path) else {
    fputs("ocr: image not found: \(path)\n", stderr)
    exit(1)
}

guard let img = NSImage(contentsOfFile: path),
      let tiff = img.tiffRepresentation,
      let rep = NSBitmapImageRep(data: tiff),
      let cgImage = rep.cgImage
else {
    fputs("ocr: failed to load image: \(path)\n", stderr)
    exit(1)
}

let request = VNRecognizeTextRequest()
request.recognitionLevel = .accurate
request.usesLanguageCorrection = true
if #available(macOS 13.0, *) {
    request.recognitionLanguages = ["en-US", "zh-Hans", "zh-Hant"]
    request.automaticallyDetectsLanguage = true
}

do {
    let handler = VNImageRequestHandler(cgImage: cgImage, options: [:])
    try handler.perform([request])
} catch {
    fputs("ocr: recognize failed: \(error.localizedDescription)\n", stderr)
    exit(1)
}

let observations = request.results ?? []
for obs in observations {
    if let candidate = obs.topCandidates(1).first {
        print(candidate.string)
    }
}
