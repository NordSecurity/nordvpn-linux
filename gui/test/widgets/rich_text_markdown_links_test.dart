import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/widgets/rich_text_markdown_links.dart';
import 'package:url_launcher_platform_interface/url_launcher_platform_interface.dart';

import '../utils/mock_url_launcher.dart';
import '../utils/test_helpers.dart';

void main() {
  testWidgets('RichTextMarkdownLinks invokes the callback of the tapped link', (
    tester,
  ) async {
    final urlLauncher = MockUrlLauncher();
    UrlLauncherPlatform.instance = urlLauncher;

    var termsTaps = 0;
    var policyTaps = 0;
    await tester.setupWidgetTest(
      RichTextMarkdownLinks(
        text:
            "Read the [terms](https://example.com/terms) and "
            "the [policy](https://example.com/policy).",
        onLinkTaps: [() => termsTaps++, () => policyTaps++],
      ),
    );

    await tester.tapOnText(find.textRange.ofSubstring('policy'));
    await tester.pumpAndSettle();

    expect(termsTaps, 0);
    expect(policyTaps, 1);
    expect(urlLauncher.launchedUrls, ['https://example.com/policy']);
  });
}
