import 'package:flutter/material.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/input_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/loading_button.dart';

enum SubmitDisplay { always, never }

final class Input extends StatefulWidget {
  final Function(String)? onChanged;
  final bool Function(String) validateInput;
  // single error message that will always be displayed on invalid value
  final String? errorMessage;
  // the error message is generated dynamically based on the text field value
  final String Function(String)? onErrorMessage;
  final String? hintText;
  final String? text;
  final String? submitText;
  final SubmitDisplay submitDisplay;
  final Function(String)? onSubmitted;
  final TextEditingController? controller;

  const Input({
    super.key,
    this.onChanged,
    required this.validateInput,
    this.errorMessage,
    this.onErrorMessage,
    this.hintText,
    this.text,
    this.submitText,
    required this.submitDisplay,
    this.onSubmitted,
    this.controller,
  }) : assert(
         (onSubmitted != null) || (onChanged != null),
         "submit or change callback must not be null",
       ),
       assert(
         (errorMessage == null) || (onErrorMessage == null),
         "only static or dynamic error message can be specified",
       );

  @override
  InputState createState() => InputState();
}

final class InputState extends State<Input> {
  late final TextEditingController _controller;
  String _errorText = "";
  final _focus = FocusNode();
  // This is used to trigger submit action on the button and to display the
  // loading indicator instead of the button.
  // Not null when submit button is created.
  LoadingButtonController? _btnController;

  @override
  void initState() {
    super.initState();
    _focus.addListener(_onFocusChange);
    _controller = widget.controller ?? TextEditingController();
    _controller.addListener(_onContentChanged);
    if (widget.text != null) {
      _controller.text = widget.text!;
    }
  }

  @override
  void dispose() {
    super.dispose();
    _focus.removeListener(_onFocusChange);
    _focus.dispose();
    _btnController?.dispose();

    _controller.removeListener(_onContentChanged);
    if (widget.controller == null) {
      // dispose internal created controller
      _controller.dispose();
    }
  }

  @override
  Widget build(BuildContext context) {
    final appTheme = context.appTheme;
    final inputTheme = context.inputTheme;

    _updateError();

    return Row(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      spacing: appTheme.verticalSpaceSmall,
      children: [
        Expanded(
          child: TextField(
            style: inputTheme.textStyle,
            controller: _controller,
            focusNode: _focus,
            onSubmitted: (value) async {
              if (_btnController != null) {
                _btnController!.triggerTap();
              } else {
                await _submit(value);
              }
            },
            decoration: InputDecoration(
              helperText: "", // hack to make room for error message
              hintText: widget.hintText,
              constraints: BoxConstraints(minHeight: inputTheme.height),
              errorStyle: inputTheme.errorStyle,
              helperStyle: inputTheme.errorStyle,
              errorText: _errorText.isEmpty ? null : _errorText,
              enabledBorder: OutlineInputBorder(
                borderSide: BorderSide(
                  color: inputTheme.enabled.borderColor,
                  width: inputTheme.enabled.borderWidth,
                ),
              ),
              focusedBorder: OutlineInputBorder(
                borderSide: BorderSide(
                  color: inputTheme.focused.borderColor,
                  width: inputTheme.focused.borderWidth,
                ),
              ),
              errorBorder: OutlineInputBorder(
                borderSide: BorderSide(
                  color: inputTheme.error.borderColor,
                  width: inputTheme.error.borderWidth,
                ),
              ),
              focusedErrorBorder: OutlineInputBorder(
                borderSide: BorderSide(
                  color: inputTheme.focusedError.borderColor,
                  width: inputTheme.focusedError.borderWidth,
                ),
              ),
              suffixIcon: _controller.text.isNotEmpty
                  ? _buildClearIcon(inputTheme)
                  : const SizedBox.shrink(),
            ),
          ),
        ),
        if (_shouldShowSubmitButton()) _buildSubmitButton(),
      ],
    );
  }

  Widget _buildClearIcon(InputTheme inputTheme) {
    return IconButton(
      padding: EdgeInsets.zero,
      color: inputTheme.icon.color,
      hoverColor: inputTheme.icon.hoverColor,
      icon: DynamicThemeImage("close.svg"),
      onPressed: () {
        _focus.requestFocus();
        _controller.clear();
        setState(() => _errorText = "");
      },
    );
  }

  void _onFocusChange() {
    if (!_focus.hasFocus && (widget.text != null) && _controller.text.isEmpty) {
      _controller.text = widget.text!;
    }
  }

  bool _shouldShowSubmitButton() {
    if ((widget.submitText == null) || (widget.onSubmitted == null)) {
      return false;
    }

    switch (widget.submitDisplay) {
      case SubmitDisplay.never:
        return false;
      case SubmitDisplay.always:
        return true;
    }
  }

  Widget _buildSubmitButton() {
    assert((widget.submitText != null) && (widget.onSubmitted != null));
    if ((widget.submitText == null) || (widget.onSubmitted == null)) {
      logger.f("cannot create submit button");
      return const SizedBox.shrink();
    }

    _btnController ??= LoadingButtonController();
    return LoadingTextButton(
      key: UniqueKey(),
      controller: _btnController,
      onPressed:
          _errorText.isEmpty &&
              _controller.text.isNotEmpty &&
              (_controller.text != widget.text)
          ? () async => await _submit(_controller.text)
          : null,
      child: Text(widget.submitText!),
    );
  }

  void _onContentChanged() {
    setState(() => _updateError());
    if (widget.onChanged != null) {
      widget.onChanged!(_controller.text);
    }
  }

  void _updateError() {
    final value = _controller.text;
    if (value.isEmpty ||
        (value == widget.text) ||
        widget.validateInput(value)) {
      setState(() => _errorText = "");
    } else {
      _errorText = widget.errorMessage ?? widget.onErrorMessage!(value);
    }
  }

  Future<void> _submit(String value) async {
    if ((widget.onSubmitted != null) &&
        (_controller.text != widget.text) &&
        widget.validateInput(value)) {
      await widget.onSubmitted!(value);
    }
  }
}
