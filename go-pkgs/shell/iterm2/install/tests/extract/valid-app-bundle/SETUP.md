# Scenario

**Feature**: zip with mini `iTerm.app` extracts to app path

```
zip(iTerm.app/Contents/...) -> ExtractApp -> .../iTerm.app
```

## Steps

1. Rely on Run's default zip fixture (`ZipWithoutApp=false`) which embeds a mini
   `iTerm.app` with `Contents/Info.plist` and `Contents/MacOS/iTerm2`.

No leaf-specific Setup beyond group `Operation=extract`.
