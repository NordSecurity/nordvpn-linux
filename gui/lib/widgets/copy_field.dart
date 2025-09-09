import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/copy_field_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';

final class CopyItem {
  final String command;
  final String? description;
  const CopyItem({required this.command, this.description});
}

final class CopyField extends StatelessWidget {
  final List<CopyItem> items;

  const CopyField({required this.items});

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

  Widget _buildCopyItem(BuildContext context, CopyItem item) {
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
      maxLines: null,

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