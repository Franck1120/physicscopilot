import 'package:flutter_test/flutter_test.dart';
import 'package:physicscopilot/utils/strings.dart';

void main() {
  group('AppStrings — app strings', () {
    test('appName is not empty', () {
      expect(AppStrings.appName, isNotEmpty);
    });

    test('appTagline is not empty', () {
      expect(AppStrings.appTagline, isNotEmpty);
    });

    test('appVersion is not empty', () {
      expect(AppStrings.appVersion, isNotEmpty);
    });
  });

  group('AppStrings — common strings', () {
    test('save is not empty', () {
      expect(AppStrings.save, isNotEmpty);
    });

    test('cancel is not empty', () {
      expect(AppStrings.cancel, isNotEmpty);
    });

    test('delete is not empty', () {
      expect(AppStrings.delete, isNotEmpty);
    });

    test('close is not empty', () {
      expect(AppStrings.close, isNotEmpty);
    });

    test('reset is not empty', () {
      expect(AppStrings.reset, isNotEmpty);
    });

    test('retry is not empty', () {
      expect(AppStrings.retry, isNotEmpty);
    });

    test('errorGeneric is not empty', () {
      expect(AppStrings.errorGeneric, isNotEmpty);
    });

    test('restartApp is not empty', () {
      expect(AppStrings.restartApp, isNotEmpty);
    });
  });

  group('AppStrings — session strings', () {
    test('sessionTitle is not empty', () {
      expect(AppStrings.sessionTitle, isNotEmpty);
    });

    test('sessionEndSession is not empty', () {
      expect(AppStrings.sessionEndSession, isNotEmpty);
    });

    test('sessionNewAnalysis is not empty', () {
      expect(AppStrings.sessionNewAnalysis, isNotEmpty);
    });

    test('sessionIdle is not empty', () {
      expect(AppStrings.sessionIdle, isNotEmpty);
    });

    test('sessionAiThinking is not empty', () {
      expect(AppStrings.sessionAiThinking, isNotEmpty);
    });

    test('sessionResponseCopied is not empty', () {
      expect(AppStrings.sessionResponseCopied, isNotEmpty);
    });

    test('sessionCopyResponse is not empty', () {
      expect(AppStrings.sessionCopyResponse, isNotEmpty);
    });

    test('sessionServerUnreachable is not empty', () {
      expect(AppStrings.sessionServerUnreachable, isNotEmpty);
    });

    test('sessionFrameError is not empty', () {
      expect(AppStrings.sessionFrameError, isNotEmpty);
    });

    test('sessionAiTimeout is not empty', () {
      expect(AppStrings.sessionAiTimeout, isNotEmpty);
    });

    test('sessionErrorFromServer is not empty', () {
      expect(AppStrings.sessionErrorFromServer, isNotEmpty);
    });

    test('sessionInvalidResponse is not empty', () {
      expect(AppStrings.sessionInvalidResponse, isNotEmpty);
    });

    test('sessionErrorUnknown is not empty', () {
      expect(AppStrings.sessionErrorUnknown, isNotEmpty);
    });

    test('sessionDescribeProblem is not empty', () {
      expect(AppStrings.sessionDescribeProblem, isNotEmpty);
    });

    test('sessionSend is not empty', () {
      expect(AppStrings.sessionSend, isNotEmpty);
    });

    test('sessionNoResponseYet is not empty', () {
      expect(AppStrings.sessionNoResponseYet, isNotEmpty);
    });

    test('sessionCameraInit is not empty', () {
      expect(AppStrings.sessionCameraInit, isNotEmpty);
    });

    test('sessionCameraUnavailable is not empty', () {
      expect(AppStrings.sessionCameraUnavailable, isNotEmpty);
    });
  });

  group('AppStrings — tutorial strings', () {
    test('tutorialHint is not empty', () {
      expect(AppStrings.tutorialHint, isNotEmpty);
    });

    test('tutorialDismiss is not empty', () {
      expect(AppStrings.tutorialDismiss, isNotEmpty);
    });
  });

  group('AppStrings — home strings', () {
    test('homeNewSession is not empty', () {
      expect(AppStrings.homeNewSession, isNotEmpty);
    });

    test('homeNewSessionSub is not empty', () {
      expect(AppStrings.homeNewSessionSub, isNotEmpty);
    });

    test('homeSectionDevice is not empty', () {
      expect(AppStrings.homeSectionDevice, isNotEmpty);
    });

    test('homeSectionSessions is not empty', () {
      expect(AppStrings.homeSectionSessions, isNotEmpty);
    });

    test('homeSeeAll is not empty', () {
      expect(AppStrings.homeSeeAll, isNotEmpty);
    });

    test('homeNoDevice is not empty', () {
      expect(AppStrings.homeNoDevice, isNotEmpty);
    });

    test('homeChangeDevice is not empty', () {
      expect(AppStrings.homeChangeDevice, isNotEmpty);
    });

    test('homeChangeDeviceLabel is not empty', () {
      expect(AppStrings.homeChangeDeviceLabel, isNotEmpty);
    });

    test('homeEmptyTitle is not empty', () {
      expect(AppStrings.homeEmptyTitle, isNotEmpty);
    });

    test('homeEmptySubtitle is not empty', () {
      expect(AppStrings.homeEmptySubtitle, isNotEmpty);
    });

    test('homeServerReachable is not empty', () {
      expect(AppStrings.homeServerReachable, isNotEmpty);
    });

    test('homeServerUnreachable is not empty', () {
      expect(AppStrings.homeServerUnreachable, isNotEmpty);
    });

    test('homeServerChecking is not empty', () {
      expect(AppStrings.homeServerChecking, isNotEmpty);
    });
  });

  group('AppStrings — history strings', () {
    test('historyTitle is not empty', () {
      expect(AppStrings.historyTitle, isNotEmpty);
    });

    test('historyClearAll is not empty', () {
      expect(AppStrings.historyClearAll, isNotEmpty);
    });

    test('historyClearConfirm is not empty', () {
      expect(AppStrings.historyClearConfirm, isNotEmpty);
    });

    test('historyEmpty is not empty', () {
      expect(AppStrings.historyEmpty, isNotEmpty);
    });

    test('historyEmptySub is not empty', () {
      expect(AppStrings.historyEmptySub, isNotEmpty);
    });

    test('historyStatusResolved is not empty', () {
      expect(AppStrings.historyStatusResolved, isNotEmpty);
    });

    test('historyStatusUnresolved is not empty', () {
      expect(AppStrings.historyStatusUnresolved, isNotEmpty);
    });

    test('historySectionAI is not empty', () {
      expect(AppStrings.historySectionAI, isNotEmpty);
    });
  });

  group('AppStrings — settings strings', () {
    test('settingsTitle is not empty', () {
      expect(AppStrings.settingsTitle, isNotEmpty);
    });

    test('settingsSectionConnection is not empty', () {
      expect(AppStrings.settingsSectionConnection, isNotEmpty);
    });

    test('settingsSectionFeatures is not empty', () {
      expect(AppStrings.settingsSectionFeatures, isNotEmpty);
    });

    test('settingsSectionServerInfo is not empty', () {
      expect(AppStrings.settingsSectionServerInfo, isNotEmpty);
    });

    test('settingsSectionAbout is not empty', () {
      expect(AppStrings.settingsSectionAbout, isNotEmpty);
    });

    test('settingsServerUrl is not empty', () {
      expect(AppStrings.settingsServerUrl, isNotEmpty);
    });

    test('settingsServerUrlHint is not empty', () {
      expect(AppStrings.settingsServerUrlHint, isNotEmpty);
    });

    test('settingsServerUrlLabel is not empty', () {
      expect(AppStrings.settingsServerUrlLabel, isNotEmpty);
    });

    test('settingsVoice is not empty', () {
      expect(AppStrings.settingsVoice, isNotEmpty);
    });

    test('settingsVoiceSub is not empty', () {
      expect(AppStrings.settingsVoiceSub, isNotEmpty);
    });

    test('settingsUrlSaved is not empty', () {
      expect(AppStrings.settingsUrlSaved, isNotEmpty);
    });

    test('settingsUrlReset is not empty', () {
      expect(AppStrings.settingsUrlReset, isNotEmpty);
    });
  });

  group('AppStrings — about strings', () {
    test('aboutPoweredBy is not empty', () {
      expect(AppStrings.aboutPoweredBy, isNotEmpty);
    });

    test('aboutEngine is not empty', () {
      expect(AppStrings.aboutEngine, isNotEmpty);
    });

    test('aboutPrivacy is not empty', () {
      expect(AppStrings.aboutPrivacy, isNotEmpty);
    });

    test('aboutTerms is not empty', () {
      expect(AppStrings.aboutTerms, isNotEmpty);
    });

    test('aboutDetails is not empty', () {
      expect(AppStrings.aboutDetails, isNotEmpty);
    });
  });

  group('AppStrings — connection / camera strings', () {
    test('connConnected is not empty', () {
      expect(AppStrings.connConnected, isNotEmpty);
    });

    test('connConnecting is not empty', () {
      expect(AppStrings.connConnecting, isNotEmpty);
    });

    test('connDisconnected is not empty', () {
      expect(AppStrings.connDisconnected, isNotEmpty);
    });

    test('connError is not empty', () {
      expect(AppStrings.connError, isNotEmpty);
    });

    test('connBannerConnecting is not empty', () {
      expect(AppStrings.connBannerConnecting, isNotEmpty);
    });

    test('connBannerUnreachable is not empty', () {
      expect(AppStrings.connBannerUnreachable, isNotEmpty);
    });

    test('cameraError is not empty', () {
      expect(AppStrings.cameraError, isNotEmpty);
    });

    test('cameraOfflineBanner is not empty', () {
      expect(AppStrings.cameraOfflineBanner, isNotEmpty);
    });

    test('cameraWriteHere is not empty', () {
      expect(AppStrings.cameraWriteHere, isNotEmpty);
    });
  });

  group('AppStrings — onboarding strings', () {
    test('onboardingNext is not empty', () {
      expect(AppStrings.onboardingNext, isNotEmpty);
    });

    test('onboardingStart is not empty', () {
      expect(AppStrings.onboardingStart, isNotEmpty);
    });

    test('onboardingSlide1Title is not empty', () {
      expect(AppStrings.onboardingSlide1Title, isNotEmpty);
    });

    test('onboardingSlide2Title is not empty', () {
      expect(AppStrings.onboardingSlide2Title, isNotEmpty);
    });
  });

  group('AppStrings — profile strings', () {
    test('profileTitle is not empty', () {
      expect(AppStrings.profileTitle, isNotEmpty);
    });

    test('profileSettings is not empty', () {
      expect(AppStrings.profileSettings, isNotEmpty);
    });

    test('profilePrivacy is not empty', () {
      expect(AppStrings.profilePrivacy, isNotEmpty);
    });

    test('profileAbout is not empty', () {
      expect(AppStrings.profileAbout, isNotEmpty);
    });
  });
}
