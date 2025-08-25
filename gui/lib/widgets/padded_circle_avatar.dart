import 'package:flutter/material.dart';

// This behaves like CircleAvatar, but it adds also a space between
// the border and the image
class PaddedCircleAvatar extends StatelessWidget {
  final double size;
  final Color borderColor;
  final double borderSize;
  final Widget child;

  const PaddedCircleAvatar({
    super.key,
    required this.size,
    required this.borderColor,
    required this.borderSize,
    required this.child,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        border: Border.all(color: borderColor, width: borderSize),
      ),
      child: Padding(padding: EdgeInsets.all(borderSize), child: child),
    );
  }
}
