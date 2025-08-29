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

    return DecoratedBox(
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
        borderRadius: BorderRadius.circular(dropdownTheme.borderRadius),
      ),
      child: Padding(
        padding: EdgeInsets.symmetric(
          horizontal: dropdownTheme.horizontalPadding,
        ),
        child: DropdownButton<T>(
          key: ValueKey(_value),
          focusNode: _focusNode,
          dropdownColor: dropdownTheme.color,
          focusColor: dropdownTheme.focusBorderColor,
          borderRadius: BorderRadius.circular(dropdownTheme.borderRadius),
          underline: SizedBox(),
          isExpanded: true,
          icon: DynamicThemeImage("dropdown_icon.svg"),
          value: _value,
          items: widget.items.map((e) {
            return DropdownMenuItem<T>(
              value: e.value,
              child: Text(e.label, style: appTheme.body, maxLines: 1),
            );
          }).toList(),
          onChanged: widget.enabled ? _onChanged : null,
        ),
      ),
    );
  }

  void _onChanged(T? value) {
    assert(value != null, "dropdown values cannot be null");
    if (value != null) {
      widget.onChanged(value);
      setState(() {
        _value = value;
      });
    }
  }
}

final class DropdownItem<T> {
  final T value;
  final String label;
  DropdownItem({required this.value, required this.label});
}
