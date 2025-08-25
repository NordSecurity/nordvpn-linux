import 'package:flutter/material.dart';

// A ConstrainedBox that uses text scaler to calculate its size
class ScalerResponsiveBox extends StatelessWidget {
  final Widget child;
  final double maxWidth;
  final double? maxHeight;

  const ScalerResponsiveBox({
    super.key,
    required this.child,
    required this.maxWidth,
    this.maxHeight,
  });

  @override
  Widget build(BuildContext context) {
    return ConstrainedBox(
      constraints: BoxConstraints(
        maxWidth: dynamicScale(maxWidth, context),
        maxHeight: maxHeight != null
            ? dynamicScale(maxHeight!, context)
            : double.infinity,
      ),
      child: child,
    );
  }
}

// Calculates the size size using the text factor.
// When context is null it will use the scale factor from WidgetsBinding
double dynamicScale(double value, [BuildContext? context]) {
  if (context == null) {
    final textScaleFactor =
        WidgetsBinding.instance.platformDispatcher.textScaleFactor;
    return value * textScaleFactor;
  }

  final textScaler = MediaQuery.textScalerOf(context);
  return textScaler.scale(value);
}
