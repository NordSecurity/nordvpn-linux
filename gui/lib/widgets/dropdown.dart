import 'package:flutter/material.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/dropdown_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';

final class Dropdown<T> extends StatefulWidget {
  final T initialValue;
  final List<DropdownItem<T>> items;
  final void Function(T value) onChanged;
  final bool enabled;
  final bool showError;

  const Dropdown({
    super.key,
    required this.initialValue,
    required this.items,
    required this.onChanged,
    this.enabled = true,
    this.showError = false,
  });

  @override
  State<Dropdown> createState() => _DropdownState<T>();
}

class _DropdownState<T> extends State<Dropdown<T>> {
  // current selected value
  late T _value;
  // FocusNode to track focus state for dropdown
  final FocusNode _focusNode = FocusNode();
  final MenuController _menuController = MenuController();
  bool _isFocused = false;

  @override
  void initState() {
    super.initState();

    _value = widget.initialValue;
    _focusNode.addListener(() {
      if (_isFocused != _focusNode.hasFocus) {
        setState(() {
          _isFocused = _focusNode.hasFocus;
        });
      }
    });
  }

  @override
  void dispose() {
    _focusNode.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final dropdownTheme = context.dropdownTheme;
    final appTheme = context.appTheme;

    final borderRadius = BorderRadius.circular(dropdownTheme.borderRadius);
    final padding = EdgeInsets.symmetric(
      horizontal: dropdownTheme.horizontalPadding,
    );
    final selected = widget.items.where((e) => e.value == _value).firstOrNull;

    return LayoutBuilder(
      builder: (context, constraints) => MenuAnchor(
        controller: _menuController,
        childFocusNode: _focusNode,
        onOpen: () => setState(() {}),
        onClose: () => setState(() {}),
        style: MenuStyle(
          backgroundColor: WidgetStatePropertyAll(dropdownTheme.color),
          shape: WidgetStatePropertyAll(
            RoundedRectangleBorder(borderRadius: borderRadius),
          ),
          minimumSize: WidgetStatePropertyAll(
            Size(constraints.hasBoundedWidth ? constraints.maxWidth : 0, 0),
          ),
        ),
        menuChildren: widget.items.map((e) {
          return MenuItemButton(
            autofocus: e.value == _value,
            onPressed: () => _onChanged(e.value),
            style: MenuItemButton.styleFrom(padding: padding),
            child: Text(e.label, style: appTheme.body, maxLines: 1),
          );
        }).toList(),
        child: DecoratedBox(
          decoration: BoxDecoration(
            color: dropdownTheme.color,
            border: Border.all(
              color: widget.showError
                  ? dropdownTheme.errorBorderColor
                  : _isFocused
                  ? dropdownTheme.focusBorderColor
                  : dropdownTheme.borderColor,
              width: dropdownTheme.borderWidth,
            ),
            borderRadius: borderRadius,
          ),
          child: Semantics(
            button: true,
            expanded: _menuController.isOpen,
            child: InkWell(
              focusNode: _focusNode,
              onTap: widget.enabled ? _toggleMenu : null,
              focusColor: dropdownTheme.focusBorderColor,
              borderRadius: borderRadius,
              child: ConstrainedBox(
                constraints: const BoxConstraints(
                  minHeight: kMinInteractiveDimension,
                ),
                child: Padding(
                  padding: padding,
                  child: Row(
                    children: [
                      Expanded(
                        child: Text(
                          selected?.label ?? "",
                          style: appTheme.body,
                          maxLines: 1,
                        ),
                      ),
                      DynamicThemeImage("dropdown_icon.svg"),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  void _toggleMenu() {
    if (_menuController.isOpen) {
      _menuController.close();
    } else {
      _menuController.open();
    }
  }

  void _onChanged(T value) {
    widget.onChanged(value);
    setState(() => _value = value);
  }
}

final class DropdownItem<T> {
  final T value;
  final String label;
  DropdownItem({required this.value, required this.label});
}
