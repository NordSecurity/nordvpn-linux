import 'dart:async';

import 'package:flutter/material.dart';
import 'package:nordvpn/data/models/allow_list.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/settings/allow_list/allow_list_helpers.dart';
import 'package:nordvpn/theme/allow_list_theme.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/input_theme.dart';
import 'package:nordvpn/widgets/input.dart';
import 'package:nordvpn/widgets/loading_button.dart';

// Display the fields needed to add a subnet into the allow list
final class AddSubnet extends StatefulWidget {
  final AllowList allowList;
  final FutureOr<bool> Function(Subnet subnet) onSubmitted;

  const AddSubnet({
    super.key,
    required this.allowList,
    required this.onSubmitted,
  });

  @override
  State<AddSubnet> createState() => _AddSubnetState();
}

class _AddSubnetState extends State<AddSubnet> {
  final _textController = TextEditingController();
  final _buttonController = LoadingButtonController();

  @override
  void dispose() {
    _textController.dispose();
    _buttonController.disable();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final appTheme = context.appTheme;
    final allowListTheme = context.allowListTheme;

    final isStartValid = isSubnetValid(_textController.text);
    final isDuplicated = isStartValid && _isSubnetInAllowList();
    final isButtonEnabled = isStartValid && !isDuplicated;

    return Stack(
      alignment: AlignmentDirectional.bottomStart,
      children: [
        Row(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisAlignment: MainAxisAlignment.start,
          spacing: appTheme.verticalSpaceMedium,
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                spacing: appTheme.verticalSpaceSmall,
                children: [
                  Text(
                    "${t.ui.enterSubnet}:",
                    style: allowListTheme.labelStyle,
                  ),
                  Row(
                    spacing: appTheme.verticalSpaceLarge,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Expanded(
                        child: Input(
                          controller: _textController,
                          hintText: "192.168.1.0/24",
                          validateInput: (_) => isStartValid && !isDuplicated,
                          onErrorMessage: _errorForPort,
                          submitDisplay: SubmitDisplay.never,
                          onChanged: (value) => setState(() {}),
                          onSubmitted: (value) =>
                              _buttonController.triggerTap(),
                        ),
                      ),
                      LoadingTextButton(
                        key: ValueKey(isButtonEnabled),
                        controller: _buttonController,
                        onPressed: isButtonEnabled ? _addSubnet : null,
                        child: Text(t.ui.add),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
        if (isDuplicated)
          Text(t.ui.subnetAlreadyInList, style: context.inputTheme.errorStyle),
      ],
    );
  }

  String _errorForPort(String value) {
    if (!isSubnetValid(value)) {
      return t.ui.invalidFormat;
    }
    if (_isSubnetInAllowList()) {
      return " ";
    }
    return "";
  }

  bool _isSubnetInAllowList() {
    try {
      final subnet = Subnet.fromString(_textController.text);

      for (final s in widget.allowList.subnets) {
        if (s.contains(subnet)) {
          logger.d("found $s contains $subnet");
          return true;
        }
      }
    } catch (_) {}

    return false;
  }

  Future<void> _addSubnet() async {
    try {
      final subnet = Subnet.fromString(_textController.text);

      if (await widget.onSubmitted(subnet)) {
        setState(() {
          _textController.clear();
        });
      }
    } catch (e) {
      logger.e("failed to add subnet $e");
    }
  }
}
