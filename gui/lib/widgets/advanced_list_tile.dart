import 'package:flutter/material.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/widgets/enabled_widget.dart';

enum TrailingLocation { center, top }

// Custom ListTile implementation that accepts trailing to be displayed on
// the same line with the title on centered.
// There are no restrictions for trailing widgets height.
final class AdvancedListTile extends StatelessWidget {
  final Widget? leading;
  final Widget? title;
  final Widget? subtitle;
  final Widget? center;
  final Widget? trailing;
  final VoidCallback? onTap;
  final TrailingLocation trailingLocation;
  final bool enabled;
  final EdgeInsetsGeometry? padding;
  final Color? color;

  const AdvancedListTile({
    super.key,
    this.leading,
    this.title,
    this.subtitle,
    this.center,
    this.trailing,
    this.onTap,
    this.trailingLocation = TrailingLocation.center,
    this.enabled = true,
    this.padding,
    this.color,
  });

  @override
  Widget build(BuildContext context) {
    final appTheme = context.appTheme;
    return EnabledWidget(
      enabled: enabled,
      disabledOpacity: appTheme.disabledOpacity,
      child: Material(
        type: MaterialType.transparency,
        child: Ink(
          color: enabled ? color : appTheme.area,
          child: InkWell(
            onTap: enabled ? onTap : null,
            child: Padding(
              padding:
                  padding ??
                  EdgeInsets.symmetric(
                    vertical: appTheme.outerPadding,
                    horizontal: appTheme.padding,
                  ),
              child: Row(
                children: [
                  if (leading != null)
                    Padding(
                      padding: EdgeInsets.only(right: appTheme.padding),
                      child: leading!,
                    ),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          mainAxisAlignment:
                              trailingLocation == TrailingLocation.center
                              ? MainAxisAlignment.start
                              : MainAxisAlignment.spaceBetween,
                          crossAxisAlignment: CrossAxisAlignment.center,
                          children: [
                            if (title != null) title!,
                            if (center != null) center!,
                            // show trailing when it needs to be on the same line
                            if (trailingLocation == TrailingLocation.top &&
                                trailing != null)
                              trailing!,
                          ],
                        ),
                        if (subtitle != null)
                          Padding(
                            padding: EdgeInsets.only(
                              top: appTheme.verticalSpaceSmall,
                            ),
                            child: subtitle!,
                          ),
                      ],
                    ),
                  ),
                  // show trailing when centered in the view
                  if (trailingLocation == TrailingLocation.center &&
                      trailing != null)
                    trailing!,
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
