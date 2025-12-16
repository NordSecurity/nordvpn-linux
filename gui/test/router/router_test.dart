import 'package:faker/faker.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/router/router.dart';
import 'package:nordvpn/router/routes.dart';

void main() {
  group('RedirectState.update', () {
    test('notifies listeners when state changes', () {
      final redirectState = RedirectState();
      int notificationCount = 0;
      redirectState.addListener(() => notificationCount++);
      // update to a known state
      redirectState.update(
        isLoading: false,
        hasError: false,
        isLoggedIn: false,
        displayConsent: false,
      );
      expect(notificationCount, equals(0));

      redirectState.update(
        isLoading: true,
        hasError: false,
        isLoggedIn: false,
        displayConsent: false,
      );

      expect(notificationCount, equals(1));
    });

    test('does not notify listeners when state remains unchanged', () {
      final redirectState = RedirectState();
      int notificationCount = 0;
      // update to a known state
      redirectState.update(
        isLoading: false,
        hasError: false,
        isLoggedIn: false,
        displayConsent: false,
      );
      redirectState.addListener(() => notificationCount++);

      // updating with the same values should not trigger a notification
      redirectState.update(
        isLoading: false,
        hasError: false,
        isLoggedIn: false,
        displayConsent: false,
      );

      expect(notificationCount, equals(0));
    });
  });

  group('RedirectState.route', () {
    test('returns loadingScreen when isLoading is true', () {
      final redirectState = RedirectState();
      redirectState.update(
        isLoading: true,
        hasError: false,
        isLoggedIn: false,
        displayConsent: false,
      );
      final notImportant = Uri.parse(faker.lorem.word());
      expect(redirectState.route(notImportant), equals(AppRoute.loadingScreen));
    });

    test('returns errorScreen when hasError is true (and not loading)', () {
      final redirectState = RedirectState();
      redirectState.update(
        isLoading: false,
        hasError: true,
        isLoggedIn: false,
        displayConsent: false,
      );
      final notImportant = Uri.parse(faker.lorem.word());
      expect(redirectState.route(notImportant), equals(AppRoute.errorScreen));
    });

    test('returns consent screen', () {
      final redirectState = RedirectState();
      redirectState.update(
        isLoading: false,
        hasError: false,
        isLoggedIn: true,
        displayConsent: true,
      );
      final consentUrl = Uri.parse(AppRoute.consentScreen.toString());
      expect(redirectState.route(consentUrl), equals(AppRoute.consentScreen));
    });

    test('returns login when not logged in (and not loading nor error)', () {
      final redirectState = RedirectState();
      redirectState.update(
        isLoading: false,
        hasError: false,
        isLoggedIn: false,
        displayConsent: false,
      );
      final notImportant = Uri.parse(faker.lorem.word());
      expect(redirectState.route(notImportant), equals(AppRoute.login));
    });

    test('returns vpn when logged in and current uri is login', () {
      final redirectState = RedirectState();
      redirectState.update(
        isLoading: false,
        hasError: false,
        isLoggedIn: true,
        displayConsent: false,
      );
      final loginUri = Uri.parse(AppRoute.login.toString());
      expect(redirectState.route(loginUri), equals(AppRoute.vpn));
    });

    test('returns null when logged in and not on login screen', () {
      final redirectState = RedirectState();
      redirectState.update(
        isLoading: false,
        hasError: false,
        isLoggedIn: true,
        displayConsent: false,
      );
      final notLoginUri = Uri.parse(faker.lorem.word());
      expect(redirectState.route(notLoginUri), isNull);
    });
  });
}
