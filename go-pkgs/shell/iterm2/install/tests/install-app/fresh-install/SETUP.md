# Scenario

**Feature**: fresh install to explicit target under case dir

```
extracted mini-app -> InstallApp -> WorkDir/dest/iTerm.app
```

## Steps

1. Leave `TargetApp` empty so Run defaults to `WorkDir/dest/iTerm.app`.
2. Do not seed an existing target (`SeedExistingTarget` remains false).
3. Do not request home default target (`UseDefaultTarget` remains false).

No leaf-specific Setup: root + group Setup and Run defaults are sufficient.
