import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/providers/snap_permissions_provider.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/scaler_responsive_box.dart';
import 'package:nordvpn/internal/urls.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/error_screen_theme.dart';
import 'package:nordvpn/widgets/copy_field.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/full_screen_scaffold.dart';
import 'package:nordvpn/widgets/loading_button.dart';
import 'package:nordvpn/widgets/rich_text_markdown_links.dart';

final class SnapScreen extends ConsumerWidget {
  const SnapScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final appTheme = context.appTheme;
    final errorScreenTheme = context.errorScreenTheme;
    final missingPermissions = ref.watch(snapPermissionsProvider).valueOrNull;

    return FullScreenScaffold(
      child: Padding(
        padding: EdgeInsets.all(appTheme.padding),
        child: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Expanded(child: SizedBox.shrink()),
              DynamicThemeImage("something_went_wrong.svg"),
              Text(
                t.ui.snapScreenTitle,
                style: errorScreenTheme.titleTextStyle,
              ),
              Padding(
                padding: EdgeInsets.only(top: appTheme.verticalSpaceSmall),
                child: ScalerResponsiveBox(
                  maxWidth: 262,
                  child: Text(
                    t.ui.snapScreenDescription,
                    textAlign: TextAlign.center,
                    softWrap: true,
                    style: errorScreenTheme.descriptionTextStyle,
                  ),
                ),
              ),
              Padding(
                padding: EdgeInsets.only(top: appTheme.verticalSpaceMedium),
                child: CopyField(
                  items: [_buildNeededCommands(missingPermissions)],
                ),
              ),
              Padding(
                padding: EdgeInsets.only(top: appTheme.verticalSpaceMedium),
                child: LoadingElevatedButton(
                  onPressed: () async {
                    await ref.read(snapPermissionsProvider.notifier).retry();
                  },
                  child: Text(t.ui.refresh),
                ),
              ),
              Padding(
                padding: EdgeInsets.only(top: appTheme.verticalSpaceMedium),
                child: RichTextMarkdownLinks(
                  text: t.ui.needHelp(supportUrl: supportCenterUrl),
                ),
              ),
              const Expanded(child: SizedBox.shrink()),
            ],
          ),
        ),
      ),
    );
  }

  CopyItem _buildNeededCommands(List<String>? missingPermissions) {
    final commands =
        missingPermissions
            ?.map((item) => "sudo snap connect nordvpn:$item")
            .join("\n") ??
        "";
    return CopyItem(command: commands);
  }
}
