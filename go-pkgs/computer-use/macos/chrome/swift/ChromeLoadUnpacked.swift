// ChromeLoadUnpacked.swift — in-process Chrome Load unpacked for a GUI host.
//
// Source of truth: go-pkgs/computer-use/macos/chrome/swift/ChromeLoadUnpacked.swift
// in github.com/xhd2015/dot-pkgs. GUI hosts (Marcus.app) copy this file via
// scripts/marcus-macos-app/sync-chrome-load-unpacked.sh — edit here, not the copy.
//
// Port of LoadUnpacked (chrome.go, load_darwin.go, ax_darwin.go, remove_darwin.go).
// Must run inside the GUI process after a user click so Accessibility and Apple
// Events attach to that app, not a helper, daemon, or spawned CLI.

import AppKit
import ApplicationServices
import Foundation

/// In-process Chrome UI: Developer mode → Load unpacked → folder.
public enum ChromeLoadUnpacked {
    public static let extensionsURL = "chrome://extensions"
    public static let defaultAppName = "Google Chrome"
    public static let defaultVerifyName = "Browser Agent"
    public static let chromeBundleID = "com.google.Chrome"

    public struct Options {
        public var extensionDir: String
        public var appName: String
        public var verifyName: String
        public var verifyVersion: String
        public var keepOlder: Bool
        public var pageTimeout: TimeInterval
        public var dialogTimeout: TimeInterval
        public var verifyTimeout: TimeInterval
        public var onLog: ((String) -> Void)?

        public init(
            extensionDir: String,
            appName: String = ChromeLoadUnpacked.defaultAppName,
            verifyName: String = ChromeLoadUnpacked.defaultVerifyName,
            verifyVersion: String = "",
            keepOlder: Bool = false,
            pageTimeout: TimeInterval = 20,
            dialogTimeout: TimeInterval = 10,
            verifyTimeout: TimeInterval = 8,
            onLog: ((String) -> Void)? = nil
        ) {
            self.extensionDir = extensionDir
            self.appName = appName
            self.verifyName = verifyName
            self.verifyVersion = verifyVersion
            self.keepOlder = keepOlder
            self.pageTimeout = pageTimeout
            self.dialogTimeout = dialogTimeout
            self.verifyTimeout = verifyTimeout
            self.onLog = onLog
        }
    }

    public struct Result: Sendable, Equatable {
        public var developerModeVisible = false
        public var loadUnpackedVisible = false
        public var extensionListed = false
        public var loaded = false
        public var submittedUnknown = false
        public var multiInstanceWarned = false
        public var removedOlder = 0
        public var removeOlderAttempted = false

        public init() {}
    }

    public struct LoadError: Error, LocalizedError, Equatable {
        public var message: String
        public var errorDescription: String? { message }
        public init(_ message: String) { self.message = message }
    }

    public static let systemEventsBundleID = "com.apple.systemevents"

    /// Automation / AX consent for one target (no sheet unless `prompt` is true).
    public enum AutomationConsent: String, Equatable, Sendable {
        case granted
        case denied
        case notDetermined
        case appNotRunning
        case unknown
    }

    public enum PermissionTarget: String, Equatable, Sendable {
        case accessibility = "Accessibility"
        case googleChrome = "Google Chrome"
        case systemEvents = "System Events"
    }

    public struct PermissionProbe: Equatable, Sendable {
        public var accessibilityTrusted: Bool
        public var chrome: AutomationConsent
        public var systemEvents: AutomationConsent

        public init(
            accessibilityTrusted: Bool,
            chrome: AutomationConsent,
            systemEvents: AutomationConsent
        ) {
            self.accessibilityTrusted = accessibilityTrusted
            self.chrome = chrome
            self.systemEvents = systemEvents
        }

        /// `appNotRunning` is not a permission miss — the process can be launched.
        public var allGranted: Bool {
            accessibilityTrusted && Self.automationOK(chrome) && Self.automationOK(systemEvents)
        }

        /// First missing permission. Accessibility is notDetermined when untrusted
        /// (TCC does not distinguish AX deny vs never-asked).
        public var blocker: (target: PermissionTarget, consent: AutomationConsent)? {
            if !accessibilityTrusted {
                return (.accessibility, .notDetermined)
            }
            if !Self.automationOK(chrome) {
                return (.googleChrome, chrome)
            }
            if !Self.automationOK(systemEvents) {
                return (.systemEvents, systemEvents)
            }
            return nil
        }

        public static func automationOK(_ c: AutomationConsent) -> Bool {
            c == .granted || c == .appNotRunning
        }
    }

    public static func isProcessTrusted() -> Bool {
        AXIsProcessTrusted()
    }

    /// Shows the system Accessibility prompt when the process is not trusted.
    @discardableResult
    public static func requestTrust() -> Bool {
        let key = kAXTrustedCheckOptionPrompt.takeUnretainedValue() as String
        return AXIsProcessTrustedWithOptions([key: true] as CFDictionary)
    }

    /// Maps `AEDeterminePermissionToAutomateTarget` OSStatus (and AX-style codes).
    public static func mapAutomationStatus(_ status: OSStatus) -> AutomationConsent {
        switch status {
        case 0:
            return .granted
        case -1743: // errAEEventNotPermitted
            return .denied
        case -1744: // errAEEventWouldRequireUserConsent
            return .notDetermined
        case -600: // procNotFound
            return .appNotRunning
        default:
            return .unknown
        }
    }

    /// Probe one app's Automation permission. `prompt` may show the system sheet
    /// and block the calling thread until the user answers — do not call on main.
    public static func automationConsent(bundleID: String, prompt: Bool) -> AutomationConsent {
        let desc = NSAppleEventDescriptor(bundleIdentifier: bundleID)
        guard let ptr = desc.aeDesc else { return .unknown }
        let status = AEDeterminePermissionToAutomateTarget(
            ptr,
            typeWildCard,
            typeWildCard,
            prompt
        )
        return mapAutomationStatus(status)
    }

