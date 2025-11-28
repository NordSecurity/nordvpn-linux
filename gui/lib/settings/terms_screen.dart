import 'package:flutter/material.dart';
import 'package:nordvpn/internal/urls.dart';
import 'package:nordvpn/settings/settings_wrapper_widget.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/widgets/link_types.dart';

final class LegalInformationKeys {
  LegalInformationKeys._();
  static const descriptionKey = Key("legalDescription");
  static const termsOfServiceLinkKey = Key("legalTermsOfServiceLink");
  static const autoRenewalTermsLinkKey = Key("legalAutoRenewalTermsKeyLink");
  static const privacyPolicyLinkKey = Key("legalPrivacyPolicyLink");
}

class LegalInformation extends StatelessWidget {
  const LegalInformation({super.key});

  @override
  Widget build(BuildContext context) {
    return SettingsWrapperWidget(
      itemsCount: 1,
      itemBuilder: (context, _) => _build(context),
    );
  }

  Widget _build(BuildContext context) {
    final appTheme = context.appTheme;
    final links = [
      (
        t.ui.termsOfService,
        termsOfServiceUrl,
        LegalInformationKeys.termsOfServiceLinkKey,
      ),
      (
        t.ui.autoRenewalTerms,
        autoRenewalTermsUrl,
        LegalInformationKeys.autoRenewalTermsLinkKey,
      ),
      (
        t.ui.privacyPolicy,
        privacyPolicyUrl,
        LegalInformationKeys.privacyPolicyLinkKey,
      ),
    ];

    return Column(
      spacing: appTheme.verticalSpaceMedium,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          key: LegalInformationKeys.descriptionKey,
          t.ui.termsAgreementDescription,
          style: appTheme.body,
        ),
        ListView.separated(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          itemCount: links.length,
          separatorBuilder: (context, index) => _buildDivider(context),
          itemBuilder: (context, index) {
            final (title, uri, key) = links[index];
            return _buildLinkEntry(context, title, uri, key);
          },
        ),
      ],
    );
  }

  Widget _buildLinkEntry(
    BuildContext context,
    String title,
    Uri link,
    Key key,
  ) {
    final appTheme = context.appTheme;
    return Row(
      children: [
        Expanded(child: Text(title, style: appTheme.body)),
        FirstPartyLink(key: key, title: t.ui.readMore, uri: link),
      ],
    );
  }

  Widget _buildDivider(BuildContext context) {
    final appTheme = context.appTheme;
    return Divider(height: 13, thickness: 1, color: appTheme.dividerColor);
  }
}
