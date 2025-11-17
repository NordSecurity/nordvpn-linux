import 'package:nordvpn/internal/uri_launch_extension.dart';

final supportCenterUrl = Uri.parse(
  "https://support.nordvpn.com/?utm_medium=app&utm_source=nordvpn-linux-gui&utm_campaign=settings-get_help&nm=app&ns=nordvpn-linux-gui&nc=settings-get_help",
);

final versionCompatibilityInfoUrl = Uri.parse(
  "https://nordvpn.com/download/linux/",
);

final whatIsNordAccountUrl = Uri.parse(
  "https://nordvpn.com/blog/introducing-nord-account/",
);

final renewSubscriptionUrl = Uri.parse(
  "https://my.nordaccount.com/plans/?product_group=nordvpn&login_target=nordvpn&utm_source=linux&utm_medium=app&utm_campaign=desktop-app&redirect_uri=nordvpn://claim-online-purchase",
);

final nordPassProductUrl = Uri.parse(
  "https://nordpass.com/?utm_medium=app&utm_source=nordvpn-linux-gui&utm_campaign=settings_apps-explore_nordpass&nm=app&ns=nordvpn-linux-gui&nc=settings-explore_nordpass",
);

final nordLockerProductUrl = Uri.parse(
  "https://nordlocker.com/?utm_medium=app&utm_source=nordvpn-linux-gui&utm_campaign=settings_apps-explore_nordlocker&nm=app&ns=nordvpn-linux-gui&nc=settings-explore_nordlocker",
);

final nordLayerProductUrl = Uri.parse(
  "https://nordlayer.com/?utm_medium=app&utm_source=nordvpn-linux-gui&utm_campaign=settings_apps-explore_nordlayer&nm=app&ns=nordvpn-linux-gui&nc=settings-explore_nordlayer",
);

final subscriptionInfoUrl = UriWithToken.parse(
  "https://my.nordaccount.com/billing/my-subscriptions/?utm_medium=app&utm_source=nordvpn-linux-gui&utm_campaign=home-my_subscriptions&nm=app&ns=nordvpn-linux-gui&nc=home-my_subscriptions",
);

final getDedicatedIpUrl = UriWithToken.parse(
  "https://my.nordaccount.com/plans/dedicated-ip/?utm_medium=app&utm_source=nordvpn-linux-gui&utm_campaign=dedicatedip-choose_plan&nm=app&ns=nordvpn-linux-gui&nc=dedicatedip-choose_plan",
);

final chooseDedicatedIpUrl = UriWithToken.parse(
  "https://my.nordaccount.com/plans/dedicated-ip/?utm_medium=app&utm_source=nordvpn-linux-gui&utm_campaign=dedicatedip-choose_plan&nm=app&ns=nordvpn-linux-gui&nc=dedicatedip-choose_plan",
);

final createAccountUrl = Uri.parse(
  "https://nordaccount.com/signup?utm_medium=app&utm_source=nordvpn-linux-gui&utm_campaign=home-sign_up&nm=app&ns=nordvpn-linux-gui&nc=home-sign_up",
);

final loginUrl = UriWithToken.parse(
  "https://nordaccount.com/login?utm_medium=app&utm_source=nordvpn-linux-gui&utm_campaign=home-log_in&nm=app&ns=nordvpn-linux-gui&nc=home-log_in",
);

final countriesApiUrl = Uri.parse(
  "https://api.nordvpn.com/v1/servers/countries",
);

final autoRenewalTermsUrl = Uri.parse(
  "https://my.nordaccount.com/legal/terms-of-service/subscription/?utm_medium=app&utm_source=nordvpn-linux-gui&utm_campaign=settings_terms-auto_renewal_terms&nm=app&ns=nordvpn-linux-gui&nc=settings-auto_renewal_terms",
);

final termsOfServiceUrl = Uri.parse(
  "https://my.nordaccount.com/legal/terms-of-service/?utm_medium=app&utm_source=nordvpn-linux-gui&utm_campaign=settings_terms-terms_of_service&nm=app&ns=nordvpn-linux-gui&nc=settings-terms_of_service",
);

final privacyPolicyUrl = Uri.parse(
  "https://my.nordaccount.com/legal/privacy-policy/?utm_medium=app&utm_source=nordvpn-linux-gui&utm_campaign=settings_terms-privacy_policy&nm=app&ns=nordvpn-linux-gui&nc=settings-privacy_policy",
);

final manageSubscriptionUrl = Uri.parse(
  "https://my.nordaccount.com/billing/my-subscriptions/?utm_medium=app&utm_source=nordvpn-linux-gui&utm_campaign=settings_account-manage_subscription&nm=app&ns=nordvpn-linux-gui&nc=settings-manage_subscription",
);

final changePasswordUrl = Uri.parse(
  "https://my.nordaccount.com/account-settings/account-management/?utm_medium=app&utm_source=nordvpn-linux-gui&utm_campaign=settings_account-change_password&nm=app&ns=nordvpn-linux-gui&nc=settings-change_password",
);