    /// No-prompt snapshot. Launch Chrome first so Chrome Automation is not `appNotRunning`.
    public static func probePermissions() -> PermissionProbe {
        PermissionProbe(
            accessibilityTrusted: isProcessTrusted(),
            chrome: automationConsent(bundleID: chromeBundleID, prompt: false),
            systemEvents: automationConsent(bundleID: systemEventsBundleID, prompt: false)
        )
    }

    /// System sheets for not-yet-determined Automation targets. Blocks the caller.
    public static func promptUndeterminedAutomation(_ probe: PermissionProbe) {
        if probe.chrome == .notDetermined || probe.chrome == .appNotRunning {
            _ = automationConsent(bundleID: chromeBundleID, prompt: true)
        }
        if probe.systemEvents == .notDetermined || probe.systemEvents == .appNotRunning {
            _ = automationConsent(bundleID: systemEventsBundleID, prompt: true)
        }
    }

    /// NSWorkspace launch (no Apple Events). Safe before the Automation probe.
    public static func launchChrome() throws {
        guard let url = NSWorkspace.shared.urlForApplication(withBundleIdentifier: chromeBundleID) else {
            throw LoadError("chrome: Google Chrome is not installed")
        }
        let cfg = NSWorkspace.OpenConfiguration()
        cfg.activates = true
        let box = WaitBox<Error?>()
        NSWorkspace.shared.openApplication(at: url, configuration: cfg) { _, err in
            box.finish(err)
        }
        if let err = box.wait(timeout: 30) ?? nil {
            throw LoadError("chrome: could not open Google Chrome: \(err.localizedDescription)")
        }
    }

    /// System Events is often not running; AppleScript starts it. Not a TCC miss.
    public static func launchSystemEvents() {
        guard let url = NSWorkspace.shared.urlForApplication(withBundleIdentifier: systemEventsBundleID) else {
            return
        }
        let cfg = NSWorkspace.OpenConfiguration()
        cfg.activates = false
        NSWorkspace.shared.openApplication(at: url, configuration: cfg)
    }

    /// True when `s` itself looks like an extension version (digits and dots).
    public static func isVersionLike(_ s: String) -> Bool {
        let t = s.trimmingCharacters(in: .whitespacesAndNewlines)
        return !t.isEmpty && inferVersion(fromDir: t) == t
    }

    public static func inferVersion(fromDir dir: String) -> String {
        var trimmed = dir.trimmingCharacters(in: .whitespacesAndNewlines)
        while trimmed.hasSuffix("/") && trimmed.count > 1 {
            trimmed.removeLast()
        }
        let base = (trimmed as NSString).lastPathComponent
        if base.isEmpty || base == "." || base == "/" { return "" }
        guard let first = base.unicodeScalars.first, first >= "0" && first <= "9" else {
            return ""
        }
        for s in base.unicodeScalars {
            if !((s >= "0" && s <= "9") || s == ".") { return "" }
        }
        return base
    }

    public static func extensionDirOK(_ path: String) -> Bool {
        let path = path.trimmingCharacters(in: .whitespacesAndNewlines)
        if path.isEmpty { return false }
        var isDir: ObjCBool = false
        guard FileManager.default.fileExists(atPath: path, isDirectory: &isDir), isDir.boolValue else {
            return false
        }
        let manifest = (path as NSString).appendingPathComponent("manifest.json")
        var isFile: ObjCBool = false
        guard FileManager.default.fileExists(atPath: manifest, isDirectory: &isFile), !isFile.boolValue else {
            return false
        }
        return true
    }

    public static func escapeAppleScript(_ s: String) -> String {
        s.replacingOccurrences(of: "\\", with: "\\\\")
            .replacingOccurrences(of: "\"", with: "\\\"")
    }

    /// Non-empty when more than one Chrome main process or a temp profile is running.
    public static func multiInstanceHint() -> String {
        let procs = chromeMainProcesses()
        if procs.isEmpty { return "" }
        let temp = procs.filter {
            $0.contains("user-data-dir=") && ($0.contains("/tmp/") || $0.contains("chrome-ext-test"))
        }.count
        if procs.count > 1 || temp > 0 {
            return "multiple Google Chrome processes detected; AppleScript may target a temp profile — quit temp Chromes (e.g. --user-data-dir=/tmp/…) before Load unpacked"
        }
        return ""
    }

    public static func load(options: Options) throws -> Result {
        try Loader(options: options).run()
    }

    public struct RemoveOlderOutcome: Equatable, Sendable {
        public var removed: Int
        public var cardCount: Int
        public init(removed: Int = 0, cardCount: Int = 0) {
            self.removed = removed
            self.cardCount = cardCount
        }
    }

    /// AX card count only (no clicks). For the standalone remove-older CLI.
    public static func scanExtensionCards(
        appName: String = defaultAppName,
        verifyName: String = defaultVerifyName
    ) -> Int {
        axCollectNamedCards(appName: appName, verifyName: verifyName).count
    }

    /// Remove older same-name cards only. Does not Load unpacked. Chrome should
    /// already be on chrome://extensions.
    public static func removeOlderCards(options: Options) throws -> RemoveOlderOutcome {
        try Loader(options: options).runRemoveOlderOnly()
    }

    /// Debug: print AX nodes whose title/desc/value mentions Cancel, Remove, or Dialog.
    public static func dumpInterestingAX(appName: String = defaultAppName) {
        if !isProcessTrusted() {
            print("dump  AX not trusted")
            return
        }
        guard let pid = try? chromePID(appName) else {
            print("dump  no Chrome pid")
            return
        }
        let app = AXUIElementCreateApplication(pid)
        var budget = 8000
        axDumpInteresting(app, depth: 28, budget: &budget, path: "app")
    }
}

