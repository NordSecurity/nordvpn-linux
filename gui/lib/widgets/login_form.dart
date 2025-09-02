import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/config.dart';
import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/providers/account_controller.dart';
import 'package:nordvpn/data/providers/vpn_settings_controller.dart';

import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/login_form_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/loading_button.dart';
import 'package:nordvpn/widgets/loading_checkbox.dart';

final class LoginForm extends ConsumerStatefulWidget {
  const LoginForm({super.key});

  @override
  ConsumerState<LoginForm> createState() => _LoginFormState();
}

class _LoginFormState extends ConsumerState<LoginForm> {
  bool _checkboxValue = false;
  bool _isCheckboxVisible = false;

  bool _killSwitchTurnedOffViaGui = false;
  bool _wasKillSwitchOn = false;
  bool _isKillSwitchOn = false;

  @override
  Widget build(BuildContext context) {
    final settingsProvider = ref.watch(vpnSettingsControllerProvider);

    _updateCheckboxVisibilityState(settingsProvider.valueOrNull);

    final loginDialogTheme = context.loginFormTheme;
    final appTheme = context.appTheme;
    return Padding(
      padding: EdgeInsets.all(appTheme.padding),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          DynamicThemeImage("nordvpn_logo_big.svg"),
          SizedBox(height: appTheme.verticalSpaceMedium),
          _title(loginDialogTheme),
          SizedBox(height: appTheme.verticalSpaceLarge),
          if (_isCheckboxVisible) _killSwitchCheckbox(ref),
          if (_isCheckboxVisible)
            SizedBox(height: appTheme.verticalSpaceMedium),
          if (_isCheckboxVisible) _checkboxDescription(loginDialogTheme),
          if (_isCheckboxVisible) SizedBox(height: appTheme.verticalSpaceLarge),
          IntrinsicWidth(
            child: Column(
              children: [
                LoginButton(isDisabled: _isKillSwitchOn),
                SizedBox(height: appTheme.verticalSpaceMedium),
                _createAccountButton(loginDialogTheme),
              ],
            ),
          ),
        ],
      ),
    );
  }

  // When KS was turned off, check if it was turned off in the GUI.
  // If yes, keep it visible. If it was turned off outside of the GUI,
  // hide the checkbox.
  // When KS was turned on, show the checkbox and reset it's value
  // (no matter if it was turned on in GUI or outside).
  void _updateCheckboxVisibilityState(ApplicationSettings? settings) {
    _isKillSwitchOn = settings?.killSwitch ?? false;
    if (_wasKillSwitchOn && !_isKillSwitchOn) {
      // KS was turned off
      if (!_killSwitchTurnedOffViaGui) {
        // turned off, but through CLI - hide checkbox
        _isCheckboxVisible = false;
      } else {
        // turned off in GUI - checkbox should still be visible
        _killSwitchTurnedOffViaGui = false;
        _isCheckboxVisible = true;
      }
    } else if (!_wasKillSwitchOn && _isKillSwitchOn) {
      // KS was turned on
      _isCheckboxVisible = true;
      _checkboxValue = false;
    }
    _wasKillSwitchOn = _isKillSwitchOn;
  }

  Widget _title(LoginFormTheme loginDialogTheme) {
    return Text(t.ui.loginTitle, style: loginDialogTheme.titleStyle);
  }

  Widget _killSwitchCheckbox(WidgetRef ref) {
    return LoadingCheckbox(
      value: _checkboxValue,
      text: t.ui.turnOffKillSwitch,
      onChanged: (value) async {
        setState(() {
          _checkboxValue = value;
          _killSwitchTurnedOffViaGui = value;
        });
        await ref
            .read(vpnSettingsControllerProvider.notifier)
            .setKillSwitch(!value);
      },
    );
  }

  Widget _checkboxDescription(LoginFormTheme theme) {
    return Text(
      t.ui.turnOffKillSwitchDescription,
      style: theme.checkboxDescStyle,
    );
  }

  Widget _createAccountButton(LoginFormTheme loginDialogTheme) {
    return LoadingOutlinedButton(
      onPressed: _isKillSwitchOn
          ? null
          : () async =>
                await ref.read(accountControllerProvider.notifier).register(),
      child: SizedBox(
        width: double.infinity, // force it to take all parent's width
        child: Center(child: Text(t.ui.createAccount)),
      ),
    );
  }
}

final class LoginButton extends ConsumerStatefulWidget {
  final bool isDisabled;
  final Config config;

  LoginButton({super.key, this.isDisabled = false, Config? config})
    : config = config ?? sl();

  @override
  ConsumerState<LoginButton> createState() => _LoginButtonState();
}

final class _LoginButtonState extends ConsumerState<LoginButton> {
  @override
  Widget build(BuildContext context) {
    final theme = context.loginFormTheme;

    final userAccount = ref.watch(accountControllerProvider);
    final isLoading = userAccount is AsyncLoading;

    return ElevatedButton(
      onPressed: isLoading || widget.isDisabled ? null : loginWithTimeout,
      child: SizedBox(
        width: double.infinity, // force it to take all parent's width
        child: Center(
          child: isLoading
              ? _progressIndicator(theme.progressIndicator)
              : Text(t.ui.logIn),
        ),
      ),
    );
  }

  void loginWithTimeout() {
    ref.read(accountControllerProvider.notifier).login();
  }

  Widget _progressIndicator(LoginButtonProgressIndicatorTheme theme) {
    return SizedBox(
      width: theme.width,
      height: theme.height,
      child: CircularProgressIndicator(
        color: theme.color,
        strokeWidth: theme.stroke,
      ),
    );
  }

  @override
  void dispose() {
    super.dispose();
  }
}
