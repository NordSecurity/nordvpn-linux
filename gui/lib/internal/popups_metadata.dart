import 'package:flutter/foundation.dart';
import 'package:nordvpn/data/models/popup_metadata.dart';
import 'package:nordvpn/data/providers/account_controller.dart';
import 'package:nordvpn/data/providers/preferences_controller.dart';
import 'package:nordvpn/data/providers/vpn_settings_controller.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/i18n/daemon_code_messages.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/popup_codes.dart';
import 'package:nordvpn/internal/uri_launch_extension.dart';
import 'package:nordvpn/internal/urls.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

typedef PopupOrErrorCode = int;

// Gives metadata based on the specified code.
PopupMetadata givePopupMetadata(PopupOrErrorCode code) {
  final metadata = switch (code) {
    // ==============================    [ triggered by app ]    ==============================

    // DNS Servers set - Turn off Custom DNS?
    PopupCodes.turnOffCustomDns => DecisionPopupMetadata(
      id: PopupCodes.turnOffCustomDns,
      title: t.ui.turnOffCustomDns,
      message: (_) => t.ui.turnOffCustomDnsDescription,
      noButtonText: t.ui.cancel,
      yesButtonText: t.ui.turnOff,
      yesAction: (ref) async {
        await ref.read(vpnSettingsControllerProvider.notifier).clearCustomDns();
      },
    ),

    // Enabling Custom DNS - Turn off Threat Protection?
    PopupCodes.turnOffThreatProtection => DecisionPopupMetadata(
      id: PopupCodes.turnOffThreatProtection,
      title: t.ui.threatProtectionWillTurnOff,
      message: (_) => t.ui.threatProtectionWillTurnOffDescription,
      noButtonText: t.ui.cancel,
      yesButtonText: t.ui.setCustomDns,
      yesAction: (ref) async {
        await ref
            .read(vpnSettingsControllerProvider.notifier)
            .setThreatProtection(false);
        ref.read(vpnSettingsControllerProvider.notifier).setCustomDns(true);
      },
    ),

    // Enabling Threat Protection - Turn off Custom DNS?
    PopupCodes.resetCustomDns => DecisionPopupMetadata(
      id: PopupCodes.resetCustomDns,
      title: t.ui.resetCustomDns,
      message: (_) => t.ui.resetCustomDnsDescription,
      noButtonText: t.ui.cancel,
      yesButtonText: t.ui.continueWord,
      yesAction: (ref) async {
        ref.read(vpnSettingsControllerProvider.notifier).setCustomDns(false);
        await ref
            .read(vpnSettingsControllerProvider.notifier)
            .setThreatProtection(true);
      },
    ),

    // Buy Dedicated IP
    PopupCodes.getDedicatedIp => RichPopupMetadata(
      id: PopupCodes.getDedicatedIp,
      header: t.ui.getYourDip,
      message: (_) => t.ui.getDipDescription,
      actionButtonText: t.ui.getDip,
      image: DynamicThemeImage("get_dedicated_ip.svg"),
      action: (_) async => await getDedicatedIpUrl.launch(),
    ),

    // Choose location for Dedicated IP
    PopupCodes.chooseDip => RichPopupMetadata(
      id: PopupCodes.chooseDip,
      header: t.ui.chooseLocationForDip,
      message: (_) => t.ui.dipSelectLocationDescription,
      image: DynamicThemeImage("dedicated_ip_select_server.svg"),
      actionButtonText: t.ui.selectLocation,
      autoClose: true,
      action: (ref) async => await chooseDedicatedIpUrl.launch(),
    ),

    // Reset settings?
    PopupCodes.resetSettings ||
    PopupCodes.resetSettingsAndDisconnect => _resetToDefaults(code),

    // Allow List has entries set - Turn it off?
    PopupCodes.turnOffAllowList => DecisionPopupMetadata(
      id: PopupCodes.turnOffAllowList,
      title: t.ui.turnOffAllowList,
      message: (_) => t.ui.turnOffAllowListDescription,
      noButtonText: t.ui.cancel,
      yesButtonText: t.ui.turnOff,
      yesAction: (ref) async {
        await ref
            .read(vpnSettingsControllerProvider.notifier)
            .disableAllowList();
      },
    ),

    // Allow List has private subnets set - Remove them?
    PopupCodes.removePrivateSubnetsFromAllowlist => DecisionPopupMetadata(
      id: PopupCodes.removePrivateSubnetsFromAllowlist,
      title: t.ui.removePrivateSubnets,
      message: (_) => t.ui.removePrivateSubnetsDescription,
      noButtonText: t.ui.cancel,
      yesButtonText: t.ui.continueWord,
      yesAction: (ref) async {
        await ref
            .read(vpnSettingsControllerProvider.notifier)
            .setLanDiscovery(true);
      },
    ),

    // ==============================    [ triggered by daemon ]    ==============================

    // Subscription expired
    DaemonStatusCode.accountExpired => RichPopupMetadata(
      id: DaemonStatusCode.accountExpired,
      header: t.ui.subscriptionHasEnded,
      message: (ref) {
        final account = ref.read(accountControllerProvider).valueOrNull;
        // if account is not set, we'll be redirected to login
        // page so the message returned here doesn't matter
        if (account == null) return "";

        return t.ui.pleaseRenewYourSubscription(email: account.email);
      },
      actionButtonText: t.ui.renewSubscription,
      image: DynamicThemeImage("subscription_ended.svg"),
      action: (_) async => await renewSubscriptionUrl.launch(),
      autoClose: false,
    ),

    // Private subnet can't be added when LAN discovery is on
    DaemonStatusCode.privateSubnetLANDiscovery => DecisionPopupMetadata(
      id: PopupCodes.removePrivateSubnetsFromAllowlist,
      title: t.ui.privateSubnetCantBeAdded,
      message: (_) => t.ui.privateSubnetCantBeAddedDescription,
      noButtonText: t.ui.close,
      yesButtonText: t.ui.turnOffLanDiscovery,
      yesAction: (ref) async {
        await ref
            .read(vpnSettingsControllerProvider.notifier)
            .setLanDiscovery(false);
      },
    ),

    // Settings not saved - configuration error
    DaemonStatusCode.configError => InfoPopupMetadata(
      id: DaemonStatusCode.configError,
      title: t.ui.settingsWereNotSaved,
      message: (_) => t.ui.couldNotSave,
    ),

    // not matched, display generic error message
    _ => infoForDaemonCode(code),
  };
  return metadata;
}

