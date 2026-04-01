import 'dart:async';
import 'package:flutter/material.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/theme/toast_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';

final class Toast extends StatefulWidget {
  const Toast({super.key, required this.duration, required this.onClose});
  static const Duration _defaultTimerStep = Duration(seconds: 1);

  final Duration duration;
  final VoidCallback? onClose;

  @override
  State<Toast> createState() => _ToastState();
}

final class _ToastState extends State<Toast> {
  late Duration _remainingTime;
  Timer? _timer;

  @override
  void initState() {
    super.initState();
    _remainingTime = widget.duration;

    void tick() {
      _remainingTime -= Toast._defaultTimerStep;
      final remainingSeconds = _remainingTime.inSeconds.clamp(
        0,
        widget.duration.inSeconds,
      );
      logger.d("Toast, remaining time: ${_remainingTime.inSeconds} s");

      if (remainingSeconds == 0) {
        _timer?.cancel();
      }
      setState(() {
        // refresh the countdown
        _remainingTime = Duration(seconds: remainingSeconds);
      });
    }

    _timer = Timer.periodic(Toast._defaultTimerStep, (_) => tick());
  }

  @override
  Widget build(BuildContext context) {
    final theme = context.toastTheme;
    final textScaler = MediaQuery.textScalerOf(context);
    return Container(
      width: textScaler.scale(theme.widgetWidth),
      height: textScaler.scale(theme.widgetHeight),
      decoration: BoxDecoration(
        borderRadius: theme.borderRadius,
        color: theme.backgroundColor,
        border: Border.all(width: theme.borderWidth, color: theme.borderColor),
      ),
      child: Container(
        padding: EdgeInsets.all(textScaler.scale(theme.spacing)),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
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
      style: theme.messageTextStyle,
      textAlign: TextAlign.center,
    );
  }

  Widget _buildCloseButton(ToastTheme theme) {
    return Padding(
      padding: theme.closeButtonPadding,
      child: GestureDetector(
        onTap: () {
          widget.onClose?.call();
        },
        child: DynamicThemeImage("toast_close_icon.svg"),
      ),
    );
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }
}
