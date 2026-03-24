
import 'package:flutter/material.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/theme/aurora_design.dart';

final class Toast extends StatelessWidget {
  const Toast({super.key});
  final double _width = 356.0;
  final double _height = 58.0;

  @override
  Widget build(BuildContext _) {
    return Container(
      width: _width,
      height: _height,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(12),
        color: AppDesign(ThemeMode.light).semanticColors.bgTertiary,
        border: Border.all(width:1, color: AppCoreColors().neutral300),
      ),
      child: Container(
        padding: const EdgeInsets.all(AppSpacing.spacing4),
        child: Row(
          children:[
        // general spacing for row -> 4 in all directions
        // pause icon
        // spacing 2 between
        // heading with spacing 0.5 from top
        // spacing 2 between text and button
        // close button
        // close button spacing 5 all directions
        _buildPauseIcon(),
        _buildWidgetText(),
        _buildCloseButton(),
          ]),
      
      ),
    );
  }

  Widget _buildPauseIcon() {
    return Container(
      width: 24,
      height: 24,
      padding: const EdgeInsets.only(right: AppSpacing.spacing2),
      child: DynamicThemeImage("toast_pause_icon.svg"),
    );
  }

  Widget _buildWidgetText() {
    return Container(
      width: 258,
      height: 22,
      padding: const EdgeInsets.only(top: AppSpacing.spacing0_5, right: AppSpacing.spacing2),
      child: Text(
        "VPN connection resumes in 4:59",
        style: AppDesign(ThemeMode.light).typography.subHeading,
        textAlign: TextAlign.center,
      ),
    );
  }

  Widget _buildCloseButton() {
        return Container(
      width: 26,
      height: 26,
      padding: const EdgeInsets.only(left: AppSpacing.spacing2),
      child: Container(
        padding: const EdgeInsets.all(5.0),
        child: DynamicThemeImage("toast_close_icon.svg"),
      )
    );
  }
}