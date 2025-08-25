import 'package:flutter/material.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/widgets/full_screen_scaffold.dart';
import 'package:nordvpn/widgets/loading_indicator.dart';

final class LoadingScreen extends StatelessWidget {
  const LoadingScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return FullScreenScaffold(
      child: LoadingIndicator(message: t.ui.waitingToConnectToDaemon),
    );
  }
}