// MARK: - Loader

private final class Loader {
    var options: ChromeLoadUnpacked.Options
    var result = ChromeLoadUnpacked.Result()

    init(options: ChromeLoadUnpacked.Options) {
        self.options = options
    }

    func run() throws -> ChromeLoadUnpacked.Result {
        try normalize()
        if let hint = nonempty(ChromeLoadUnpacked.multiInstanceHint()) {
            result.multiInstanceWarned = true
            step("warning: \(hint)")
        }
        step("  app       \(options.appName)")
        step("  extension \(options.extensionDir)")
        step("  mode      ui (no Chrome flags, default profile)")

        try launchChromeBare()
        step("  step      launch_chrome        ok  (bare, no flags)")
        sleep(2)

        let how = try ensureOnExtensionsPage()
        step("  step      open_extensions      ok  (\(ChromeLoadUnpacked.extensionsURL), \(how))")
        sleep(1.5)

        let probe = probeExtensionsUI()
        result.developerModeVisible = probe.dev
        result.loadUnpackedVisible = probe.loadBtn
        result.extensionListed = probe.listed
        if let err = probe.error {
            step("warning: could not probe extensions UI: \(err)")
        }
        step("  step      probe                dev=\(yn(probe.dev)) load_unpacked=\(yn(probe.loadBtn)) listed=\(yn(probe.listed))")

        try ensureDeveloperMode()
        let afterDev = probeExtensionsUI()
        result.loadUnpackedVisible = afterDev.loadBtn

        try clickLoadUnpacked()
        step("  step      load_unpacked        click")

        try waitOpenSheet(timeout: options.dialogTimeout)
        step("  step      open_dialog          ok")

        try pickFolderViaKeystrokes()
        step("  step      pick_folder          ok  (\(options.extensionDir))")
        sleep(1)

        let deadline = Date().addingTimeInterval(options.verifyTimeout)
        while Date() < deadline {
            if axExistsNamed(options.appName, options.verifyName) {
                result.extensionListed = true
                result.loaded = true
                step("  status    loaded  (\(options.verifyName) visible via AX)")
                maybeRemoveOlder()
                return result
            }
            let listed = probeExtensionsUI().listed
            if listed {
                result.extensionListed = true
                result.loaded = true
                step("  status    loaded  (\(options.verifyName) visible on extensions page)")
                maybeRemoveOlder()
                return result
            }
            sleep(0.35)
        }
        result.submittedUnknown = true
        step("warning: folder selected; could not confirm extension card on page")
        step("  status    submitted_unknown")
        maybeRemoveOlder()
        return result
    }

