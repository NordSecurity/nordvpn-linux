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
import 'package:nordvpn/widgets/link_types.dart';
import 'package:nordvpn/widgets/loading_button.dart';
import 'package:nordvpn/widgets/loading_indicator.dart';

final _dateFormat = DateFormat('d/M/y');

final class ProductItem {
  final String title;
  final String subtitle;
  final Uri uri;
  final String imageName;

  const ProductItem({
    required this.title,
    required this.subtitle,
    required this.uri,
    required this.imageName,
  });
}

final class AccountWidgetKeys {
  AccountWidgetKeys._();
  static const userInfo = Key("accountUserInfo");
  static const productsList = Key("accountProductsList");
  static const logoutButton = Key("logoutButton");
  static const subscriptionInfo = Key("subscriptionInfo");
  static const accountInfo = Key("accountInfo");
}

final class AccountDetailsSettings extends ConsumerWidget {
  const AccountDetailsSettings({super.key});

  static final _products = [
    ProductItem(
      title: t.ui.nordPass,
      subtitle: t.ui.nordPassDescription,
      uri: nordPassProductUrl,
      imageName: "nordpass.svg",
    ),
    ProductItem(
      title: t.ui.nordLocker,
      subtitle: t.ui.nordLockerDescription,
      uri: nordLockerProductUrl,
      imageName: "nordlocker.svg",
    ),
    ProductItem(
      title: t.ui.nordLayer,
      subtitle: t.ui.nordLayerDescription,
      uri: nordLayerProductUrl,
      imageName: "nordlayer.svg",
    ),
  ];

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
    final appTheme = context.appTheme;
    return SettingsWrapperWidget(
      itemsCount: 1,
      itemBuilder: (_, _) => Column(
        spacing: appTheme.verticalSpaceExtraLarge,
        children: [
          UserInfo(
            key: AccountWidgetKeys.userInfo,
            userAccount: userAccount,
            onLogout: () =>
                ref.read(accountControllerProvider.notifier).logout(),
          ),
          ProductsList(
            key: AccountWidgetKeys.productsList,
            products: _products,
          ),
        ],
      ),
      useSeparator: false,
    );
  }
}

final class UserInfoEntry extends StatelessWidget {
  final String title;
  final String description;
  final String linkText;
  final Uri link;
  const UserInfoEntry({
    super.key,
    required this.title,
    required this.description,
    required this.linkText,
    required this.link,
  });

  @override
  Widget build(BuildContext context) {
    final theme = context.appTheme;
    return Padding(
      padding: EdgeInsetsGeometry.fromLTRB(
        0,
        theme.verticalSpaceMedium,
        theme.verticalSpaceMedium,
        theme.verticalSpaceMedium,
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          Expanded(
            child: Column(
              spacing: 2,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(title, style: theme.bodyStrong),
                Text(description, style: theme.caption),
              ],
            ),
          ),
          FirstPartyLink(title: linkText, uri: link),
        ],
      ),
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
    return Column(
      spacing: 12,
      children: [
        Column(
          children: [
            UserInfoEntry(
              key: AccountWidgetKeys.subscriptionInfo,
              title: t.ui.subscription,
              description: _getSubscriptionExpirationDate(userAccount),
              linkText: t.ui.manageSubscription,
              link: manageSubscriptionUrl,
            ),
            UserInfoEntry(
              key: AccountWidgetKeys.accountInfo,
              title: _getAccountEmail(userAccount),
              description: _getAccountCreationDate(userAccount),
              linkText: t.ui.changePassword,
              link: changePasswordUrl,
            ),
          ],
        ),
        Row(children: [_buildLogoutButton()]),
      ],
    );
  }

  Widget _buildLogoutButton() {
    return LoadingOutlinedButton(
      key: AccountWidgetKeys.logoutButton,
      onPressed: onLogout,
      child: Text(t.ui.logout),
    );
  }

  String _getAccountEmail(UserAccount? account) {
    return account?.email ?? "";
  }

  String _getSubscriptionExpirationDate(UserAccount? account) {
    if (account == null || account.vpnExpirationDate == null) {
      return "";
    }

    if (account.isSubscriptionExpired) {
      return t.ui.subscriptionInactive;
    }

    final date = _dateFormat.format(account.vpnExpirationDate!);
    return t.ui.subscriptionValidationDate(expirationDate: date);
  }

  String _getAccountCreationDate(UserAccount? account) {
    if (account == null || account.createdOn == null) {
      return "";
    }

    final date = _dateFormat.format(account.createdOn!);
    return t.ui.accountCreatedOn(creation_date: date);
  }
}

final class ProductsList extends StatelessWidget {
  final List<ProductItem> products;

  const ProductsList({super.key, required this.products});

  @override
  Widget build(BuildContext context) {
    final appTheme = context.appTheme;
    return Column(
      children: [
        Row(children: [Text(t.ui.productHub, style: appTheme.caption)]),
        Column(
          children: products
              .map((product) => _buildProductHubItem(context, product))
              .toList(),
        ),
      ],
    );
  }

  Widget _buildProductHubItem(BuildContext context, ProductItem product) {
    final settingsTheme = context.settingsTheme;
    final appTheme = context.appTheme;
    const iconSize = 24.0;
    const iconSpacing = 4.0;
    const iconPadding = 4.0;
    const itemHorizontalPadding = 4.0;
    const itemVerticalPadding = 10.0;
    const itemTextHorizontalPadding = 4.0;
    const itemTextVerticalPadding = 2.0;
    const itemTextLinkSpacing = 6.0;

    return Padding(
      padding: EdgeInsets.symmetric(
        horizontal: itemHorizontalPadding,
        vertical: itemVerticalPadding,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        spacing: itemTextLinkSpacing,
        children: [
          // Row 1: Icon and Column of (title, subtitle)
          Row(
            spacing: iconSpacing,
            crossAxisAlignment: CrossAxisAlignment.center,
            children: [
              Padding(
                padding: EdgeInsetsGeometry.all(iconPadding),
                child: SizedBox(
                  width: iconSize,
                  height: iconSize,
                  child: DynamicThemeImage(product.imageName),
                ),
              ),
              Expanded(
                child: Padding(
                  padding: EdgeInsetsGeometry.symmetric(
                    horizontal: itemTextHorizontalPadding,
                    vertical: itemTextVerticalPadding,
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        product.title,
                        style: settingsTheme.otherProductsTitle,
                      ),
                      Text(
                        product.subtitle,
                        style: settingsTheme.otherProductsSubtitle,
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
          // Row 2: Empty area offset + learn more link
          Row(
            children: [
              SizedBox(
                width:
                    appTheme.trailingIconSize +
                    iconSpacing +
                    itemTextHorizontalPadding,
              ),
              FirstPartyLink(title: t.ui.learnMore, uri: product.uri),
            ],
          ),
        ],
      ),
    );
  }
}
