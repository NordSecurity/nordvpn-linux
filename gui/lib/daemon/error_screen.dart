import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/application_error.dart';
import 'package:nordvpn/data/providers/account_controller.dart';
import 'package:nordvpn/data/providers/consent_status_provider.dart';
import 'package:nordvpn/data/providers/grpc_connection_controller.dart';
import 'package:nordvpn/data/providers/login_status_provider.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/urls.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/copy_field_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/full_screen_error.dart';
import 'package:nordvpn/widgets/full_screen_scaffold.dart';
import 'package:nordvpn/widgets/rich_text_markdown_links.dart';

final class ErrorScreen extends ConsumerWidget {
  const ErrorScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final connectionProvider = ref.watch(grpcConnectionControllerProvider);
    final consentStatus = ref.watch(consentStatusProvider);
    final loginStatus = ref.watch(loginStatusProvider);
    final account = ref.watch(accountControllerProvider);

    if (connectionProvider case AsyncError(:final error)) {
      logger.i("connection error $error");
      return _displayError(_dataForGrpcError(error));
    }

    if (consentStatus case AsyncError(:final error)) {
      logger.i("consent error $error");
      return _displayError(_dataForConsentError(ref));
    }

    if (loginStatus case AsyncError(:final error)) {
      logger.i("login status $error");
      return _displayError(_dataForLoginError(ref));
    }

    if (account case AsyncError(:final error)) {
      logger.i("display account info error $error");
      return _displayError(_dataForAccountError(ref));
    }

    logger.e("error screen displayed");
    return _displayError(_genericErrorMessage());
  }

  Widget _displayError(ErrorData error) {
    return FullScreenScaffold(child: FullScreenError(errorData: error));
  }
}

final class _CopyItem {
  final String command;
  final String? description;
  const _CopyItem({required this.command, this.description});
}

final class _CopyField extends StatelessWidget {
  final List<_CopyItem> items;

  const _CopyField({required this.items});

  @override
  Widget build(BuildContext context) {
    final appTheme = context.appTheme;

    return LayoutBuilder(
      builder: (context, constraints) {
        return Padding(
          padding: EdgeInsets.only(bottom: appTheme.verticalSpaceLarge),
          child: FractionallySizedBox(
            widthFactor: 0.5,
            child: Column(
              spacing: appTheme.verticalSpaceMedium,
              children: [
                for (final item in items) _buildCopyItem(context, item),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildCopyItem(BuildContext context, _CopyItem item) {
    final appTheme = context.appTheme;
    final copyFieldTheme = context.copyFieldTheme;

    return Column(
      spacing: appTheme.verticalSpaceSmall,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (item.description != null)
          Text(item.description!, style: copyFieldTheme.descriptionTextStyle),
        Container(
          decoration: BoxDecoration(
            color: appTheme.areaBackgroundColor,
            borderRadius: BorderRadius.circular(copyFieldTheme.borderRadius),
          ),
          child: Row(
            children: [
              Expanded(child: _textArea(copyFieldTheme, item.command)),
              _copyButton(item.command),
            ],
          ),
        ),
      ],
    );
  }

  Widget _textArea(CopyFieldTheme copyFieldTheme, String text) {
    return TextField(
      controller: TextEditingController(text: text),
      readOnly: true,
      decoration: const InputDecoration(enabledBorder: InputBorder.none),
      style: copyFieldTheme.commandTextStyle,
    );
  }

  IconButton _copyButton(String text) {
    return IconButton(
      tooltip: t.ui.copy,
      onPressed: () {
        Clipboard.setData(ClipboardData(text: text));
      },
      icon: DynamicThemeImage("copy_icon.svg"),
    );
  }
}

ErrorData _dataForGrpcError(Object error) {
  switch (error) {
    case ApplicationError error:
      return _dataForApplicationError(error);
  }

  return _genericErrorMessage();
}

ErrorData _genericErrorMessage() {
  return ErrorData(
    title: t.ui.failedToLoadService,
    footer: RichTextMarkdownLinks(
      text: t.ui.needHelp(supportUrl: supportCenterUrl),
    ),
  );
}

ErrorData _dataForApplicationError(ApplicationError error) {
  switch (error.code) {
    case AppStatusCode.compatibilityIssue:
      return ErrorData(
        title: t.ui.appVersionIsIncompatible,
        subtitle: t.ui.appVersionIsIncompatibleDescription,
        recommendation: RichTextMarkdownLinks(
          text: t.ui.appVersionCompatibilityRecommendation(
            compatibilityUrl: versionCompatibilityInfoUrl,
          ),
        ),
      );

    case AppStatusCode.socketNotFound:
      return ErrorData(
        title: t.ui.weCouldNotConnectToService,
        subtitle: t.ui.tryRunningOneCommand,
        recommendation: _CopyField(
          items: [
            _CopyItem(
              command: "sudo systemctl enable --now nordvpnd",
              description: t.ui.systemdDistribution,
            ),
            _CopyItem(
              command: "/etc/init.d/nordvpn start",
              description: t.ui.nonSystemdDistro,
            ),
          ],
        ),
        footer: RichTextMarkdownLinks(
          text: t.ui.forTroubleshooting(supportUrl: supportCenterUrl),
        ),
      );

    case AppStatusCode.permissionsDenied:
      return ErrorData(
        title: t.ui.weCouldNotConnectToService,
        subtitle: t.ui.tryRunningTheseCommands,
        recommendation: const _CopyField(
          items: [
            _CopyItem(command: "sudo groupadd nordvpn"),
            _CopyItem(command: "sudo usermod -aG nordvpn \$USER"),
          ],
        ),
      );

    case AppStatusCode.unknown:
      // for unknown issues display the generic screen
      break;
  }

  return _genericErrorMessage();
}

ErrorData _dataForConsentError(WidgetRef ref) {
  return ErrorData(
    title: t.ui.weHitAnError,
    subtitle: t.ui.failedToFetchConsentData,
    retryCallback: () async {
      await ref.read(consentStatusProvider.notifier).retry();
    },
    footer: RichTextMarkdownLinks(
      text: t.ui.issuePersists(supportUrl: supportCenterUrl),
    ),
  );
}

ErrorData _dataForLoginError(WidgetRef ref) {
  return ErrorData(
    title: t.ui.weHitAnError,
    subtitle: t.ui.failedToFetchAccountData,
    retryCallback: () async {
      await ref.read(loginStatusProvider.notifier).retry();
    },
    footer: RichTextMarkdownLinks(
      text: t.ui.issuePersists(supportUrl: supportCenterUrl),
    ),
  );
}

ErrorData _dataForAccountError(WidgetRef ref) {
  return ErrorData(
    title: t.ui.weHitAnError,
    subtitle: t.ui.failedToFetchAccountData,
    retryCallback: () async {
      // add some delay to show the loading indicator for the retry button
      await Future.delayed(Duration(milliseconds: 100));
      // ignore: unused_result
      await ref.refresh(accountControllerProvider.future);
    },
    footer: RichTextMarkdownLinks(
      text: t.ui.issuePersists(supportUrl: supportCenterUrl),
    ),
  );
}