    func normalize() throws {
        let raw = options.extensionDir.trimmingCharacters(in: .whitespacesAndNewlines)
        if raw.isEmpty {
            throw ChromeLoadUnpacked.LoadError("chrome: ExtensionDir is required")
        }
        options.extensionDir = (raw as NSString).standardizingPath
        if !ChromeLoadUnpacked.extensionDirOK(options.extensionDir) {
            throw ChromeLoadUnpacked.LoadError(
                "chrome: extension dir missing or no manifest.json: \(options.extensionDir)"
            )
        }
        if options.appName.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            options.appName = ChromeLoadUnpacked.defaultAppName
        }
        if options.verifyName.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            options.verifyName = ChromeLoadUnpacked.defaultVerifyName
        }
        if options.verifyVersion.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            options.verifyVersion = ChromeLoadUnpacked.inferVersion(fromDir: options.extensionDir)
        }
        if options.pageTimeout <= 0 { options.pageTimeout = 20 }
        if options.dialogTimeout <= 0 { options.dialogTimeout = 10 }
        if options.verifyTimeout <= 0 { options.verifyTimeout = 8 }
    }

    func launchChromeBare() throws {
        if options.appName == ChromeLoadUnpacked.defaultAppName,
           let url = NSWorkspace.shared.urlForApplication(
               withBundleIdentifier: ChromeLoadUnpacked.chromeBundleID
           )
        {
            let cfg = NSWorkspace.OpenConfiguration()
            cfg.activates = true
            let box = WaitBox<Error?>()
            NSWorkspace.shared.openApplication(at: url, configuration: cfg) { _, err in
                box.finish(err)
            }
            if let err = box.wait(timeout: 30) ?? nil {
                throw ChromeLoadUnpacked.LoadError(
                    "chrome: could not open \"\(options.appName)\": \(err.localizedDescription)"
                )
            }
            return
        }
        let app = ChromeLoadUnpacked.escapeAppleScript(options.appName)
        _ = try appleScript("tell application \"\(app)\" to activate")
    }

    func ensureOnExtensionsPage() throws -> String {
        let app = ChromeLoadUnpacked.escapeAppleScript(options.appName)
        let url = ChromeLoadUnpacked.extensionsURL
        let script = """
        tell application "\(app)"
          activate
          set didCreate to false
          if (count of windows) = 0 then
            make new window
            set didCreate to true
          end if
          set u to ""
          try
            set u to URL of active tab of front window as text
          end try
          if u starts with "chrome://extensions" then
            return "reused"
          end if
          try
            tell front window
              set t to make new tab with properties {URL:"\(url)"}
              set active tab index to (index of t)
            end tell
            delay 0.7
          on error
            open location "\(url)"
            delay 0.7
          end try
          set u to ""
          try
            set u to URL of active tab of front window as text
          end try
          if u does not start with "chrome://extensions" then
            try
              set URL of active tab of front window to "\(url)"
              delay 0.6
              set u to URL of active tab of front window as text
            end try
          end if
          if u does not start with "chrome://extensions" then
            return "fail:" & u
          end if
          if didCreate then
            return "created"
          end if
          return "new-tab"
        end tell
        """
        let out = try appleScript(script)
        if out.hasPrefix("fail:") {
            throw ChromeLoadUnpacked.LoadError(
                "chrome: front tab is not \(url) (got \(out.dropFirst(5))) — refuse to click (avoids misclick on restored pages)"
            )
        }
        return out.isEmpty ? "reused" : out
    }

    func frontTabURL() -> String {
        let app = ChromeLoadUnpacked.escapeAppleScript(options.appName)
        let script = """
        tell application "\(app)"
          try
            return URL of active tab of front window as text
          on error
            return ""
          end try
        end tell
        """
        return (try? appleScript(script)) ?? ""
    }

    func requireExtensionsURL() throws {
        if frontTabURL().hasPrefix("chrome://extensions") { return }
        _ = try ensureOnExtensionsPage()
        let u = frontTabURL()
        if u.hasPrefix("chrome://extensions") { return }
        throw ChromeLoadUnpacked.LoadError(
            "chrome: expected front tab \(ChromeLoadUnpacked.extensionsURL), got \"\(u)\" (refusing clicks)"
        )
    }

    func probeExtensionsUI() -> (dev: Bool, loadBtn: Bool, listed: Bool, error: String?) {
        let app = ChromeLoadUnpacked.escapeAppleScript(options.appName)
        let name = ChromeLoadUnpacked.escapeAppleScript(options.verifyName)
        let script = """
        tell application "System Events"
          tell process "\(app)"
            set frontmost to true
            set devOn to false
            set loadOn to false
            set listedOn to false
            try
              if exists (first checkbox whose name is "Developer mode") then set devOn to true
            end try
            try
              if exists (first UI element whose name is "Developer mode") then set devOn to true
            end try
            try
              if exists (first button whose name is "Load unpacked") then set loadOn to true
            end try
            try
              if exists (first UI element whose name is "Load unpacked") then set loadOn to true
            end try
            try
              if exists (first UI element whose name is "\(name)") then set listedOn to true
            end try
            try
              if exists (first static text whose value contains "\(name)") then set listedOn to true
            end try
            return (devOn as text) & "," & (loadOn as text) & "," & (listedOn as text)
          end tell
        end tell
        """
        do {
            let out = try appleScript(script)
            var parts = out.split(separator: ",", omittingEmptySubsequences: false).map(String.init)
            while parts.count < 3 { parts.append("false") }
            return (parts[0] == "true", parts[1] == "true", parts[2] == "true", nil)
        } catch {
            return (false, false, false, error.localizedDescription)
        }
    }

    func ensureDeveloperMode() throws {
        try requireExtensionsURL()
        if axExistsNamed(options.appName, "Load unpacked") {
            step("  step      developer_mode       already_on  (AX Load unpacked)")
            return
        }
        if probeExtensionsUI().loadBtn {
            step("  step      developer_mode       already_on")
            return
        }
        do {
            try axClickNamed(options.appName, "Developer mode")
        } catch {
            throw ChromeLoadUnpacked.LoadError("chrome: enable Developer mode: \(error.localizedDescription)")
        }
        step("  step      developer_mode       toggled-ax")
        let deadline = Date().addingTimeInterval(10)
        while Date() < deadline {
            let u = frontTabURL()
            if !u.isEmpty && !u.hasPrefix("chrome://extensions") {
                throw ChromeLoadUnpacked.LoadError(
                    "chrome: left \(ChromeLoadUnpacked.extensionsURL) during developer mode (now \"\(u)\")"
                )
            }
            if axExistsNamed(options.appName, "Load unpacked") {
                step("  step      developer_mode       on  (AX Load unpacked visible)")
                return
            }
            if probeExtensionsUI().loadBtn {
                step("  step      developer_mode       on  (Load unpacked visible)")
                return
            }
            sleep(0.35)
        }
        throw ChromeLoadUnpacked.LoadError("chrome: Developer mode toggle did not reveal Load unpacked")
    }

    func clickLoadUnpacked() throws {
        try requireExtensionsURL()
        if (try? axClickNamed(options.appName, "Load unpacked")) != nil {
            return
        }
        let app = ChromeLoadUnpacked.escapeAppleScript(options.appName)
        let script = """
        tell application "System Events"
          tell process "\(app)"
            set frontmost to true
            delay 0.25
            try
              click (first button whose name is "Load unpacked")
              return "ok"
            end try
            try
              click (first UI element whose name is "Load unpacked")
              return "ok"
            end try
            return "err:Load unpacked button not found"
          end tell
        end tell
        """
        let out = try appleScript(script)
        if out.hasPrefix("err:") {
            throw ChromeLoadUnpacked.LoadError("chrome: click Load unpacked: \(out)")
        }
    }

    func waitOpenSheet(timeout: TimeInterval) throws {
        let app = ChromeLoadUnpacked.escapeAppleScript(options.appName)
        let deadline = Date().addingTimeInterval(timeout)
        let script = """
        tell application "System Events"
          tell process "\(app)"
            try
              return (count of sheets of front window) as text
            on error
              return "0"
            end try
          end tell
        end tell
        """
        while Date() < deadline {
            if let out = try? appleScript(script), !out.isEmpty, out != "0" {
                return
            }
            sleep(0.3)
        }
        throw ChromeLoadUnpacked.LoadError("chrome: open folder dialog did not appear after Load unpacked")
    }

    func pickFolderViaKeystrokes() throws {
        let app = ChromeLoadUnpacked.escapeAppleScript(options.appName)
        let path = ChromeLoadUnpacked.escapeAppleScript(options.extensionDir)
        let script = """
        tell application "\(app)" to activate
        delay 0.3
        tell application "System Events"
          tell process "\(app)"
            set frontmost to true
            delay 0.2
            keystroke "g" using {command down, shift down}
            delay 0.7
            keystroke "\(path)"
            delay 0.25
            keystroke return
            delay 0.9
            try
              click button "Open" of sheet 1 of front window
            on error
              try
                click button "Open" of front window
              on error
                keystroke return
              end try
            end try
          end tell
        end tell
        return "ok"
        """
        _ = try appleScript(script)
    }

    var lastCardCount = 0

    func runRemoveOlderOnly() throws -> ChromeLoadUnpacked.RemoveOlderOutcome {
        result.removeOlderAttempted = true
        step("  step      remove_older         scanning...")
        let n = try removeOlderExtensions()
        result.removedOlder = n
        if n > 0 {
            step("  step      remove_older         removed \(n) same-name card(s)")
        } else {
            step("  step      remove_older         none")
        }
        return ChromeLoadUnpacked.RemoveOlderOutcome(removed: n, cardCount: lastCardCount)
    }

    func maybeRemoveOlder() {
        if options.keepOlder { return }
        if !result.loaded && !result.submittedUnknown { return }
        result.removeOlderAttempted = true
        step("  step      remove_older         scanning...")
        do {
            let n = try removeOlderExtensions()
            result.removedOlder = n
            if n > 0 {
                step("  step      remove_older         removed \(n) same-name card(s)")
            } else {
                step("  step      remove_older         none")
            }
        } catch {
            step("warning: could not remove older extensions: \(error.localizedDescription)")
        }
    }

    /// AX only: a card is name → version → Remove, not every node named Browser Agent.
    /// Confirm is the in-page Cancel/Remove dialog (not a macOS sheet).
    func removeOlderExtensions() throws -> Int {
        let keepVer = options.verifyVersion.trimmingCharacters(in: .whitespacesAndNewlines)
        let name = options.verifyName.trimmingCharacters(in: .whitespacesAndNewlines)
        if name.isEmpty { return 0 }
        var removed = 0
        for _ in 0..<8 {
            let cards = axCollectNamedCards(appName: options.appName, verifyName: name)
            lastCardCount = cards.count
            step("  step      remove_older         cards=\(cards.count)")
            // Empty keep-version: remove every matching card (including the last).
            // Keep-version set: never delete the sole remaining card.
            if cards.isEmpty { return removed }
            if !keepVer.isEmpty && cards.count <= 1 { return removed }
            let victim = cards.first { card in
                if keepVer.isEmpty { return true }
                return card.version != keepVer
            }
            guard let victim else { return removed }
            axClickElement(victim.remove)
            step("  step      remove_older         clicked card Remove")
            if !confirmRemoveDialog() {
                step("  step      remove_older         confirm dialog not clicked — stop")
                return removed
            }
            sleep(0.5)
            let after = axCollectNamedCards(appName: options.appName, verifyName: name).count
            lastCardCount = after
            step("  step      remove_older         cards after=\(after)")
            if after >= cards.count {
                step("  step      remove_older         card count did not drop — stop")
                return removed
            }
            removed += cards.count - after
        }
        return removed
    }

    /// Chrome's unload confirm sits top-right (Cancel | Remove). AX often has no
    /// Cancel node; click the top-right Remove button, then Return as backup.
    func confirmRemoveDialog() -> Bool {
        sleep(0.45)
        if axClickTopRightRemove(appName: options.appName) {
            step("  step      remove_older         confirmed Remove (top-right)")
            sleep(0.5)
            return true
        }
        if axClickRemoveInConfirmDialog(appName: options.appName) {
            step("  step      remove_older         confirmed Remove (AX pair)")
            sleep(0.5)
            return true
        }
        pressReturnKey()
        step("  step      remove_older         confirmed Remove (return)")
        sleep(0.5)
        return true
    }

    func step(_ message: String) {
        let line = message.hasSuffix("\n") ? message : message + "\n"
        options.onLog?(line)
    }

    func sleep(_ t: TimeInterval) {
        Thread.sleep(forTimeInterval: t)
    }

    func yn(_ b: Bool) -> String { b ? "yes" : "no" }

    func nonempty(_ s: String) -> String? {
        let t = s.trimmingCharacters(in: .whitespacesAndNewlines)
        return t.isEmpty ? nil : t
    }
}

