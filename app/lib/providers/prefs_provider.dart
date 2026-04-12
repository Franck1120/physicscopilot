// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Exposes the [SharedPreferences] instance initialised before [runApp].
///
/// Must be overridden via [ProviderScope.overrides] in [main]:
/// ```dart
/// sharedPrefsProvider.overrideWithValue(prefs)
/// ```
final sharedPrefsProvider = Provider<SharedPreferences>((ref) {
  throw UnimplementedError('sharedPrefsProvider must be overridden in main()');
});
