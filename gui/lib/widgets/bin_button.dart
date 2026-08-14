import 'dart:async';

import 'package:flutter/material.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/loading_button.dart';

// Class that defines a remove button with the bin.svg icon
final class BinButton extends StatelessWidget {
  final FutureOr<void> Function()? onPressed;
  final String? semanticLabel;

  const BinButton({super.key, this.onPressed, this.semanticLabel});

  @override
  Widget build(BuildContext context) {
    final label = semanticLabel ?? t.ui.delete;
    return Tooltip(
      message: label,
      child: LoadingIconButton(
        onPressed: onPressed,
        child: Semantics(label: label, child: DynamicThemeImage("bin.svg")),
      ),
    );
  }
}
