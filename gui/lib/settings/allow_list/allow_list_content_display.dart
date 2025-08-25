import 'dart:async';

import 'package:flutter/material.dart';
import 'package:nordvpn/data/models/allow_list.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/settings/allow_list/allow_list_helpers.dart';
import 'package:nordvpn/theme/allow_list_theme.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/widgets/bin_button.dart';

// Display the content of allow list
final class AllowListContentDisplay extends StatelessWidget {
  final AllowList allowList;
  final FutureOr<void> Function({PortInterval? port, Subnet? subnet}) onDeleted;

  const AllowListContentDisplay({
    super.key,
    required this.allowList,
    required this.onDeleted,
  });

  @override
  Widget build(BuildContext context) {
    if (allowList.isEmpty) {
      return SizedBox.shrink();
    }

    final singlePorts = <PortInterval>[];
    final portRanges = <PortInterval>[];

    for (final p in allowList.ports) {
      if (p.isRange) {
        portRanges.add(p);
      } else {
        singlePorts.add(p);
      }
    }

    final appTheme = context.appTheme;
    return SingleChildScrollView(
      padding: EdgeInsets.symmetric(horizontal: appTheme.verticalSpaceSmall),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.start,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        spacing: appTheme.verticalSpaceMedium,
        children: [
          if (singlePorts.isNotEmpty) _buildSinglePorts(context, singlePorts),
          if (portRanges.isNotEmpty) _buildRangePorts(context, portRanges),
          if (allowList.subnets.isNotEmpty)
            _buildSubnets(context, allowList.subnets),
        ],
      ),
    );
  }

  Widget _buildSinglePorts(BuildContext context, List<PortInterval> ports) {
    return Column(
      children: [
        _header(context, [t.ui.port, t.ui.protocol]),
        Divider(color: context.allowListTheme.dividerColor),
        _buildPortsList(ports),
      ],
    );
  }

  Widget _buildRangePorts(BuildContext context, List<PortInterval> ports) {
    return Column(
      children: [
        _header(context, [t.ui.portRange, t.ui.protocol]),
        Divider(color: context.allowListTheme.dividerColor),
        _buildPortsList(ports),
      ],
    );
  }

  Widget _buildSubnets(BuildContext context, List<Subnet> subnets) {
    final allowListTheme = context.allowListTheme;

    return Column(
      children: [
        _header(context, [t.ui.subnet]),
        ListView.builder(
          physics: NeverScrollableScrollPhysics(),
          shrinkWrap: true,
          itemCount: subnets.length,
          itemBuilder: (context, index) {
            final subnet = subnets[index];
            return Container(
              color: (index % 2 == 0)
                  ? null
                  : allowListTheme.listItemBackgroundColor,
              padding: EdgeInsets.all(context.appTheme.verticalSpaceSmall),
              child: Row(
                children: [
                  Expanded(
                    child: Text(
                      subnet.value,
                      style: allowListTheme.tableItemsStyle,
                    ),
                  ),
                  Spacer(),
                  BinButton(
                    onPressed: () async => await onDeleted(subnet: subnet),
                  ),
                ],
              ),
            );
          },
        ),
      ],
    );
  }

  Widget _header(BuildContext context, List<String> rows) {
    final allowListTheme = context.allowListTheme;
    return Row(
      children: [
        for (final name in rows)
          Expanded(child: Text(name, style: allowListTheme.tableHeaderStyle)),
      ],
    );
  }

  ListView _buildPortsList(List<PortInterval> ports) {
    return ListView.builder(
      physics: NeverScrollableScrollPhysics(),
      shrinkWrap: true,
      itemCount: ports.length,
      itemBuilder: (context, index) {
        final port = ports[index];
        final allowListTheme = context.allowListTheme;

        return Container(
          color: (index % 2 == 0)
              ? null
              : allowListTheme.listItemBackgroundColor,
          padding: EdgeInsets.all(context.appTheme.verticalSpaceSmall),
          child: Row(
            children: [
              Expanded(child: _buildPort(context, port)),
              Expanded(
                child: Row(
                  children: [
                    Text(
                      portTypeToString(port.type),
                      style: allowListTheme.tableItemsStyle,
                    ),
                    Spacer(),
                    BinButton(
                      onPressed: () async => await onDeleted(port: port),
                    ),
                  ],
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  Widget _buildPort(BuildContext context, PortInterval port) {
    final allowListTheme = context.allowListTheme;
    if (port.isRange) {
      return Row(
        spacing: context.appTheme.verticalSpaceMedium,
        children: [
          Text("${port.start}", style: allowListTheme.tableItemsStyle),
          Text(t.ui.to, style: allowListTheme.tableItemsStyle),
          Text("${port.end}", style: allowListTheme.tableItemsStyle),
        ],
      );
    }
    return Text("${port.start}", style: allowListTheme.tableItemsStyle);
  }
}
