import 'package:flutter/material.dart';
import 'package:nordvpn/widgets/round_container.dart';

class FullScreenScaffold extends StatelessWidget {
  final Widget child;

  const FullScreenScaffold({super.key, required this.child});

  @override
  Widget build(BuildContext context) {
    return Center(child: RoundContainer(child: child));
  }
}
