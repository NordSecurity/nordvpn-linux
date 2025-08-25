import 'dart:async';

import 'package:flutter/material.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/delayed_loading_manager.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/on_off_switch_theme.dart';
import 'package:nordvpn/widgets/enabled_widget.dart';
import 'package:nordvpn/widgets/inline_loading_indicator.dart';

final class OnOffSwitch extends StatefulWidget {
  final bool value;
  // Called before changing the value to check if it should change. When provided,
  // this is called before `onChanged` is executed. If this returns
  // false then the value is not changed and `onChanged` is not executed.
  // `toValue` is the value to which it should change.
  final Future<bool> Function(bool toValue)? shouldChange;

  // When onChanged is null the widget is disabled
  final Future<void> Function(bool)? onChanged;

  const OnOffSwitch({
    super.key,
    required this.onChanged,
    this.value = false,
    this.shouldChange,
  });

  @override
  OnOffSwitchState createState() => OnOffSwitchState();
}

final class OnOffSwitchState extends State<OnOffSwitch> {
  bool _isSwitched = false;
  bool _isDisabled = false;
  late DelayedLoadingManager _loadingManager;

  @override
  void initState() {
    super.initState();
    _isSwitched = widget.value;
    _isDisabled = widget.onChanged == null;
    _loadingManager = DelayedLoadingManager(
      onUpdate: _rebuild,
      onDone: () => setState(() => _isSwitched = !_isSwitched),
    );
  }

  void _rebuild() {
    if (mounted) setState(() {});
  }

  @override
  Widget build(BuildContext context) {
    final appTheme = context.appTheme;
    final switchTheme = context.onOffSwitchTheme;

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        _label(appTheme, switchTheme),
        Stack(
          alignment: Alignment.center,
          children: [
            if (_loadingManager.isLoading) InlineLoadingIndicator(),
            EnabledWidget(
              enabled: !_loadingManager.isLoading,
              child: _thumbSlider(switchTheme),
            ),
          ],
        ),
      ],
    );
  }

  Widget _label(AppTheme appTheme, OnOffSwitchTheme switchTheme) {
    return Padding(
      padding: EdgeInsets.only(right: switchTheme.label.paddingRight),
      child: SizedBox(
        child: Text(
          textAlign: TextAlign.right,
          _isSwitched ? t.ui.on : t.ui.off,
          style: _isDisabled
              ? switchTheme.label.disabledTextStyle
              : switchTheme.label.textStyle,
        ),
      ),
    );
  }

  Widget _thumbSlider(OnOffSwitchTheme switchTheme) {
    return GestureDetector(
      onTap: _isDisabled ? null : _toggle,
      child: AnimatedContainer(
        duration: animationDuration,
        width: switchTheme.slider.width,
        height: switchTheme.slider.height,
        decoration: _buildBackground(switchTheme),
        child: Stack(
          children: [
            AnimatedPositioned(
              duration: animationDuration,
              curve: Curves.easeIn,
              left: _isSwitched
                  ? switchTheme.slider.on.leftOffset
                  : switchTheme.slider.off.leftOffset,
              right: _isSwitched
                  ? switchTheme.slider.on.rightOffset
                  : switchTheme.slider.off.rightOffset,
              bottom: switchTheme.slider.bottomOffset,
              top: switchTheme.slider.topOffset,
              child: AnimatedContainer(
                duration: animationDuration,
                decoration: _buildSlider(switchTheme),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _toggle() async {
    assert(widget.onChanged != null);
    // if `widget.shouldChange` was not provided, default to always change
    final shouldChangeFn = widget.shouldChange ?? _yes;
    final shouldChange = await shouldChangeFn(!_isSwitched);
    if (!shouldChange) {
      logger.d("should change returned false, ignoring");
      return;
    }

    _loadingManager.startLoading();
    widget.onChanged!(!_isSwitched)
        .then((_) {
          _loadingManager.stopLoading(false);
        })
        .catchError((e) {
          logger.e("error while switching toggle: $e");
          _loadingManager.stopLoading(true);
        });
  }

  BoxDecoration _buildBackground(OnOffSwitchTheme switchTheme) {
    return BoxDecoration(
      border: Border.all(color: _borderColor(switchTheme)),
      borderRadius: BorderRadius.circular(switchTheme.slider.borderRadius),
      color: _backgroundColor(switchTheme),
    );
  }

  Color _borderColor(OnOffSwitchTheme switchTheme) {
    if (_isDisabled) {
      return _isSwitched
          ? switchTheme.slider.disabledOn.borderColor
          : switchTheme.slider.disabledOff.borderColor;
    }
    return _isSwitched
        ? switchTheme.slider.on.borderColor
        : switchTheme.slider.off.borderColor;
  }

  Color _backgroundColor(OnOffSwitchTheme switchTheme) {
    if (_isDisabled) {
      return _isSwitched
          ? switchTheme.slider.disabledOn.backgroundColor
          : switchTheme.slider.disabledOff.backgroundColor;
    }
    return _isSwitched
        ? switchTheme.slider.on.backgroundColor
        : switchTheme.slider.off.backgroundColor;
  }

  BoxDecoration _buildSlider(OnOffSwitchTheme switchTheme) {
    return BoxDecoration(
      shape: BoxShape.circle,
      color: _sliderColor(switchTheme),
    );
  }

  Color _sliderColor(OnOffSwitchTheme switchTheme) {
    if (_isDisabled) {
      return _isSwitched
          ? switchTheme.slider.disabledOn.color
          : switchTheme.slider.disabledOff.color;
    }
    return _isSwitched
        ? switchTheme.slider.on.color
        : switchTheme.slider.off.color;
  }

  @override
  void dispose() {
    _loadingManager.dispose();
    super.dispose();
  }
}

Future<bool> _yes(_) => Future.value(true);
