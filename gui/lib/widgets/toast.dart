import 'dart:async';
import 'package:flutter/material.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/theme/aurora_design.dart';

import 'package:nordvpn/internal/scaler_responsive_box.dart';

final class Toast extends StatefulWidget {
  const Toast({super.key, required this.duration});

  final Duration duration;

  @override
  State<Toast> createState() => _ToastState();
}

class _ToastState extends State<Toast> {
  final double _width = 356.0;
  final double _height = 58.0;

  late Duration _remainingTime;
  Timer? _timer;
  bool _isVisible = true;


 @override
  void initState() {
    super.initState();
    _remainingTime = widget.duration;

    void tick() {
      _remainingTime -= Duration(seconds: 1);
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
    _timer = Timer.periodic(const Duration(seconds: 1), (_) => tick());
  }

  @override
  Widget build(BuildContext context) {
    if (!_isVisible) return const SizedBox.shrink();

    return Container(
      width: dynamicScale(_width),
      height: dynamicScale(_height),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(12),
        color: AppDesign(ThemeMode.light).semanticColors.bgTertiary,
        border: Border.all(width: 1, color: AppCoreColors().neutral300),
      ),
      child: Container(
        padding: const EdgeInsets.all(AppSpacing.spacing4),
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
            _buildWidgetText(),
            _buildCloseButton(),
          ],
        ),
      ),
    );
  }

  Widget _buildPauseIcon() {
    return DynamicThemeImage("toast_pause_icon.svg");
  }

  Widget _buildWidgetText() {
    final m = _remainingTime.inMinutes.remainder(60).toString().padLeft(2, '0');
    final s = _remainingTime.inSeconds.remainder(60).toString().padLeft(2, '0');
    return Text(
      t.ui.VPNResumesIn(minutes: m, seconds: s),
      style: AppDesign(ThemeMode.light).typography.subHeading,
      textAlign: TextAlign.center,
    );
  }

  Widget _buildCloseButton() {
    return Padding(
      padding: const EdgeInsets.all(5.0),
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
}

