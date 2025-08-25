import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/providers/popups_provider.dart';
import 'package:nordvpn/data/providers/vpn_settings_controller.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/popup_codes.dart';
import 'package:nordvpn/internal/scaler_responsive_box.dart';
import 'package:nordvpn/settings/settings_wrapper_widget.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/custom_dns_theme.dart';
import 'package:nordvpn/theme/settings_theme.dart';
import 'package:nordvpn/widgets/advanced_list_tile.dart';
import 'package:nordvpn/widgets/bin_button.dart';
import 'package:nordvpn/widgets/custom_error_widget.dart';
import 'package:nordvpn/widgets/loading_button.dart';
import 'package:nordvpn/widgets/loading_indicator.dart';
import 'package:nordvpn/widgets/on_off_switch.dart';
import 'package:nordvpn/widgets/input.dart';

// Widget to manage the custom DNS servers list
final class CustomDns extends ConsumerStatefulWidget {
  const CustomDns({super.key});

  @override
  ConsumerState<CustomDns> createState() => _CustomDnsState();
}

// this are used for tests to easier identify the elements
final class CustomDnsKeys {
  CustomDnsKeys._();
  static final onOffSwitch = UniqueKey();
  static final addDnsForm = UniqueKey();
  static final serversList = UniqueKey();
  static ValueKey removeButton(String server) => ValueKey("remove_$server");
}

enum _CustomDnsSections { status, addForm, dnsList }

class _CustomDnsState extends ConsumerState<CustomDns> {
  final _controller = TextEditingController();
  // used to trigger a tap for onSubmitted from text field is called
  final _buttonController = LoadingButtonController();

