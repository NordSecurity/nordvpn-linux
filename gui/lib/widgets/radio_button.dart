import 'dart:async';

import 'package:flutter/material.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/internal/delayed_loading_manager.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/radio_button_theme.dart';
import 'package:nordvpn/widgets/enabled_widget.dart';
import 'package:nordvpn/widgets/inline_loading_indicator.dart';

final class RadioButton<T> extends StatefulWidget {
  final T value;
  final T groupValue;
  final FutureOr<void> Function(T value) onChanged;
  final String label;
  final TextStyle? labelStyle;

  const RadioButton({
    super.key,
    required this.value,
    required this.groupValue,
    required this.onChanged,
    required this.label,
    this.labelStyle,
  });

  @override
  State<RadioButton> createState() => RadioButtonState<T>();
}

final class RadioButtonState<T> extends State<RadioButton<T>> {
  late DelayedLoadingManager _loadingManager;
  bool isSelected = false;
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _loadingManager = DelayedLoadingManager(
      onUpdate: () => _rebuild(true),
      onDone: () => _rebuild(false),
      onError: () => _rebuild(false),
    );
  }

  @override
  void dispose() {
    _loadingManager.dispose();
    super.dispose();
  }

  void _rebuild(bool value) {
    if (mounted) setState(() => _isLoading = value);
  }

  @override
  @override
  Widget build(BuildContext context) {
    final appTheme = Theme.of(context).extension<AppTheme>()!;
    final radioTheme = context.radioButtonTheme;
    isSelected = (widget.value == widget.groupValue);

    return GestureDetector(
      onTap: _onTap,
      child: Padding(
        padding: EdgeInsets.all(radioTheme.padding),
        child: Row(
          children: [
            Stack(
              alignment: Alignment.center,
              children: [
                if (_isLoading) InlineLoadingIndicator(),
                EnabledWidget(enabled: !_isLoading, child: _radio(radioTheme)),
              ],
            ),
            _label(radioTheme, appTheme),
          ],
        ),
      ),
    );
  }

  AnimatedContainer _radio(RadioButtonTheme radioTheme) {
    return AnimatedContainer(
      duration: animationDuration,
      width: radioTheme.radio.width,
      height: radioTheme.radio.height,
      decoration: _buildBackground(radioTheme),
      child: Center(
        child: AnimatedContainer(
          duration: animationDuration,
          width: isSelected
              ? radioTheme.radio.on.dotWidth
              : radioTheme.radio.off.dotWidth,
          height: isSelected
              ? radioTheme.radio.on.dotHeight
              : radioTheme.radio.off.dotHeight,
          decoration: _buildDot(radioTheme),
        ),
      ),
    );
  }

  Padding _label(RadioButtonTheme radioTheme, AppTheme appTheme) {
    return Padding(
      padding: EdgeInsets.only(left: radioTheme.label.paddingLeft),
      child: Text(widget.label, style: widget.labelStyle ?? appTheme.body),
    );
  }

  BoxDecoration _buildBackground(RadioButtonTheme radioTheme) {
    return BoxDecoration(
      color: isSelected
          ? radioTheme.radio.on.fillColor
          : radioTheme.radio.off.fillColor,
      shape: BoxShape.circle,
      border: Border.all(
        color: isSelected
            ? radioTheme.radio.on.borderColor
            : radioTheme.radio.off.borderColor,
        width: isSelected
            ? radioTheme.radio.on.borderWidth
            : radioTheme.radio.off.borderWidth,
      ),
    );
  }

  BoxDecoration _buildDot(RadioButtonTheme radioTheme) {
    return BoxDecoration(
      shape: BoxShape.circle,
      color: isSelected
          ? radioTheme.radio.on.dotColor
          : radioTheme.radio.off.dotColor,
    );
  }

  Future<void> _onTap() async {
    if (isSelected) {
      return;
    }
    _loadingManager.startLoading();
    try {
      await widget.onChanged(widget.value);
      _loadingManager.stopLoading(false);
    } catch (e) {
      logger.e("error while radio button pressed: $e");
      _loadingManager.stopLoading(true);
    }
  }
}
