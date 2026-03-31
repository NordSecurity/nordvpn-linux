import 'dart:async';
import 'package:flutter/material.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/theme/toast_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';

final class Toast extends StatefulWidget {
  const Toast({super.key, required this.duration});
  static const Duration _defaultTimerStep = Duration(seconds: 1);

  final Duration duration;

  @override
  State<Toast> createState() => _ToastState();
}

class _ToastState extends State<Toast> {
  late Duration _remainingTime;
  Timer? _timer;
  bool _isVisible = true;

 @override
  void initState() {
    super.initState();
    _remainingTime = widget.duration;

    void tick() {
      _remainingTime -= Toast._defaultTimerStep;
      final remainingSeconds = _remainingTime.inSeconds.clamp(0, widget.duration.inSeconds);
      logger.e("Toast, remaining time: ${_remainingTime.inSeconds} s");

      setState(() {
        // refresh the countdown
        if (remainingSeconds == 0) {
          _timer?.cancel();
        }
        _remainingTime = Duration(seconds: remainingSeconds);
      });
    }
    _timer = Timer.periodic(Toast._defaultTimerStep, (_) => tick());
  }

  @override
  Widget build(BuildContext context) {
    if (!_isVisible) return const SizedBox.shrink();

    final theme = context.toastTheme;
    final textScaler = MediaQuery.textScalerOf(context);
    return Container(
      width: textScaler.scale(theme.widgetWidth),
      height: textScaler.scale(theme.widgetHeight),
      decoration: BoxDecoration(
        borderRadius: theme.toastBorderRadius,
        color: theme.toastBackgroundColor,
        border: Border.all(width: theme.toastBorderWidth, color: theme.toastBorderColor),
      ),
      child: Container(
        padding: EdgeInsets.all(textScaler.scale(theme.toastSpacing)),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            // general spacing for row -> 4 in all directions
            // pause icon
            // spacing 2 between
            // heading with spacing 0.5 from top
            // spacing 2 between text and button
            // close button
            // close button spacing 5 all directions
            _buildPauseIcon(),
            _buildWidgetText(theme),
            _buildCloseButton(theme),
          ],
        ),
      ),
    );
  }

  Widget _buildPauseIcon() {
    return DynamicThemeImage("toast_pause_icon.svg");
  }

  Widget _buildWidgetText(ToastTheme theme) {
    final m = _remainingTime.inMinutes.remainder(60).toString().padLeft(2, '0');
    final s = _remainingTime.inSeconds.remainder(60).toString().padLeft(2, '0');
    return Text(
      t.ui.VPNResumesIn(minutes: m, seconds: s),
      style: theme.toastMessageTextStyle,
      textAlign: TextAlign.center,
    );
  }

  Widget _buildCloseButton(ToastTheme theme) {
    return Padding(
      padding: theme.toastCloseButtonPadding,
      child: GestureDetector(
        onTap: () {
          setState(() {
            _isVisible = false;
          });
        },
        child: DynamicThemeImage("toast_close_icon.svg"),
      )
    );
  }

  @override
  void dispose() {
    logger.e("Toast::dispose called!");
    if (_timer != null) {
      _timer!.cancel();
    }
    super.dispose();
  }
}

