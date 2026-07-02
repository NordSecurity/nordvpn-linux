import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:nordvpn/i18n/strings.g.dart';
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
  FocusNode? _previousFocus;
  final FocusNode _closeButtonFocusNode = FocusNode(
    debugLabel: 'ToastCloseButton',
  );
  bool _isCloseButtonFocused = false;

  @override
  void initState() {
    super.initState();
    _remainingTime = widget.duration;
    _previousFocus = FocusManager.instance.primaryFocus;

    _closeButtonFocusNode.addListener(() {
      if (_isCloseButtonFocused != _closeButtonFocusNode.hasFocus) {
        setState(() => _isCloseButtonFocused = _closeButtonFocusNode.hasFocus);
      }
    });

    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      _closeButtonFocusNode.requestFocus();
    });

    void tick() {
      _remainingTime -= Toast._defaultTimerStep;
      final remainingSeconds = _remainingTime.inSeconds.clamp(
        0,
        widget.duration.inSeconds,
      );

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
        boxShadow: theme.shadow,
      ),
      child: MergeSemantics(
        child: Focus(
          onKeyEvent: _onKeyEvent,
          child: Padding(
            padding: EdgeInsets.all(textScaler.scale(theme.spacing)),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                ExcludeSemantics(child: _buildPauseIcon()),
                Expanded(
                  child: Semantics(
                    excludeSemantics: true,
                    label: _semanticsLabel(),
                    child: _buildWidgetText(theme),
                  ),
                ),
                _buildCloseButton(theme),
              ],
            ),
          ),
        ),
      ),
    );
  }

  KeyEventResult _onKeyEvent(FocusNode node, KeyEvent event) {
    if (event is! KeyDownEvent) return KeyEventResult.ignored;
    final key = event.logicalKey;

    if (key == LogicalKeyboardKey.enter ||
        key == LogicalKeyboardKey.numpadEnter ||
        key == LogicalKeyboardKey.space) {
      _restorePreviousFocus();
      widget.onClose?.call();
    }

    if (key == LogicalKeyboardKey.escape || key == LogicalKeyboardKey.tab) {
      _restorePreviousFocus();
    }

    return KeyEventResult.handled;
  }

  void _restorePreviousFocus() {
    final prev = _previousFocus;
    if (prev != null && prev.context != null) {
      prev.requestFocus();
    }
  }

  Widget _buildPauseIcon() {
    return DynamicThemeImage("toast_pause_icon.svg");
  }

  String _resumeMessage() {
    final m = _remainingTime.inMinutes.remainder(60).toString().padLeft(2, '0');
    final s = _remainingTime.inSeconds.remainder(60).toString().padLeft(2, '0');
    final h = _remainingTime.inHours.remainder(60);
    return h > 0
        ? t.ui.VPNResumesInWithHours(
            hours: h.toString().padLeft(2, '0'),
            minutes: m,
            seconds: s,
          )
        : t.ui.VPNResumesIn(minutes: m, seconds: s);
  }

  String _semanticsLabel() {
    final seconds = _remainingTime.inSeconds.remainder(60);
    final minutes = _remainingTime.inMinutes.remainder(60);
    final hours = _remainingTime.inHours.remainder(60);
    return hours > 0
        ? t.ui.VPNResumesInWithHours_a11y(
            hours: hours,
            minutes: minutes,
            seconds: seconds,
          )
        : t.ui.VPNResumesIn_a11y(minutes: minutes, seconds: seconds);
  }

  Widget _buildWidgetText(ToastTheme theme) {
    return Text(
      _resumeMessage(),
      style: theme.messageTextStyle,
      textAlign: TextAlign.center,
    );
  }

  Widget _buildCloseButton(ToastTheme theme) {
    return Padding(
      padding: theme.closeButtonPadding,
      child: Semantics(
        button: true,
        label: t.ui.close,
        child: Focus(
          focusNode: _closeButtonFocusNode,
          child: DecoratedBox(
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(4),
              border: Border.all(
                color: _isCloseButtonFocused
                    ? theme.focusBorderColor
                    : Colors.transparent,
                width: theme.borderWidth,
              ),
            ),
            child: GestureDetector(
              onTap: () => widget.onClose?.call(),
              child: SizedBox(
                height: 20,
                width: 20,
                child: DynamicThemeImage("toast_close_icon.svg"),
              ),
            ),
          ),
        ),
      ),
    );
  }

  @override
  void dispose() {
    _timer?.cancel();
    _closeButtonFocusNode.dispose();
    super.dispose();
  }
}
