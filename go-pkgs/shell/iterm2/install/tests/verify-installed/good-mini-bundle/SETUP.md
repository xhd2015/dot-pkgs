# Scenario

**Feature**: good mini-bundle passes VerifyInstalled

```
iTerm.app with BundleID + MacOS/iTerm2 -> nil
```

## Steps

1. Rely on Run defaults: correct `install.BundleID` and MacOS binary present
   (`OmitMacOSBinary=false`, empty `BundleIDOverride`).

No leaf-specific Setup beyond group `Operation=verify-installed`.