// MARK: - AppleScript (in-process; sender = this GUI app)

private func appleScript(_ source: String) throws -> String {
    // Same thread as the caller. main.sync from Task.detached deadlocks the UI
    // when NSAppleScript needs the run loop (Marcus Cancel stuck).
    try appleScriptNow(source)
}

private func appleScriptNow(_ source: String) throws -> String {
    var err: NSDictionary?
    guard let script = NSAppleScript(source: source) else {
        throw ChromeLoadUnpacked.LoadError("chrome: could not compile AppleScript")
    }
    let result = script.executeAndReturnError(&err)
    if let err {
        let msg = (err["NSAppleScriptErrorMessage"] as? String)
            ?? err.description
        throw ChromeLoadUnpacked.LoadError(msg)
    }
    return result.stringValue?.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
}

// MARK: - Chrome process list / pid

private func chromeMainProcesses() -> [String] {
    let proc = Process()
    proc.executableURL = URL(fileURLWithPath: "/bin/ps")
    proc.arguments = ["-ww", "-eo", "args="]
    let out = Pipe()
    proc.standardOutput = out
    proc.standardError = FileHandle.nullDevice
    do { try proc.run() } catch { return [] }
    // Drain before wait — a full pipe + waitUntilExit deadlocks (Cancel stuck).
    let data = out.fileHandleForReading.readDataToEndOfFile()
    proc.waitUntilExit()
    guard let text = String(data: data, encoding: .utf8) else { return [] }
    return text.split(separator: "\n").compactMap { line in
        let s = line.trimmingCharacters(in: .whitespaces)
        if s.isEmpty { return nil }
        if !s.contains("/MacOS/Google Chrome") { return nil }
        if s.contains("Helper") { return nil }
        return String(s)
    }
}

