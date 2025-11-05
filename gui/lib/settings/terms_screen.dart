import 'package:flutter/material.dart';
import 'package:nordvpn/internal/urls.dart';
import 'package:nordvpn/settings/settings_wrapper_widget.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/widgets/link.dart';

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
    final links = [
      (t.ui.termsOfService, termsOfServiceUrl),
      (t.ui.autoRenewalTerms, autoRenewalTermsUrl),
      (t.ui.privacyPolicy, privacyPolicyUrl),
    ];

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(t.ui.termsAgreementDescription, style: context.body),
        SizedBox(height: 16),
        ListView.separated(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          itemCount: links.length,
          separatorBuilder: (context, index) => _buildDivider(context),
          itemBuilder: (context, index) {
            final (title, uri) = links[index];
            return _buildLinkEntry(context, title, uri);
          },
        ),
      ],
    );
  }

  Widget _buildLinkEntry(BuildContext context, String title, Uri link) {
    return Row(
      children: [
        Expanded(child: Text(title, style: context.body)),
        IconLink(
          title: "Read more",
          uri: link,
          iconPath: "external_link.svg",
          size: LinkSize.normal,
        ),
      ],
    );
  }

  Widget _buildDivider(BuildContext context) {
    return Divider(height: 13, thickness: 1, color: context.dividerColor);
  }
}