PopupMetadata infoForDaemonCode(int code) {
  var title = titleForCode(code);
  var message = messageForCode(code);

  if (title.isEmpty) {
    logger.e("Missing error message for $code");
    title = t.daemon.genericErrorTitle;
  }
  if (message.isEmpty) {
    if (kDebugMode) {
      message = "${t.daemon.genericErrorMessage} [Code $code]";
    } else {
      message = t.daemon.genericErrorMessage;
    }
  }

  assert(title.isNotEmpty && message.isNotEmpty);
  return InfoPopupMetadata(id: code, title: title, message: (_) => message);
}

PopupMetadata _resetToDefaults(int code) {
  assert(
    code == PopupCodes.resetSettings ||
        code == PopupCodes.resetSettingsAndDisconnect,
    "code $code incorrect for reset to defaults popup",
  );
  final resetAndDisconnect = code == PopupCodes.resetSettingsAndDisconnect;
  return DecisionPopupMetadata(
    id: PopupCodes.resetSettings,
    title: t.ui.resetAllCustomSettings,
    message: (_) => resetAndDisconnect
        ? t.ui.resetAndDisconnectDesc
        : t.ui.resetSettingsAlertDescription,
    noButtonText: t.ui.cancel,
    yesButtonText: resetAndDisconnect
        ? t.ui.resetAndDisconnect
        : t.ui.resetSettings,
    yesAction: (ref) async {
      await ref.read(vpnSettingsControllerProvider.notifier).resetToDefaults();
      await ref.read(preferencesControllerProvider.notifier).resetToDefaults();
    },
  );
}
