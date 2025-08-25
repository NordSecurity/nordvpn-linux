import 'dart:async';

import 'package:flutter/material.dart';
import 'package:nordvpn/data/models/allow_list.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/settings/allow_list/allow_list_helpers.dart';
import 'package:nordvpn/theme/allow_list_theme.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/input_theme.dart';
import 'package:nordvpn/widgets/dropdown.dart';
import 'package:nordvpn/widgets/input.dart';
import 'package:nordvpn/widgets/loading_button.dart';

// Display the fields needed to add a port or a port range into the allow list
final class AddPort extends StatefulWidget {
  final AllowList allowList;
  // When a port range needs to be added it will be set to true
  final bool addPortRange;
  final FutureOr<bool> Function(PortInterval port) onSubmitted;

  const AddPort({
    super.key,
    required this.allowList,
    required this.addPortRange,
    required this.onSubmitted,
  });

  @override
  State<AddPort> createState() => _AddPortState();
}

enum _ErrorType { none, start, end, duplicate, invalidRange }

class _AddPortState extends State<AddPort> {
  PortType _portType = PortType.both;
  final _textControllerStart = TextEditingController();
  final _textControllerEnd = TextEditingController();
  final _buttonController = LoadingButtonController();

  @override
  void dispose() {
    _buttonController.disable();
    _textControllerEnd.dispose();
    _textControllerStart.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final appTheme = context.appTheme;
    final allowListTheme = context.allowListTheme;

    final error = _calculateError();
    final isButtonEnabled =
        (error == _ErrorType.none) &&
        _textControllerStart.text.isNotEmpty &&
        (!widget.addPortRange || _textControllerEnd.text.isNotEmpty);

    return Stack(
      alignment: AlignmentDirectional.bottomStart,
      children: [
        Row(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisAlignment: MainAxisAlignment.start,
          spacing: appTheme.verticalSpaceMedium,
          children: [
            Flexible(
              flex: widget.addPortRange ? 5 : 4,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                spacing: appTheme.verticalSpaceSmall,
                children: [
                  Text(
                    "${widget.addPortRange ? t.ui.enterPortRange : t.ui.enterPort}:",
                    style: allowListTheme.labelStyle,
                  ),
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    spacing: appTheme.verticalSpaceMedium,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Expanded(
                        child: Input(
                          controller: _textControllerStart,
                          validateInput: (_) =>
                              (error == _ErrorType.none) ||
                              (error == _ErrorType.end),
                          onErrorMessage: _errorForPort,
                          submitDisplay: SubmitDisplay.never,
                          hintText: "0",
                          onChanged: (value) => setState(() {}),
                          onSubmitted: (value) =>
                              _buttonController.triggerTap(),
                        ),
                      ),
                      if (widget.addPortRange)
                        // add some padding to center with the text field
                        Padding(
                          padding: EdgeInsets.only(
                            top: appTheme.borderRadiusMedium,
                          ),
                          child: Text(t.ui.to, style: appTheme.body),
                        ),
                      if (widget.addPortRange)
                        Expanded(
                          child: Input(
                            controller: _textControllerEnd,
                            hintText: "0",
                            validateInput: (_) =>
                                (error == _ErrorType.none) ||
                                (error == _ErrorType.start),
                            onErrorMessage: _errorForPort,
                            submitDisplay: SubmitDisplay.never,
                            onChanged: (value) => setState(() {}),
                            onSubmitted: (value) =>
                                _buttonController.triggerTap(),
                          ),
                        ),
                    ],
                  ),
                ],
              ),
            ),
            Flexible(
              flex: 4,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                spacing: appTheme.verticalSpaceSmall,
                children: [
                  Text(
                    "${t.ui.selectProtocol}:",
                    style: allowListTheme.labelStyle,
                  ),
                  Row(
                    spacing: appTheme.verticalSpaceMedium,
                    children: [
                      Expanded(
                        child: Dropdown(
                          // set to unique to recreate it when _portType changes programmatically
                          key: UniqueKey(),
                          items: [
                            DropdownItem(value: PortType.both, label: t.ui.all),
                            DropdownItem(value: PortType.tcp, label: "TCP"),
                            DropdownItem(value: PortType.udp, label: "UDP"),
                          ],
                          initialValue: _portType,
                          showError: (error == _ErrorType.duplicate),
                          onChanged: (value) => setState(() {
                            _portType = value;
                          }),
                        ),
                      ),
                      LoadingTextButton(
                        key: ValueKey(isButtonEnabled),
                        controller: _buttonController,
                        onPressed: isButtonEnabled ? _addPort : null,
                        child: Text(t.ui.add),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
        if ((error == _ErrorType.duplicate) ||
            (error == _ErrorType.invalidRange))
          Text(
            widget.addPortRange
                ? ((error == _ErrorType.invalidRange)
                      ? t.ui.startPortBiggerThanEnd
                      : t.ui.portRangeAlreadyInList)
                : t.ui.portAlreadyInList,
            style: context.inputTheme.errorStyle,
          ),
      ],
    );
  }

  _ErrorType _calculateError() {
    if (_textControllerStart.text.isEmpty) {
      return _ErrorType.none;
    }
    if (widget.addPortRange && _textControllerEnd.text.isEmpty) {
      return _ErrorType.none;
    }

    if (!isValidPortNumber(_textControllerStart.text)) {
      return _ErrorType.start;
    }

    if (widget.addPortRange && !isValidPortNumber(_textControllerEnd.text)) {
      return _ErrorType.end;
    }

    if (widget.addPortRange && !_isRangeValid()) {
      return _ErrorType.invalidRange;
    }

    if (_isPortInAllowList()) {
      return _ErrorType.duplicate;
    }

    return _ErrorType.none;
  }

  String _errorForPort(String value) {
    if (!isValidPortNumber(value)) {
      return t.ui.invalidFormat;
    }

    if (_calculateError() != _ErrorType.none) {
      // send empty string for error to display only the border red
      return " ";
    }
    return "";
  }

  bool _isRangeValid() {
    return null != _constructPort();
  }

  bool _isPortInAllowList() {
    final port = _constructPort();
    if (port == null) {
      return true;
    }

    for (final range in widget.allowList.ports) {
      if (range.contains(port)) {
        return true;
      }
    }

    return false;
  }

  PortInterval? _constructPort() {
    final start = int.tryParse(_textControllerStart.text);
    if (start == null) {
      logger.e("failed to parse port $start");
      return null;
    }

    final end = widget.addPortRange
        ? int.tryParse(_textControllerEnd.text)
        : start;
    if (end == null) {
      logger.e("failed to parse port $end");
      return null;
    }

    if (start > end) {
      logger.e("invalid port range $start to $end");
      return null;
    }

    return PortInterval(start: start, end: end, type: _portType);
  }

  Future<void> _addPort() async {
    final port = _constructPort();
    if (port == null) {
      return;
    }

    if (await widget.onSubmitted(port)) {
      setState(() {
        _portType = PortType.both;
        _textControllerStart.clear();
        _textControllerEnd.clear();
      });
    }
  }
}
