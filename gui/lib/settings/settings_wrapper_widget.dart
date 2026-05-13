import 'package:flutter/material.dart';
import 'package:nordvpn/settings/navigation.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/settings_theme.dart';
import 'package:nordvpn/widgets/advanced_list_tile.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/round_container.dart';

// Helper class to display the common part for each screen from settings
final class SettingsWrapperWidget extends StatelessWidget {
  final int itemsCount;
  final IndexedWidgetBuilder itemBuilder;
  final Widget? stickyHeader;
  final Widget? stickyFooter;
  final Widget? breadcrumbsSubtitle;
  final bool useSeparator;
  final double? spaceSize;

  const SettingsWrapperWidget({
    super.key,
    required this.itemsCount,
    required this.itemBuilder,
    this.stickyHeader,
    this.stickyFooter,
    this.useSeparator = true,
    this.breadcrumbsSubtitle,
    this.spaceSize,
  });

  @override
  Widget build(BuildContext context) {
    final appTheme = context.appTheme;

    return RoundContainer(
      padding: EdgeInsets.zero,
      child: Column(
        children: [
          Padding(
            padding: EdgeInsets.all(appTheme.outerPadding),
            child: Column(
              children: [
                NavBreadcrumbs(),
                if (breadcrumbsSubtitle != null) breadcrumbsSubtitle!,
              ],
            ),
          ),
          if (stickyHeader != null) stickyHeader!,
          if (stickyHeader != null) _divider(),
          Expanded(child: _pageContents(context)),
          if (stickyFooter != null) _divider(),
          if (stickyFooter != null) stickyFooter!,
        ],
      ),
    );
  }

  Widget _pageContents(BuildContext context) {
    if (useSeparator || (spaceSize != null)) {
      return ListView.separated(
        itemCount: itemsCount,
        padding: EdgeInsets.symmetric(
          horizontal: context.appTheme.outerPadding,
        ),
        itemBuilder: (context, index) => itemBuilder(context, index),
        separatorBuilder: (_, _) =>
            useSeparator ? _divider() : SizedBox(height: spaceSize),
      );
    } else {
      return ListView.builder(
        itemCount: itemsCount,
        padding: EdgeInsets.symmetric(
          horizontal: context.appTheme.outerPadding,
        ),
        itemBuilder: (context, index) => itemBuilder(context, index),
      );
    }
  }

  // helper function to create the items from the list
  static AdvancedListTile buildListItem(
    BuildContext context, {
    Key? key,
    String? iconName,
    required String title,
    TextStyle? titleStyle,
    String? subtitle,
    Widget? subtitleWidget,
    Widget? center,
    Widget? trailing,
    TrailingLocation trailingLocation = TrailingLocation.top,
    VoidCallback? onTap,
    bool enabled = true,
    EdgeInsetsGeometry? padding,
    Color? color,
  }) {
    final settingsTheme = context.settingsTheme;

    return AdvancedListTile(
      key: key ?? UniqueKey(),
      leading: iconName != null ? DynamicThemeImage(iconName) : null,
      title: Expanded(
        child: Text(title, style: titleStyle ?? settingsTheme.itemTitleStyle),
      ),
      subtitle: subtitle != null
          ? Text(subtitle, style: settingsTheme.itemSubtitleStyle)
          : subtitleWidget,
      center: center,
      trailing: trailing,
      onTap: onTap,
      trailingLocation: trailingLocation,
      enabled: enabled,
      padding: padding ?? settingsTheme.itemPadding,
      color: color,
    );
  }
}

Widget _divider() => Divider(height: 33);

final class SingleChildSettingsWrapperWidget extends StatelessWidget {
  final Widget? stickyHeader;
  final Widget child;
  final Widget? stickyFooter;
  final bool showDivider;

  const SingleChildSettingsWrapperWidget({
    super.key,
    this.stickyHeader,
    required this.child,
    this.stickyFooter,
    this.showDivider = true,
  });

  @override
  Widget build(BuildContext context) {
    final appTheme = context.appTheme;

    return RoundContainer(
      padding: EdgeInsets.zero,
      child: Column(
        children: [
          Padding(
            padding: EdgeInsets.all(appTheme.outerPadding),
            child: NavBreadcrumbs(),
          ),
          if (stickyHeader != null) stickyHeader!,
          if (showDivider && stickyHeader != null) _divider(),
          Expanded(child: child),
          if (showDivider && stickyFooter != null) _divider(),
          if (stickyFooter != null) stickyFooter!,
        ],
      ),
    );
  }
}