private func chromePID(_ appName: String) throws -> pid_t {
    if appName == ChromeLoadUnpacked.defaultAppName {
        let running = NSRunningApplication.runningApplications(
            withBundleIdentifier: ChromeLoadUnpacked.chromeBundleID
        )
        if let pid = running.first(where: { !$0.isTerminated })?.processIdentifier, pid > 0 {
            return pid
        }
    }
    let app = ChromeLoadUnpacked.escapeAppleScript(appName)
    let out = try appleScript("""
    tell application "System Events"
      try
        return unix id of process "\(app)" as text
      on error
        return ""
      end try
    end tell
    """)
    guard let pid = Int32(out), pid > 0 else {
        throw ChromeLoadUnpacked.LoadError("chrome pid: process \"\(appName)\" not found")
    }
    return pid
}

// MARK: - Accessibility (same DFS as ax_darwin.go)

private func axCF(_ attr: String) -> CFString { attr as CFString }

private func axCopyString(_ el: AXUIElement, attr: String) -> String {
    var ref: CFTypeRef?
    guard AXUIElementCopyAttributeValue(el, axCF(attr), &ref) == .success, let ref else {
        return ""
    }
    return (ref as? String) ?? ""
}

private func axFrameCenter(_ el: AXUIElement) -> CGPoint? {
    var posRef: CFTypeRef?
    var sizeRef: CFTypeRef?
    guard AXUIElementCopyAttributeValue(el, axCF(kAXPositionAttribute), &posRef) == .success,
          let posRef,
          AXUIElementCopyAttributeValue(el, axCF(kAXSizeAttribute), &sizeRef) == .success,
          let sizeRef
    else { return nil }
    var p = CGPoint.zero
    var s = CGSize.zero
    guard AXValueGetValue(posRef as! AXValue, .cgPoint, &p),
          AXValueGetValue(sizeRef as! AXValue, .cgSize, &s)
    else { return nil }
    return CGPoint(x: p.x + s.width / 2, y: p.y + s.height / 2)
}

private func axNameMatches(_ el: AXUIElement, want: String) -> (match: Bool, interactive: Bool) {
    let title = axCopyString(el, attr: kAXTitleAttribute)
    let desc = axCopyString(el, attr: kAXDescriptionAttribute)
    let role = axCopyString(el, attr: kAXRoleAttribute)
    let value = axCopyString(el, attr: kAXValueAttribute)
    let match = (title == want) || (desc == want) || (value == want)
    let interactive =
        role.contains("CheckBox") || role.contains("Button")
        || role.contains("checkBox") || role.contains("button")
        || role == "AXCheckBox" || role == "AXButton"
    return (match, interactive)
}

private func axChildren(_ el: AXUIElement) -> [AXUIElement] {
    var ref: CFTypeRef?
    guard AXUIElementCopyAttributeValue(el, axCF(kAXChildrenAttribute), &ref) == .success,
          let ref
    else { return [] }
    return (ref as? [AXUIElement]) ?? []
}

private func quartzClick(_ pt: CGPoint) {
    if let move = CGEvent(mouseEventSource: nil, mouseType: .mouseMoved, mouseCursorPosition: pt, mouseButton: .left) {
        move.post(tap: .cghidEventTap)
    }
    if let down = CGEvent(mouseEventSource: nil, mouseType: .leftMouseDown, mouseCursorPosition: pt, mouseButton: .left) {
        down.post(tap: .cghidEventTap)
    }
    if let up = CGEvent(mouseEventSource: nil, mouseType: .leftMouseUp, mouseCursorPosition: pt, mouseButton: .left) {
        up.post(tap: .cghidEventTap)
    }
}

private func axFindClick(
    _ el: AXUIElement,
    want: String,
    depth: Int,
    budget: inout Int,
    wantInteractive: Bool
) -> Bool {
    if depth < 0 || budget <= 0 { return false }
    budget -= 1
    let m = axNameMatches(el, want: want)
    if m.match && (!wantInteractive || m.interactive) {
        if let c = axFrameCenter(el), c.x > 1, c.y > 1 {
            quartzClick(c)
            return true
        }
        if m.interactive {
            AXUIElementPerformAction(el, axCF(kAXPressAction))
            return true
        }
    }
    for child in axChildren(el) {
        if axFindClick(child, want: want, depth: depth - 1, budget: &budget, wantInteractive: wantInteractive) {
            return true
        }
    }
    return false
}

private func axFindNamed(
    _ el: AXUIElement,
    want: String,
    depth: Int,
    budget: inout Int,
    wantInteractive: Bool
) -> Bool {
    if depth < 0 || budget <= 0 { return false }
    budget -= 1
    let m = axNameMatches(el, want: want)
    if m.match && (!wantInteractive || m.interactive) {
        if let c = axFrameCenter(el), c.x > 1, c.y > 1 { return true }
        if m.interactive { return true }
    }
    for child in axChildren(el) {
        if axFindNamed(child, want: want, depth: depth - 1, budget: &budget, wantInteractive: wantInteractive) {
            return true
        }
    }
    return false
}

private func axCountNamedWalk(
    _ el: AXUIElement,
    want: String,
    depth: Int,
    budget: inout Int
) -> Int {
    if depth < 0 || budget <= 0 { return 0 }
    budget -= 1
    var n = axNameMatches(el, want: want).match ? 1 : 0
    for child in axChildren(el) {
        n += axCountNamedWalk(child, want: want, depth: depth - 1, budget: &budget)
    }
    return n
}

