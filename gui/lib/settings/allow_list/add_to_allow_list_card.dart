import 'dart:async';

import 'package:flutter/material.dart';
import 'package:nordvpn/data/models/allow_list.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/settings/allow_list/add_port.dart';
import 'package:nordvpn/settings/allow_list/add_subnet.dart';
import 'package:nordvpn/theme/allow_list_theme.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/widgets/enabled_widget.dart';
import 'package:nordvpn/widgets/radio_button.dart';
import 'package:nordvpn/widgets/round_container.dart';

// Display the card to add a port, port range or subnet to allow list
final class AddToAllowListCard extends StatefulWidget {
  final bool enabled;
  final AllowList allowList;
  final FutureOr<bool> Function({PortInterval? port, Subnet? subnet})
  onSubmitted;

  const AddToAllowListCard({
    super.key,
    required this.enabled,
    required this.allowList,
    required this.onSubmitted,
  });

  @override
  State<AddToAllowListCard> createState() => _AddToAllowListCardState();
}

enum _Add { port, portRange, subnet }

class _AddToAllowListCardState extends State<AddToAllowListCard> {
  var _addType = _Add.port;

  @override
  Widget build(BuildContext context) {
    final appTheme = context.appTheme;
    final allowListTheme = context.allowListTheme;

    return EnabledWidget(
      enabled: widget.enabled,
      disabledOpacity: appTheme.disabledOpacity,
      child: RoundContainer(
        radius: appTheme.borderRadiusSmall,
        color: allowListTheme.addCardBackground,
        margin: EdgeInsets.zero,
        padding: EdgeInsets.all(appTheme.verticalSpaceSmall),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          mainAxisAlignment: MainAxisAlignment.start,
          crossAxisAlignment: CrossAxisAlignment.start,
          spacing: appTheme.verticalSpaceSmall,
          children: [
            Row(
              spacing: appTheme.verticalSpaceMedium,
              children: [
                RadioButton(
                  value: _Add.port,
                  groupValue: _addType,
                  onChanged: (value) => _changeAddType(value),
                  label: t.ui.port,
                ),
                RadioButton(
                  value: _Add.portRange,
                  groupValue: _addType,
                  onChanged: (value) => _changeAddType(value),
                  label: t.ui.portRange,
                ),
                RadioButton(
                  value: _Add.subnet,
                  groupValue: _addType,
                  onChanged: (value) => _changeAddType(value),
                  label: t.ui.subnet,
                ),
              ],
            ),
            if (_addType != _Add.subnet)
              AddPort(
                key: ValueKey(_addType),
                allowList: widget.allowList,
                addPortRange: _addType == _Add.portRange,
                onSubmitted: (port) => widget.onSubmitted(port: port),
              ),
            if (_addType == _Add.subnet)
              AddSubnet(
                allowList: widget.allowList,
                onSubmitted: (subnet) => widget.onSubmitted(subnet: subnet),
              ),
          ],
        ),
      ),
    );
  }

  void _changeAddType(_Add value) {
    setState(() {
      _addType = value;
    });
  }
}
