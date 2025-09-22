import 'dart:async';

import 'package:flutter/material.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/loading_button.dart';

// Class that defines a remove button with the bin.svg icon
final class BinButton extends StatelessWidget {
  final FutureOr<void> Function()? onPressed;

  const BinButton({super.key, this.onPressed});

  @override
  Widget build(BuildContext context) {
    return Tooltip(
      message: t.ui.delete,
      child: LoadingIconButton(
        onPressed: onPressed,
        child: DynamicThemeImage("bin.svg"),
      ),
    );
  }
}
