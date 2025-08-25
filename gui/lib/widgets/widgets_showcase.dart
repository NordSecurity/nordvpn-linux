import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:grpc/grpc.dart';
import 'package:nordvpn/analytics/consent_screen.dart';
import 'package:nordvpn/data/providers/popups_provider.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/popup_codes.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/widgets/input.dart';
import 'package:nordvpn/widgets/loading_button.dart';
import 'package:nordvpn/widgets/loading_checkbox.dart';
import 'package:nordvpn/widgets/link.dart';
import 'package:nordvpn/widgets/on_off_switch.dart';
import 'package:nordvpn/widgets/radio_button.dart';

final class WidgetsShowcase extends ConsumerStatefulWidget {
  const WidgetsShowcase({super.key});

  @override
  ConsumerState<WidgetsShowcase> createState() => _WidgetsShowcaseState();
}

class _WidgetsShowcaseState extends ConsumerState<WidgetsShowcase> {
  int _groupValue = 0;

  @override
  Widget build(BuildContext context) {
    final theme = context.appTheme;
    return Container(
      color: theme.backgroundColor,
      child: SingleChildScrollView(
        child: Padding(
          padding: const EdgeInsets.all(8.0),
          child: Column(
            children: [
              Wrap(
                spacing: 10,
                runSpacing: 10,
                children: [
                  const SizedBox(height: 10),
                  OnOffSwitch(
                    onChanged: (val) async {
                      await Future.delayed(const Duration(milliseconds: 1000));
                      logger.i("switch changed to $val");
                    },
                  ),
                  LoadingCheckbox(
                    value: false,
                    text: "deselected, disabled",
                    onChanged: null,
                  ),
                  LoadingCheckbox(
                    value: true,
                    text: "selected, disabled",
                    onChanged: null,
                  ),
                  LoadingCheckbox(
                    value: false,
                    text: "deselected, enabled",
                    onChanged: (_) async {},
                  ),
                  LoadingCheckbox(
                    value: true,
                    text: "selected, enabled",
                    onChanged: (_) async {},
                  ),
                  LoadingCheckbox(
                    value: true,
                    text: "enabled, delayed",
                    onChanged: (_) async =>
                        await Future.delayed(Duration(seconds: 3)),
                  ),
                  const SizedBox(height: 10),
                  RadioButton(
                    value: 1,
                    groupValue: _groupValue,
                    onChanged: (val) => setState(() => _groupValue = val),
                    label: "NordLynx",
                  ),
                  RadioButton(
                    value: 2,
                    groupValue: _groupValue,
                    onChanged: (val) async {
                      await Future.delayed(Duration(seconds: 2));
                      setState(() => _groupValue = val);
                    },
                    label: "OpenVPN",
                  ),
                  Link(
                    title: "Normal link",
                    uri: Uri.parse("https://nordvpn.com"),
                  ),
                  Link(
                    title: "Small link",
                    uri: Uri.parse("https://nordvpn.com"),
                    size: LinkSize.small,
                  ),
                  const SizedBox(height: 10),
                  ElevatedButton(
                    onPressed: () => logger.i("elevated button pressed"),
                    child: const Text("Reconnect"),
                  ),
                  const SizedBox(height: 5),
                  OutlinedButton(
                    onPressed: () => logger.i("text button pressed"),
                    child: const Text("Cancel"),
                  ),
                  TextButton(
                    onPressed: () => logger.i("text button pressed"),
                    child: const Text("Manage"),
                  ),
                  const SizedBox(height: 5),
                  const Divider(),
                  const SizedBox(height: 5),
                  Input(
                    submitDisplay: SubmitDisplay.never,
                    onChanged: (value) => logger.i("input changed: $value"),
                    errorMessage: t.ui.invalidFormat,
                    validateInput: (value) =>
                        value.isEmpty || RegExp(r'^[a-zA-Z]+$').hasMatch(value),
                  ),
                  TextButton(
                    onPressed: () => logger.i("text button pressed"),
                    child: const Text("Manage"),
                  ),
                  LoadingTextButton(
                    child: Text("Loading"),
                    onPressed: () => Future.delayed(Duration(seconds: 2)),
                  ),
                  LoadingIconButton(
                    child: Icon(Icons.delete),
                    onPressed: () => Future.delayed(Duration(seconds: 2)),
                  ),
                  LoadingOutlinedButton(
                    child: Icon(Icons.delete),
                    onPressed: () => Future.delayed(Duration(seconds: 2)),
                  ),
                ],
              ),
              Row(
                spacing: 10,
                children: [
                  Expanded(
                    child: LoadingTextButton(
                      child: Text("Loading"),
                      onPressed: () => Future.delayed(Duration(seconds: 2)),
                    ),
                  ),
                  Expanded(
                    child: LoadingElevatedButton(
                      child: Text("Loading1"),
                      onPressed: () => Future.delayed(Duration(seconds: 2)),
                    ),
                  ),
                  Expanded(
                    child: LoadingElevatedButton(child: Text("Loading2")),
                  ),
                ],
              ),
              Text("popups", style: TextStyle(fontSize: 18)),
              SizedBox(height: 10),
              Column(
                spacing: 2,
                children: [
                  ElevatedButton(
                    onPressed: () => ConsentScreen(),
                    child: const Text("Consent screen"),
                  ),
                  ElevatedButton(
                    onPressed: () => ref
                        .read(popupsProvider.notifier)
                        .show(DaemonStatusCode.accountExpired),
                    child: const Text("Subscription expired"),
                  ),
                  ElevatedButton(
                    onPressed: () => ref
                        .read(popupsProvider.notifier)
                        .show(PopupCodes.getDedicatedIp),
                    child: const Text("Get your Dedicated IP"),
                  ),
                  ElevatedButton(
                    onPressed: () => ref
                        .read(popupsProvider.notifier)
                        .show(PopupCodes.chooseDip),
                    child: const Text("Choose your Dedicated IP"),
                  ),
                  ElevatedButton(
                    onPressed: () => ref
                        .read(popupsProvider.notifier)
                        .show(PopupCodes.turnOffCustomDns),
                    child: const Text("Turn off Custom DNS?"),
                  ),
                  ElevatedButton(
                    onPressed: () => ref
                        .read(popupsProvider.notifier)
                        .show(PopupCodes.turnOffAllowList),
                    child: const Text("Turn off Allow List?"),
                  ),
                  ElevatedButton(
                    onPressed: () => ref
                        .read(popupsProvider.notifier)
                        .show(PopupCodes.removePrivateSubnetsFromAllowlist),
                    child: const Text("Remove private subnets from allowlist?"),
                  ),
                  ElevatedButton(
                    onPressed: () => ref
                        .read(popupsProvider.notifier)
                        .show(DaemonStatusCode.privateSubnetLANDiscovery),
                    child: const Text("Turn off LAN Discovery?"),
                  ),
                  ElevatedButton(
                    onPressed: () => ref
                        .read(popupsProvider.notifier)
                        .show(PopupCodes.turnOffThreatProtection),
                    child: const Text("Turn off Threat Protection?"),
                  ),
                  ElevatedButton(
                    onPressed: () => ref
                        .read(popupsProvider.notifier)
                        .show(DaemonStatusCode.configError),
                    child: const Text("Settings not saved"),
                  ),
                  ElevatedButton(
                    onPressed: () => ref
                        .read(popupsProvider.notifier)
                        .show(DaemonStatusCode.failure),
                    child: const Text("Generic failure"),
                  ),
                  Input(
                    submitText: "Error Popup",
                    hintText: "Enter status code or error message",
                    submitDisplay: SubmitDisplay.always,
                    onSubmitted: _showErrorPopup,
                    errorMessage: t.ui.invalidFormat,
                    validateInput: (value) => value.isNotEmpty,
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _showErrorPopup(String value) {
    int status =
        int.tryParse(value) ??
        DaemonStatusCode.fromGrpcError(GrpcError.unknown(value));

    ref.read(popupsProvider.notifier).show(status);
  }
}
