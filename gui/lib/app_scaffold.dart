import 'package:flutter/material.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/internal/scaler_responsive_box.dart';

// The widget will be created for each route and will contain a navigation bar,
// app bar and display the screen specific screen
class AppScaffold extends StatelessWidget {
  final Widget child;

  const AppScaffold({super.key, required this.child});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: ScalerResponsiveBox(maxWidth: windowMaxSize.width, child: child),
    );
  }
}
