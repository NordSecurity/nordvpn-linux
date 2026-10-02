import 'package:flutter_test/flutter_test.dart';

import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/markdown_text.dart';

Matcher isPlain(String text) =>
    isA<MarkdownPlainText>().having((p) => p.text, 'text', text);

Matcher isLink(String label, String url) => isA<MarkdownLink>()
    .having((l) => l.label, 'label', label)
    .having((l) => l.url, 'url', url);

void main() {
  group('MarkdownText.parse', () {
    test('keeps the raw text', () {
      const raw = 'See [Help](https://nordvpn.com/help) now';
      expect(MarkdownText.parse(raw).raw, raw);
    });

    test('empty text has no parts', () {
      final markdown = MarkdownText.parse('');
      expect(markdown.parts, isEmpty);
      expect(markdown.linksCount, 0);
    });

    test('text without links is a single plain part', () {
      final markdown = MarkdownText.parse('No links here.');
      expect(markdown.parts, [isPlain('No links here.')]);
      expect(markdown.linksCount, 0);
    });

    test('text that is only a link is a single link part', () {
      final markdown = MarkdownText.parse('[Help](https://nordvpn.com/help)');
      expect(markdown.parts, [isLink('Help', 'https://nordvpn.com/help')]);
      expect(markdown.linksCount, 1);
    });

    test('link in the middle splits the text around it', () {
      final markdown = MarkdownText.parse(
        'See [Help](https://nordvpn.com/help) for more.',
      );
      expect(markdown.parts, [
        isPlain('See '),
        isLink('Help', 'https://nordvpn.com/help'),
        isPlain(' for more.'),
      ]);
    });

    test('leading and trailing links produce no empty plain parts', () {
      final markdown = MarkdownText.parse(
        '[A](https://a.com) middle [B](https://b.com)',
      );
      expect(markdown.parts, [
        isLink('A', 'https://a.com'),
        isPlain(' middle '),
        isLink('B', 'https://b.com'),
      ]);
    });

    test('adjacent links produce no empty plain part between them', () {
      final markdown = MarkdownText.parse(
        '[A](https://a.com)[B](https://b.com)',
      );
      expect(markdown.parts, [
        isLink('A', 'https://a.com'),
        isLink('B', 'https://b.com'),
      ]);
      expect(markdown.linksCount, 2);
    });

    test('keeps the query string and path of the url', () {
      final markdown = MarkdownText.parse(
        '[Policy](https://my.nordaccount.com/legal/privacy-policy/?utm_source=app&nm=app)',
      );
      expect(markdown.parts, [
        isLink(
          'Policy',
          'https://my.nordaccount.com/legal/privacy-policy/?utm_source=app&nm=app',
        ),
      ]);
    });

    test('label may contain spaces and punctuation', () {
      final markdown = MarkdownText.parse(
        '[Contact support, 24/7!](https://support.nordvpn.com)',
      );
      expect(markdown.parts, [
        isLink('Contact support, 24/7!', 'https://support.nordvpn.com'),
      ]);
    });

    test('accepts http and upper case scheme', () {
      final markdown = MarkdownText.parse(
        '[A](http://a.com) [B](HTTPS://b.com)',
      );
      expect(markdown.parts, [
        isLink('A', 'http://a.com'),
        isPlain(' '),
        isLink('B', 'HTTPS://b.com'),
      ]);
    });

    test('url ends at the first closing parenthesis', () {
      final markdown = MarkdownText.parse('[A](https://a.com/x)y)');
      expect(markdown.parts, [isLink('A', 'https://a.com/x'), isPlain('y)')]);
    });

    for (final (description, text) in [
      ('non http scheme', '[Mail](mailto:support@nordvpn.com)'),
      ('relative url', '[Settings](/settings)'),
      ('url without scheme', '[Site](nordvpn.com)'),
      ('url containing whitespace', '[A](https://a.com/x y)'),
      ('empty label', '[](https://a.com)'),
      ('space between label and url', '[A] (https://a.com)'),
      ('label without url', '[A]'),
    ]) {
      test('$description is kept as plain text', () {
        final markdown = MarkdownText.parse(text);
        expect(markdown.parts, [isPlain(text)]);
        expect(markdown.linksCount, 0);
      });
    }
  });

  group('semanticsText', () {
    test('plain part announces its text', () {
      expect(const MarkdownPlainText('Hello').semanticsText, 'Hello');
    });

    test('link part announces only its label, without the url', () {
      const link = MarkdownLink(label: 'Help', url: 'https://nordvpn.com/help');
      expect(link.semanticsText, t.a11y.link(name: 'Help'));
      expect(link.semanticsText, isNot(contains('https://')));
    });

    test('text without links is announced unchanged', () {
      expect(
        MarkdownText.parse('No links here.').semanticsText,
        'No links here.',
      );
    });

    test('empty text is announced as empty', () {
      expect(MarkdownText.parse('').semanticsText, '');
    });

    test('links are replaced by their announcement in place', () {
      final markdown = MarkdownText.parse(
        'Read [Terms](https://a.com/terms) and [Privacy](https://a.com/privacy).',
      );
      expect(
        markdown.semanticsText,
        'Read ${t.a11y.link(name: 'Terms')} and '
        '${t.a11y.link(name: 'Privacy')}.',
      );
      expect(markdown.semanticsText, isNot(contains('https://')));
      expect(markdown.semanticsText, isNot(contains('](')));
    });
  });
}
