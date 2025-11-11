import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/internal/urls.dart';
import 'package:nordvpn/settings/account_details_screen.dart';

void main() {
  group('Account Details Screen Link Tests', () {
    // Table of URLs with their expected properties
    final items = [
      (
        name: 'Manage Subscription',
        url: manageSubscriptionUrl,
        host: 'my.nordaccount.com',
        path: 'billing/my-subscriptions',
        campaign: 'settings_account-manage_subscription',
        ncValue: 'settings-manage_subscription',
        ownerId: 'nordvpn',
      ),
      (
        name: 'Change Password',
        url: changePasswordUrl,
        host: 'my.nordaccount.com',
        path: 'account-settings/account-management',
        campaign: 'settings_account-change_password',
        ncValue: 'settings-change_password',
        ownerId: 'nordvpn',
      ),
      (
        name: 'NordPass',
        url: nordPassProductUrl,
        host: 'nordpass.com',
        path: null,
        campaign: 'settings_apps-explore_nordpass',
        ncValue: 'settings-explore_nordpass',
        ownerId: null,
      ),
      (
        name: 'NordLocker',
        url: nordLockerProductUrl,
        host: 'nordlocker.com',
        path: null,
        campaign: 'settings_apps-explore_nordlocker',
        ncValue: 'settings-explore_nordlocker',
        ownerId: null,
      ),
      (
        name: 'NordLayer',
        url: nordLayerProductUrl,
        host: 'nordlayer.com',
        path: null,
        campaign: 'settings_apps-explore_nordlayer',
        ncValue: 'settings-explore_nordlayer',
        ownerId: null,
      ),
    ];

    for (final item in items) {
      test('${item.name} URL is correct', () {
        expect(
          item.url.scheme,
          equals('https'),
          reason: '${item.name} should use HTTPS',
        );
        expect(
          item.url.host,
          equals(item.host),
          reason: '${item.name} should point to ${item.host}',
        );
        expect(
          item.url.path,
          contains(item.path ?? '/'),
          reason: '${item.name} path should contain ${item.path}',
        );
        expect(
          item.url.queryParameters['utm_medium'],
          equals('app'),
          reason: '${item.name} should have utm_medium=app',
        );
        expect(
          item.url.queryParameters['utm_source'],
          equals('nordvpn-linux-gui'),
          reason: '${item.name} should have utm_source=nordvpn-linux-gui',
        );
        expect(
          item.url.queryParameters['utm_campaign'],
          equals(item.campaign),
          reason: '${item.name} should have correct utm_campaign',
        );
        // Check Nord-specific tracking
        expect(
          item.url.queryParameters['nm'],
          equals('app'),
          reason: '${item.name} should have nm=app',
        );
        expect(
          item.url.queryParameters['ns'],
          equals('nordvpn-linux-gui'),
          reason: '${item.name} should have ns=nordvpn-linux-gui',
        );
        expect(
          item.url.queryParameters['nc'],
          equals(item.ncValue),
          reason: '${item.name} should have correct nc parameter',
        );
        expect(
          item.url.queryParameters['owner_id'],
          equals(item.ownerId),
          reason: '${item.name} should have owner_id=${item.ownerId}',
        );
      });
    }

    test('ProductItem constructor creates valid object', () {
      final testProduct = ProductItem(
        title: 'Test Product',
        subtitle: 'Test Description',
        uri: nordPassProductUrl,
        imageName: 'test.svg',
      );

      expect(testProduct.title, equals('Test Product'));
      expect(testProduct.subtitle, equals('Test Description'));
      expect(testProduct.uri, equals(nordPassProductUrl));
      expect(testProduct.imageName, equals('test.svg'));
    });
  });
}
