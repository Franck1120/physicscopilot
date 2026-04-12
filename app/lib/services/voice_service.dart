import 'dart:async';

import 'package:flutter_tts/flutter_tts.dart';
import 'package:permission_handler/permission_handler.dart';
import 'package:speech_to_text/speech_to_text.dart' as stt;

/// Manages Speech-to-Text (STT) and Text-to-Speech (TTS) for voice I/O.
///
/// STT uses [speech_to_text]; TTS uses [flutter_tts].
/// Call [initialize] once before any other method.
class VoiceService {
  final _stt = stt.SpeechToText();
  final _tts = FlutterTts();

  final _recognizedTextController = StreamController<String>.broadcast();

  /// Emits `true` when TTS starts speaking, `false` when it stops.
  /// VoiceNotifier subscribes to this to keep [VoiceState.isSpeaking] in sync.
  final _speakingController = StreamController<bool>.broadcast();

  final List<String> _ttsQueue = [];

  bool _isListening = false;
  bool _isSpeaking = false;
  bool _sttReady = false;

  /// Emits finalised, non-empty recognised words from the microphone.
  Stream<String> get recognizedText => _recognizedTextController.stream;

  /// Emits speaking state changes so callers can stay in sync.
  Stream<bool> get speakingStream => _speakingController.stream;

  bool get isListening => _isListening;
  bool get isSpeaking => _isSpeaking;

  // ── Lifecycle ─────────────────────────────────────────────────────────────

  /// Initialises TTS configuration and STT engine.
  Future<void> initialize() async {
    await _configureTts();
    _sttReady = await _stt.initialize(
      onError: (error) => _isListening = false,
      onStatus: (status) {
        if (status == stt.SpeechToText.notListeningStatus) {
          _isListening = false;
        }
      },
    );
  }

  // ── Language ─────────────────────────────────────────────────────────────

  /// Maps a BCP-47 language code to the corresponding TTS locale string.
  static String _toTtsLocale(String lang) => switch (lang) {
        'it' => 'it-IT',
        'en' => 'en-US',
        'fr' => 'fr-FR',
        'de' => 'de-DE',
        'es' => 'es-ES',
        'pt' => 'pt-PT',
        _ => '$lang-${lang.toUpperCase()}', // best-effort fallback
      };

  /// Updates the TTS and STT language at runtime.
  Future<void> setLanguage(String langCode) async {
    await _tts.setLanguage(_toTtsLocale(langCode));
  }

  Future<void> _configureTts() async {
    await _tts.setLanguage('it-IT');
    await _tts.setSpeechRate(0.9);
    await _tts.setPitch(1.0);
    await _tts.setVolume(1.0);

    _tts.setCompletionHandler(() {
      _isSpeaking = false;
      _speakingController.add(false);
      _processQueue();
    });

    _tts.setErrorHandler((_) {
      _isSpeaking = false;
      _speakingController.add(false);
      _processQueue();
    });
  }

  // ── STT ───────────────────────────────────────────────────────────────────

  /// Requests microphone permission (if not already granted) and starts
  /// listening. No-op if STT is not ready or already listening.
  Future<void> startListening() async {
    if (!_sttReady || _isListening) return;

    final status = await Permission.microphone.request();
    if (!status.isGranted) return;

    await _stt.listen(
      onResult: (result) {
        if (result.finalResult &&
            result.recognizedWords.isNotEmpty &&
            !_recognizedTextController.isClosed) {
          _recognizedTextController.add(result.recognizedWords);
        }
      },
      listenFor: const Duration(seconds: 30),
      pauseFor: const Duration(seconds: 3),
      localeId: 'en-US',
    );
    _isListening = true;
  }

  /// Stops the active STT listener.
  Future<void> stopListening() async {
    if (!_isListening) return;
    await _stt.stop();
    _isListening = false;
  }

  // ── TTS ───────────────────────────────────────────────────────────────────

  /// Enqueues [text] for TTS playback.
  /// Starts immediately if nothing is currently playing.
  void speak(String text) {
    _ttsQueue.add(text);
    if (!_isSpeaking) _processQueue();
  }

  void _processQueue() {
    if (_ttsQueue.isEmpty) return;
    final text = _ttsQueue.removeAt(0);
    _isSpeaking = true;
    _speakingController.add(true);
    _tts.speak(text);
  }

  /// Stops TTS immediately and clears any queued utterances.
  Future<void> stop() async {
    _ttsQueue.clear();
    _isSpeaking = false;
    _speakingController.add(false);
    await _tts.stop();
  }

  // ── Dispose ───────────────────────────────────────────────────────────────

  Future<void> dispose() async {
    await stopListening();
    await stop();
    if (!_recognizedTextController.isClosed) {
      await _recognizedTextController.close();
    }
    if (!_speakingController.isClosed) {
      await _speakingController.close();
    }
  }
}
