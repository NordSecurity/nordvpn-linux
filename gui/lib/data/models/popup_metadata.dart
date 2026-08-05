import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

// Action executed after the user made a choice in the popup.
//
// It receives a container scoped [Ref] and not a [WidgetRef] on purpose: the
// popup is already closed when the action runs, so a `WidgetRef` would be
// disposed while the action is still awaiting the daemon. Actions are executed
// by `PopupActions`.
typedef PopupAction = FutureOr<void> Function(Ref ref);

// Base class for popups metadata, specifies `id`, optional `title`
// and popup `message`.
sealed class PopupMetadata {
  final int id;
  String? title;
  // Evaluated while the popup is built, so a [WidgetRef] is the correct scope
  // here. Unlike [PopupAction] it can never outlive the popup.
  final String Function(WidgetRef) message;

  PopupMetadata({required this.id, required this.message, this.title});

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is PopupMetadata &&
            runtimeType == other.runtimeType &&
            id == other.id);
  }

  @override
  int get hashCode => id.hashCode;
}

// Metadata for popups with yes/no decision. Besides the base of [PopupMetadata],
// it specifies also labels for "no" and "yes" buttons and actions executed
// after clicking on "yes" or "no" button. The popup is closed before the action
// is started and the progress is displayed by [PopupActionProgressOverlay].
final class DecisionPopupMetadata extends PopupMetadata {
  final String noButtonText;
  final String yesButtonText;
  final PopupAction yesAction;
  final PopupAction? noAction;
  // Displayed together with the progress indicator while the action runs.
  final String? progressMessage;

  DecisionPopupMetadata({
    required super.id,
    required super.message,
    required this.noButtonText,
    required this.yesButtonText,
    required this.yesAction,
    this.noAction,
    this.progressMessage,
    super.title,
  });
}

// Metadata for popups that can be only closed. Has just `id`, `title` nad `message`
// Optionally accepts [buttonText] to customize the close button label.
// If not provided, defaults to "Close".
final class InfoPopupMetadata extends PopupMetadata {
  final String? buttonText;
  InfoPopupMetadata({
    required super.id,
    required super.title,
    required super.message,
    this.buttonText,
  });
}

// Metadata for popups containing styled `header`, `image` and single action
// button. Also specifies text for the action button, action. Can be auto-closed
// after clicking the action button, otherwise it stays visible.
final class RichPopupMetadata extends PopupMetadata {
  final String header;
  final String actionButtonText;
  final PopupAction action;
  final Widget image;
  bool autoClose;
  // Displayed together with the progress indicator while the action runs.
  final String? progressMessage;

  RichPopupMetadata({
    required super.id,
    required super.message,
    required this.header,
    required this.actionButtonText,
    required this.action,
    required this.image,
    super.title,
    this.autoClose = true,
    this.progressMessage,
  });
}
