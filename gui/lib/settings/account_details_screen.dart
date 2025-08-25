import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:nordvpn/data/models/user_account.dart';
import 'package:nordvpn/data/providers/account_controller.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/urls.dart';
import 'package:nordvpn/settings/settings_wrapper_widget.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/settings_theme.dart';
import 'package:nordvpn/widgets/custom_error_widget.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/link.dart';
import 'package:nordvpn/widgets/loading_button.dart';
import 'package:nordvpn/widgets/loading_indicator.dart';

enum Products { nordPass, nordLocker, nordLayer }

final _dateFormat = DateFormat('d MMM y');
const products = [Products.nordPass, Products.nordLocker, Products.nordLayer];

final class AccountDetailsSettings extends ConsumerWidget {
  const AccountDetailsSettings({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final accountController = ref.watch(accountControllerProvider);
    return accountController.when(
      loading: () => const LoadingIndicator(),
      error: (error, stackTrace) => CustomErrorWidget(message: "$error"),
      data: (userAccount) => _build(context, ref, userAccount),
    );
  }

  Widget _build(BuildContext context, WidgetRef ref, UserAccount? userAccount) {
    return SettingsWrapperWidget(
      itemsCount: 1,
      stickyHeader: UserInfo(
        userAccount: userAccount,
        onLogout: () => ref.read(accountControllerProvider.notifier).logout(),
      ),
      itemBuilder: (context, _) => ProductsList(products: products),
      stickyFooter: FooterLinks(),
    );
  }
}

final class UserInfo extends StatelessWidget {
  final UserAccount? userAccount;
  final VoidCallback onLogout;

  const UserInfo({
    super.key,
    required this.userAccount,
    required this.onLogout,
  });

  @override
  Widget build(BuildContext context) {
    final theme = context.appTheme;
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Padding(
          padding: const EdgeInsets.only(left: 30),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [_accountEmail(theme), _accountExpirationDate(theme)],
          ),
        ),
        Padding(
          padding: const EdgeInsets.only(right: 16),
          child: _logoutButton(),
        ),
      ],
    );
  }

  Text _accountEmail(AppTheme theme) {
    return Text(userAccount?.email ?? "", style: theme.bodyStrong);
  }

  Text _accountExpirationDate(AppTheme theme) =>
      Text(_expirationDateFrom(userAccount), style: theme.caption);

  Widget _logoutButton() {
    return LoadingOutlinedButton(onPressed: onLogout, child: Text(t.ui.logout));
  }
}

String _expirationDateFrom(UserAccount? user) {
  if ((user == null) || user.vpnExpirationDate == null) {
    return "";
  }
  final date = _dateFormat.format(user.vpnExpirationDate!);
  return t.ui.subscriptionValidationDate(expirationDate: date);
}

final class ProductsList extends StatelessWidget {
  final List<Products> products;

  const ProductsList({super.key, required this.products});

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        for (final item in products)
          switch (item) {
            Products.nordPass => _buildItemForProduct(
              context: context,
              title: t.ui.nordPass,
              subtitle: t.ui.nordPassDescription,
              uri: nordPassProductUrl,
              imageName: "nordpass.svg",
            ),
            Products.nordLocker => _buildItemForProduct(
              context: context,
              title: t.ui.nordLocker,
              subtitle: t.ui.nordLockerDescription,
              uri: nordLockerProductUrl,
              imageName: "nordlocker.svg",
            ),
            Products.nordLayer => _buildItemForProduct(
              context: context,
              title: t.ui.nordLayer,
              subtitle: t.ui.nordLayerDescription,
              uri: nordLayerProductUrl,
              imageName: "nordlayer.svg",
            ),
          },
      ],
    );
  }

  Widget _buildItemForProduct({
    required BuildContext context,
    required String title,
    required String subtitle,
    required Uri uri,
    required String imageName,
  }) {
    final settingsTheme = context.settingsTheme;
    final appTheme = context.appTheme;

    return ListTile(
      leading: SizedBox(
        width: appTheme.trailingIconSize,
        height: appTheme.trailingIconSize,
        child: DynamicThemeImage(imageName),
      ),
      title: Text(title, style: settingsTheme.otherProductsTitle),
      subtitle: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(subtitle, style: settingsTheme.otherProductsSubtitle),
          Link(title: t.ui.learnMore, uri: uri),
        ],
      ),
    );
  }
}

final class FooterLinks extends StatelessWidget {
  const FooterLinks({super.key});

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceEvenly,
      children: [
        Link(
          title: t.ui.termsOfService,
          uri: termsOfServiceUrl,
          size: LinkSize.small,
        ),
        Link(
          title: t.ui.privacyPolicy,
          uri: privacyPolicyUrl,
          size: LinkSize.small,
        ),
        Link(
          title: t.ui.subscriptionInfo,
          uri: subscriptionInfoUrl,
          size: LinkSize.small,
        ),
      ],
    );
  }
}