  @override
  void initState() {
    super.initState();
    final settings = ref.read(vpnSettingsControllerProvider).valueOrNull;
    if (settings == null || settings.customDnsServers.isNotEmpty) return;
    // Custom DNS servers are not set and user just opened this page - switch
    // Custom DNS setting to false so the OnOffSwitch is off.
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(vpnSettingsControllerProvider.notifier).setCustomDns(false);
    });
  }

  @override
  Widget build(BuildContext context) {
    return ref
        .watch(vpnSettingsControllerProvider)
        .when(
          loading: () => const LoadingIndicator(),
          error: (error, stackTrace) => CustomErrorWidget(message: "$error"),
          data: (settings) => _build(context, settings),
        );
  }

  Widget _build(BuildContext context, ApplicationSettings settings) {
    final customDnsTheme = context.customDnsTheme;
    final settingsTheme = context.settingsTheme;

    // NOTE: If the toggle was changed in UI, then it stays on even when user
    // clears Custom DNS by using daemon
    final isCustomDnsEnabled =
        settings.customDns || settings.customDnsServers.isNotEmpty;

    return SettingsWrapperWidget(
      useSeparator: false,
      itemsCount: _CustomDnsSections.values.length,
      itemBuilder: (context, index) {
        switch (_CustomDnsSections.values[index]) {
          case _CustomDnsSections.status:
            return SettingsWrapperWidget.buildListItem(
              context,
              title: t.ui.useCustomDns,
              subtitle: t.ui.useCustomDnsDescription,
              trailingLocation: TrailingLocation.center,
              trailing: OnOffSwitch(
                key: CustomDnsKeys.onOffSwitch,
                value: isCustomDnsEnabled,
                shouldChange: (toValue) => _canChange(settings, toValue),
                onChanged: (value) => _toggleCustomDns(value),
              ),
            );
          case _CustomDnsSections.addForm:
            return SettingsWrapperWidget.buildListItem(
              key: CustomDnsKeys.addDnsForm,
              context,
              color: customDnsTheme.formBackground,
              title: "${t.ui.enterDnsAddress}:",
              titleStyle: settingsTheme.itemSubtitleStyle,
              subtitleWidget: ScalerResponsiveBox(
                maxWidth: customDnsTheme.dnsInputWidth,
                child: Input(
                  key: UniqueKey(),
                  hintText: "0.0.0.0",
                  submitDisplay: SubmitDisplay.never,
                  controller: _controller,
                  validateInput: (value) => _isValid(settings),
                  onSubmitted: (value) async {
                    _buttonController.triggerTap();
                  },
                  submitText: t.ui.add,
                  onErrorMessage: (value) => _errorMessage(value, settings),
                ),
              ),
              trailingLocation: TrailingLocation.center,
              trailing: ValueListenableBuilder<TextEditingValue>(
                valueListenable: _controller,
                builder: (context, value, child) {
                  final isValid = _isValid(settings);
                  return LoadingTextButton(
                    key: ValueKey(isValid),
                    controller: _buttonController,
                    onPressed: isValid ? () => _addServer(settings) : null,
                    child: Text(t.ui.add),
                  );
                },
              ),
              enabled:
                  isCustomDnsEnabled &&
                  settings.customDnsServers.length < maxCustomDnsServers,
            );
          case _CustomDnsSections.dnsList:
            if (settings.customDnsServers.isEmpty) {
              return SizedBox.shrink();
            } else {
              return AdvancedListTile(
                title: Expanded(child: _buildServersList(context, settings)),
              );
            }
        }
      },
    );
  }

  Widget _buildServersList(BuildContext context, ApplicationSettings settings) {
    final appTheme = context.appTheme;
    final customDnsTheme = context.customDnsTheme;

    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(t.ui.customDns, style: appTheme.captionStrong),
        Divider(
          height: appTheme.verticalSpaceSmall,
          color: customDnsTheme.dividerColor,
        ),
        ListView.builder(
          key: CustomDnsKeys.serversList,
          shrinkWrap: true,
          itemCount: settings.customDnsServers.length,
          itemBuilder: (context, index) {
            final server = settings.customDnsServers[index];
            return Row(
              children: [
                Text(server, style: appTheme.body),
                Spacer(),
                BinButton(
                  key: CustomDnsKeys.removeButton(server),
                  onPressed: () =>
                      _deleteServer(settings.customDnsServers[index]),
                ),
              ],
            );
          },
        ),
      ],
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    _buttonController.disable();
    super.dispose();
  }

  bool _isAddressValid(String value) {
    return ipv4Regex.hasMatch(value);
  }

  String _errorMessage(String value, ApplicationSettings settings) {
    if (settings.customDnsServers.contains(value)) {
      return t.ui.duplicatedDnsServer;
    }
    if (!_isAddressValid(value)) {
      return t.ui.invalidFormat;
    }
    return "";
  }

  bool _isValid(ApplicationSettings settings) {
    final value = _controller.text;
    return value.isNotEmpty &&
        _isAddressValid(value) &&
        !settings.customDnsServers.contains(value);
  }

  Future<void> _addServer(ApplicationSettings settings) async {
    if (!_isValid(settings)) return;

    final res = await ref
        .read(vpnSettingsControllerProvider.notifier)
        .addCustomDns(_controller.text);

    if (res == DaemonStatusCode.success) {
      _controller.clear();
    }
  }

  Future<void> _deleteServer(String server) async {
    final res = await ref
        .read(vpnSettingsControllerProvider.notifier)
        .removeCustomDns(server);

    if (res == DaemonStatusCode.success) {
      _controller.clear();
    }
  }

  Future<bool> _canChange(ApplicationSettings settings, bool toValue) async {
    // when user tries to disable it and Custom DNS is not empty, show
    // popup with warning and don't allow to switch to off here (it will
    // be done in popup)
    if (!toValue && settings.customDnsServers.isNotEmpty) {
      ref.read(popupsProvider.notifier).show(PopupCodes.turnOffCustomDns);
      return false;
    }

    // when user tries to enable it, but threat protection is on, we need to
    // disable TP first - ask the user and don't allow to switch custom DNS
    // here (it will be done in popup)
    if (toValue && settings.threatProtection) {
      ref
          .read(popupsProvider.notifier)
          .show(PopupCodes.turnOffThreatProtection);
      return false;
    }

    // allow to switch only when custom DNS is empty and no conflict with TP
    return true;
  }

  Future<void> _toggleCustomDns(bool value) async {
    ref.read(vpnSettingsControllerProvider.notifier).setCustomDns(value);
    setState(() {
      if (!value) {
        _controller.clear();
      }
    });
  }
}
