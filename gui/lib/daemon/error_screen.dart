import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:grpc/grpc.dart';
import 'package:nordvpn/data/models/application_error.dart';
import 'package:nordvpn/data/providers/account_controller.dart';
import 'package:nordvpn/data/providers/consent_status_provider.dart';
import 'package:nordvpn/data/providers/grpc_connection_controller.dart';
import 'package:nordvpn/data/providers/login_status_provider.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/urls.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/pb/snapconf/snapconf.pb.dart';
import 'package:nordvpn/snap/snap_helpers.dart';
import 'package:nordvpn/snap/snap_screen.dart';
import 'package:nordvpn/widgets/copy_field.dart';
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
      return _displayError(_dataForGrpcError(ref, error));
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
    switch (error.errorType) {
      case ErrorType.snap:
        return SnapScreen(
          missingPermissions: error.snapInterfaces,
          retryCallback: error.retryCallback,
        );
      case ErrorType.generic:
        return FullScreenScaffold(child: FullScreenError(errorData: error));
    }
  }
}

ErrorData _dataForGrpcError(WidgetRef ref, Object error) {
  switch (error) {
    case ApplicationError error:
      return _dataForApplicationError(ref, error);
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

ErrorData _dataForApplicationError(WidgetRef ref, ApplicationError error) {
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
        recommendation: _buildCopyFieldForSocketNotFound(),
        footer: RichTextMarkdownLinks(
          text: t.ui.forTroubleshooting(supportUrl: supportCenterUrl),
        ),
      );

    case AppStatusCode.permissionsDenied:
      return ErrorData(
        title: t.ui.weCouldNotConnectToService,
        subtitle: t.ui.tryRunningTheseCommands,
        recommendation: const CopyField(
          items: [
            CopyItem(command: "sudo groupadd nordvpn"),
            CopyItem(command: "sudo usermod -aG nordvpn \$USER"),
          ],
        ),
      );

    case AppStatusCode.snapInterfaces:
      return _dataForSnapInterfaces(ref, error);

    case AppStatusCode.unknown:
      // for unknown issues display the generic screen
      break;
  }

  return _genericErrorMessage();
}

ErrorData _dataForSnapInterfaces(WidgetRef ref, ApplicationError error) {
  final missingPermissions = _extractMissingConnections(error.originalError);

  if (missingPermissions.isEmpty) {
    logger.w("_extractMissingConnections returned an empty list");
  }

  return ErrorData(
    title: t.ui.snapScreenTitle,
    errorType: ErrorType.snap,
    snapInterfaces: missingPermissions,
    retryCallback: () async {
      await ref.read(grpcConnectionControllerProvider.notifier).retry();
    },
  );
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

CopyField _buildCopyFieldForSocketNotFound() {
  if (SnapHelpers.isSnapContext()) {
    return CopyField(items: [CopyItem(command: "sudo snap start nordvpn")]);
  }

  return CopyField(
    items: [
      CopyItem(
        command: "sudo systemctl enable --now nordvpnd",
        description: t.ui.systemdDistribution,
      ),
      CopyItem(
        command: "/etc/init.d/nordvpn start",
        description: t.ui.nonSystemdDistro,
      ),
    ],
  );
}

List<String> _extractMissingConnections(Object? error) {
  if (error is! GrpcError) return const [];

  if (error.details != null) {
    for (var detail in error.details!) {
      if (detail is Any && detail.typeUrl.endsWith('ErrMissingConnections')) {
        try {
          return ErrMissingConnections.fromBuffer(
            detail.value,
          ).missingConnections;
        } catch (e) {
          logger.e('Failed to parse ErrMissingConnections: $e');
          return const [];
        }
      }
    }
  }

  return const [];
}
