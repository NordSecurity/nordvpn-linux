import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

class AppRouteInfo {
  String label;
  String path;
  Icon icon;
  Icon selectedIcon;
  GoRouterWidgetBuilder builder;

  AppRouteInfo(
      {required this.label,
      required this.path,
      required this.icon,
      required this.selectedIcon,
      required this.builder});
}
