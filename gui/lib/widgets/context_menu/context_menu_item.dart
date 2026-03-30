import 'package:flutter/material.dart';

/// A single item in a [ContextMenu].
final class ContextMenuItem {
  /// The label displayed in the menu row.
  final String label;

  /// Optional color for the label text. When null the theme default is used.
  final Color? labelColor;

  /// Called when the user taps this item. The menu closes automatically.
  final VoidCallback onTap;

  const ContextMenuItem({
    required this.label,
    this.labelColor,
    required this.onTap,
  });
}