private func axClickNamed(_ appName: String, _ name: String) throws {
    if !ChromeLoadUnpacked.isProcessTrusted() {
        throw ChromeLoadUnpacked.LoadError(
            "chrome: process is not Accessibility-trusted (System Settings → Privacy & Security → Accessibility)"
        )
    }
    let pid = try chromePID(appName)
    let app = AXUIElementCreateApplication(pid)
    var budget = 4000
    if axFindClick(app, want: name, depth: 28, budget: &budget, wantInteractive: true) {
        return
    }
    budget = 4000
    if axFindClick(app, want: name, depth: 28, budget: &budget, wantInteractive: false) {
        return
    }
    throw ChromeLoadUnpacked.LoadError("chrome: AX element \"\(name)\" not found")
}

private func axExistsNamed(_ appName: String, _ name: String) -> Bool {
    if !ChromeLoadUnpacked.isProcessTrusted() { return false }
    guard let pid = try? chromePID(appName) else { return false }
    let app = AXUIElementCreateApplication(pid)
    var budget = 4000
    if axFindNamed(app, want: name, depth: 28, budget: &budget, wantInteractive: true) {
        return true
    }
    budget = 4000
    return axFindNamed(app, want: name, depth: 28, budget: &budget, wantInteractive: false)
}

private func axCountNamed(_ appName: String, _ name: String) -> Int {
    if !ChromeLoadUnpacked.isProcessTrusted() { return -1 }
    guard let pid = try? chromePID(appName) else { return -1 }
    let app = AXUIElementCreateApplication(pid)
    var budget = 8000
    return axCountNamedWalk(app, want: name, depth: 28, budget: &budget)
}

private struct AXExtCard {
    var remove: AXUIElement
    var version: String
}

private func axNodeTexts(_ el: AXUIElement) -> [String] {
    [axCopyString(el, attr: kAXTitleAttribute),
     axCopyString(el, attr: kAXDescriptionAttribute),
     axCopyString(el, attr: kAXValueAttribute)]
        .map { $0.trimmingCharacters(in: .whitespacesAndNewlines) }
        .filter { !$0.isEmpty }
}

private func axIsRemoveButton(_ el: AXUIElement) -> Bool {
    axIsNamedButton(el, names: ["Remove", "Remove from Chrome"])
}

private func axIsNamedButton(_ el: AXUIElement, names: [String]) -> Bool {
    let texts = axNodeTexts(el)
    guard texts.contains(where: { names.contains($0) }) else { return false }
    let role = axCopyString(el, attr: kAXRoleAttribute)
    return role.contains("Button") || role.contains("button")
}

/// Chrome web UI often puts "Cancel"/"Remove" on a child static text, not the button.
private func axLabeled(_ el: AXUIElement, names: [String], labelDepth: Int = 2) -> Bool {
    if axNodeTexts(el).contains(where: { names.contains($0) }) { return true }
    if labelDepth <= 0 { return false }
    return axChildren(el).contains { axLabeled($0, names: names, labelDepth: labelDepth - 1) }
}

/// Deepest subtree that contains both Cancel and Remove; click the Remove-labeled node.
private func axFindConfirmRemoveButton(
    _ el: AXUIElement,
    depth: Int,
    budget: inout Int
) -> AXUIElement? {
    if depth < 0 || budget <= 0 { return nil }
    budget -= 1
    var childPair: AXUIElement?
    var sawCancel = axLabeled(el, names: ["Cancel"])
    var remove: AXUIElement? = axLabeled(el, names: ["Remove", "Remove from Chrome"]) ? el : nil
    for child in axChildren(el) {
        if let p = axFindConfirmRemoveButton(child, depth: depth - 1, budget: &budget) {
            childPair = p
        }
        if axLabeled(child, names: ["Cancel"]) { sawCancel = true }
        if axLabeled(child, names: ["Remove", "Remove from Chrome"]) { remove = child }
    }
    if let childPair { return childPair }
    if sawCancel, let remove { return remove }
    return nil
}

private func axClickRemoveInConfirmDialog(appName: String) -> Bool {
    if !ChromeLoadUnpacked.isProcessTrusted() { return false }
    guard let pid = try? chromePID(appName) else { return false }
    let app = AXUIElementCreateApplication(pid)
    var budget = 8000
    guard let remove = axFindConfirmRemoveButton(app, depth: 28, budget: &budget) else {
        return false
    }
    axClickElement(remove)
    return true
}

private func axConfirmDialogPresent(appName: String) -> Bool {
    if !ChromeLoadUnpacked.isProcessTrusted() { return false }
    guard let pid = try? chromePID(appName) else { return false }
    let app = AXUIElementCreateApplication(pid)
    var budget = 8000
    return axFindConfirmRemoveButton(app, depth: 28, budget: &budget) != nil
}

/// DFS: pair verifyName → version-like → Remove as one card (not raw name hits).
private func axCollectNamedCards(appName: String, verifyName: String) -> [AXExtCard] {
    if !ChromeLoadUnpacked.isProcessTrusted() { return [] }
    guard let pid = try? chromePID(appName) else { return [] }
    let app = AXUIElementCreateApplication(pid)
    var cards: [AXExtCard] = []
    var pendingName = false
    var pendingVer = ""
    var budget = 8000
    axWalkCards(
        app,
        verifyName: verifyName,
        depth: 28,
        budget: &budget,
        pendingName: &pendingName,
        pendingVer: &pendingVer,
        cards: &cards
    )
    return cards
}

