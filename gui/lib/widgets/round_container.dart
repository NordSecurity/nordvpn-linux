import 'package:flutter/material.dart';
import 'package:nordvpn/theme/app_theme.dart';

// Custom container that has round corners and padding.
// This is the background container used for all the widgets
final class RoundContainer extends StatelessWidget {
  final Widget child;
  final double? minHeight;
  final double? minWidth;
  final double? maxWidth;
  final double? maxHeight;
  final double? width;
  final EdgeInsetsGeometry? padding;
  final EdgeInsetsGeometry? margin;
  final double? radius;
  final Color? color;

  const RoundContainer({
    super.key,
    required this.child,
    this.minHeight,
    this.maxWidth,
    this.maxHeight,
    this.minWidth,
    this.width,
    this.padding,
    this.margin,
    this.radius,
    this.color,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final appTheme = theme.extension<AppTheme>()!;

    return Container(
      width: width,
      constraints: _calculateConstraints(),
      padding:
          padding ??
          EdgeInsets.symmetric(
            vertical: appTheme.outerPadding,
            horizontal: appTheme.outerPadding,
          ),
      decoration: BoxDecoration(
        color: color ?? theme.colorScheme.surface,
        borderRadius: BorderRadius.circular(
          radius ?? appTheme.borderRadiusMedium,
        ),
      ),
      margin: margin ?? EdgeInsets.all(appTheme.margin),
      child: child,
    );
  }

  BoxConstraints? _calculateConstraints() {
    if (minHeight == null &&
        maxWidth == null &&
        maxHeight == null &&
        minWidth == null) {
      return null;
    }

    return BoxConstraints(
      minHeight: minHeight ?? 0,
      minWidth: minWidth ?? 0,
      maxHeight: maxHeight ?? double.infinity,
      maxWidth: maxWidth ?? double.infinity,
    );
  }
}
