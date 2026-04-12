# Mobile Build Guide

## Release Checklist (before every release)

- [ ] `flutter analyze --fatal-warnings` passes
- [ ] `flutter test` passes
- [ ] No `debugPrint` left in production code
- [ ] Version bumped in `pubspec.yaml` (`version: X.Y.Z+buildNumber`)
- [ ] CHANGELOG.md updated
- [ ] `SERVER_URL` points to production server
- [ ] Signing configured (see below)

---

## Android

### Prerequisites
- Android Studio (or command-line tools)
- Java 17
- A signing keystore

### Create signing keystore (first time only)
```bash
keytool -genkey -v \
  -keystore android/app/upload-keystore.jks \
  -keyalg RSA -keysize 2048 -validity 10000 \
  -alias upload \
  -storepass <KEYSTORE_PASSWORD> \
  -keypass <KEY_PASSWORD>
```

Create `android/key.properties` (gitignored):
```
storePassword=<KEYSTORE_PASSWORD>
keyPassword=<KEY_PASSWORD>
keyAlias=upload
storeFile=upload-keystore.jks
```

### Build release AAB (Play Store)
```bash
cd app

flutter build appbundle --release \
  --dart-define=SERVER_URL=https://your-server.onrender.com

# Output: build/app/outputs/bundle/release/app-release.aab
```

### Build release APK (direct install)
```bash
flutter build apk --release \
  --dart-define=SERVER_URL=https://your-server.onrender.com \
  --split-per-abi   # creates separate APKs per architecture
```

### Play Store upload
1. Go to Google Play Console → Release → Production
2. Upload `app-release.aab`
3. Fill release notes
4. Roll out

---

## iOS

### Prerequisites
- macOS 14+ with Xcode 15+
- Apple Developer account ($99/year)
- App Store Connect access

### Configure signing
1. Open `app/ios/Runner.xcworkspace` in Xcode
2. Select Runner target → Signing & Capabilities
3. Set Team to your Apple Developer account
4. Xcode will auto-manage provisioning profiles

### Build release IPA
```bash
cd app

flutter build ios --release \
  --dart-define=SERVER_URL=https://your-server.onrender.com

# Then in Xcode:
# Product → Archive → Distribute App → App Store Connect
```

### Camera and microphone permissions

`Info.plist` must include:
```xml
<key>NSCameraUsageDescription</key>
<string>PhysicsCopilot uses the camera to analyze your device in real time.</string>
<key>NSMicrophoneUsageDescription</key>
<string>PhysicsCopilot uses the microphone for hands-free voice guidance.</string>
<key>NSSpeechRecognitionUsageDescription</key>
<string>PhysicsCopilot uses speech recognition to accept voice commands.</string>
```

These are already present in `app/ios/Runner/Info.plist`.

---

## Version bumping

```bash
# pubspec.yaml
version: 1.2.3+45
#        ^     ^
#        |     build number (auto-incremented each release)
#        semantic version
```

The build number must be higher than the previous release — both Play Store and App Store reject builds with lower build numbers.

---

## Over-the-air config updates

The `SERVER_URL` is baked into the binary at build time (`--dart-define`). To change it without a new release, the app reads from `SharedPreferences` as an override (Settings screen). This allows pointing the app at a new server URL without a full rebuild.

---

## Troubleshooting builds

**"Gradle build failed"**: Run `cd app/android && ./gradlew clean` then retry.

**"CocoaPods error"**: Run `cd app/ios && pod install --repo-update` then retry.

**"Flutter tool version mismatch"**: Run `flutter upgrade` to match the version in `ci.yml`.

**"Signing not configured"**: Check `android/key.properties` and Xcode signing settings.
