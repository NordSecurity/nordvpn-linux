import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

// Base class for popups metadata, specifies `id`, optional `title`
// and popup `message`.
sealed class PopupMetadata {
  final int id;
  String? title;
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
// it specifies also labels for "no" and "yes" buttons and acton executed
// after clicking on "yes" button.
final class DecisionPopupMetadata extends PopupMetadata {
  final String noButtonText;
  final String yesButtonText;
  final Function(WidgetRef ref) yesAction;

  DecisionPopupMetadata({
    required super.id,
    required super.message,
    required this.noButtonText,
    required this.yesButtonText,
    required this.yesAction,
    super.title,
  });
}

// Metadata for popups that can be only closed. Has just `id`, `title` nad `message`
final class InfoPopupMetadata extends PopupMetadata {
  InfoPopupMetadata({
    required super.id,
    required super.title,
    required super.message,
  });
}

// Metadata for popups containing styled `header`, `image` and single action
// button. Also specifies text for the action button, action. Can be auto-closed
// after clicking the action button, otherwise it stays visible.
final class RichPopupMetadata extends PopupMetadata {
  final String header;
  final String actionButtonText;
  final Function(WidgetRef ref) action;
  final Widget image;
  bool autoClose;

  RichPopupMetadata({
    required super.id,
    required super.message,
    required this.header,
    required this.actionButtonText,
    required this.action,
    required this.image,
    super.title,
    this.autoClose = true,
  });
}