private func axWalkCards(
    _ el: AXUIElement,
    verifyName: String,
    depth: Int,
    budget: inout Int,
    pendingName: inout Bool,
    pendingVer: inout String,
    cards: inout [AXExtCard]
) {
    if depth < 0 || budget <= 0 { return }
    budget -= 1
    if axNameMatches(el, want: verifyName).match {
        pendingName = true
        pendingVer = ""
    } else if pendingName {
        for t in axNodeTexts(el) where ChromeLoadUnpacked.isVersionLike(t) {
            pendingVer = t
            break
        }
    }
    if pendingName && axIsRemoveButton(el) {
        cards.append(AXExtCard(remove: el, version: pendingVer))
        pendingName = false
        pendingVer = ""
        return
    }
    for child in axChildren(el) {
        axWalkCards(
            child,
            verifyName: verifyName,
            depth: depth - 1,
            budget: &budget,
            pendingName: &pendingName,
            pendingVer: &pendingVer,
            cards: &cards
        )
    }
}

private func axDumpInteresting(_ el: AXUIElement, depth: Int, budget: inout Int, path: String) {
    if depth < 0 || budget <= 0 { return }
    budget -= 1
    let role = axCopyString(el, attr: kAXRoleAttribute)
    let texts = axNodeTexts(el)
    let blob = (texts + [role]).joined(separator: " ").lowercased()
    if blob.contains("cancel") || blob.contains("remove") || blob.contains("dialog")
        || blob.contains("sheet") || blob.contains("alert")
    {
        print("dump  \(path) role=\(role) texts=\(texts)")
    }
    let kids = axChildren(el)
    for (i, child) in kids.enumerated() {
        axDumpInteresting(child, depth: depth - 1, budget: &budget, path: "\(path)/\(i)")
    }
}

private func axFrontWindowFrame(appName: String) -> CGRect? {
    guard let pid = try? chromePID(appName) else { return nil }
    let app = AXUIElementCreateApplication(pid)
    var winRef: CFTypeRef?
    guard AXUIElementCopyAttributeValue(app, axCF(kAXFocusedWindowAttribute), &winRef) == .success
            || AXUIElementCopyAttributeValue(app, axCF(kAXMainWindowAttribute), &winRef) == .success,
          let winRef
    else { return nil }
    let win = winRef as! AXUIElement
    guard let c = axFrameCenter(win) else { return nil }
    var posRef: CFTypeRef?
    var sizeRef: CFTypeRef?
    guard AXUIElementCopyAttributeValue(win, axCF(kAXPositionAttribute), &posRef) == .success,
          let posRef,
          AXUIElementCopyAttributeValue(win, axCF(kAXSizeAttribute), &sizeRef) == .success,
          let sizeRef
    else {
        _ = c
        return nil
    }
    var p = CGPoint.zero
    var s = CGSize.zero
    guard AXValueGetValue(posRef as! AXValue, .cgPoint, &p),
          AXValueGetValue(sizeRef as! AXValue, .cgSize, &s)
    else { return nil }
    return CGRect(origin: p, size: s)
}

private func axCollectRemoveButtons(
    _ el: AXUIElement,
    depth: Int,
    budget: inout Int,
    into out: inout [(el: AXUIElement, center: CGPoint)]
) {
    if depth < 0 || budget <= 0 { return }
    budget -= 1
    if axIsRemoveButton(el) || (axLabeled(el, names: ["Remove"]) && axCopyString(el, attr: kAXRoleAttribute).contains("Button")) {
        if let c = axFrameCenter(el), c.x > 1, c.y > 1 {
            out.append((el, c))
        }
    }
    for child in axChildren(el) {
        axCollectRemoveButtons(child, depth: depth - 1, budget: &budget, into: &out)
    }
}

/// The confirm dialog is top-right; its Remove is the rightmost button in the upper band.
private func axClickTopRightRemove(appName: String) -> Bool {
    if !ChromeLoadUnpacked.isProcessTrusted() { return false }
    guard let pid = try? chromePID(appName) else { return false }
    let app = AXUIElementCreateApplication(pid)
    var budget = 8000
    var buttons: [(el: AXUIElement, center: CGPoint)] = []
    axCollectRemoveButtons(app, depth: 28, budget: &budget, into: &buttons)
    if buttons.isEmpty { return false }
    let pick: (el: AXUIElement, center: CGPoint)
    if let win = axFrontWindowFrame(appName: appName) {
        let upper = win.minY + win.height * 0.4
        let right = buttons.filter { $0.center.y < upper && $0.center.x > win.midX }
        pick = right.max(by: { $0.center.x < $1.center.x })
            ?? buttons.max(by: { $0.center.x < $1.center.x })!
    } else {
        pick = buttons.max(by: { $0.center.x < $1.center.x })!
    }
    quartzClick(pick.center)
    return true
}

private func pressReturnKey() {
    let src = CGEventSource(stateID: .hidSystemState)
    let down = CGEvent(keyboardEventSource: src, virtualKey: 36, keyDown: true)
    let up = CGEvent(keyboardEventSource: src, virtualKey: 36, keyDown: false)
    down?.post(tap: .cghidEventTap)
    up?.post(tap: .cghidEventTap)
}

private func axClickElement(_ el: AXUIElement) {
    if let c = axFrameCenter(el), c.x > 1, c.y > 1 {
        quartzClick(c)
        return
    }
    AXUIElementPerformAction(el, axCF(kAXPressAction))
}

// MARK: - Wait

private final class WaitBox<T>: @unchecked Sendable {
    private let lock = NSLock()
    private let sem = DispatchSemaphore(value: 0)
    private var value: T?
    private var done = false

    func finish(_ v: T) {
        lock.lock()
        if !done {
            value = v
            done = true
            sem.signal()
        }
        lock.unlock()
    }

    func wait(timeout: TimeInterval) -> T? {
        _ = sem.wait(timeout: .now() + timeout)
        lock.lock()
        defer { lock.unlock() }
        return value
    }
}
