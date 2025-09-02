# Table of contents

- [How to setup the development environment](#how-to-setup-the-development-environment)
- [Building](#building)
- [Testing](#testing)
- [Internationalization](#internationalization)
- [Working with Riverpod](#working-with-riverpod)
- [Linting](#linting)

## How to setup the development environment

Please follow the instructions for setting up the development environment:

1. Install [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
2. Install [flutter](https://docs.flutter.dev/get-started/install/linux)
3. Install [NordVPN](https://support.nordvpn.com/hc/en-us/articles/20398283005457-Installing-NordVPN-on-RHEL-and-CentOS-Linux)
4. (optional) Install [protoc](https://pub.dev/packages/protoc_plugin) to be
   able to compile protobuf files.
5. (optional) Configure the IDE to use [flutter plugin](https://docs.flutter.dev/get-started/editor).
6. (optional) Install [docker](https://www.docker.com/) to be able to compile
   protobuf files into the environment used by us.
7. (optional) Install [rps](https://pub.dev/packages/rps) to use the predefined
   scripts. e.g. rps generate protobuf

## Dev tools

Install `rps`:

```bash
dart pub global activate rps
```

Add to the path:

```bash
export PATH="$PATH":"$HOME/.pub-cache/bin"
```

Show list of commands:

```bash
rps ls
```

## Building

Before building check that `flutter doctor` doesn't return any errors.

To compile and run the application use `flutter run`.

## Testing

1. All tests

- use `rps test all`

1. Integration tests

- for details on the integration tests approach see [tests README.md](./integration_test/README.md)
- use `rps test int`

1. Unit tests

- use `rps test unit`

1. Code Coverage

- use `rps test cov`

[!NOTE]
You need `genhtml` installed to produce HTML report from coverage info.

## Internationalization

The application utilizes the [slang](https://pub.dev/packages/slang) package
for internationalization. The English version of all the strings used in the
application is stored in [en.i18n.json](lib/i18n/en.i18n.json). After modifying
this file, the translation file need be regenerated using either
`rps generate translations` or `dart run slang`.

For handling dynamic strings, the tr() function from
[translation.dart](lib/translation.dart) is available. This function attempts
to translate the string, and if not found, it returns the original value. For
static strings like labels, it is recommended to use the variables generated
from the JSON, e.g.: `t.filename.keyFromJsonFile`

## Working with Riverpod

When changes are made to classes that use Riverpod annotations, it is necessary
to regenerate the corresponding auto-generated counterparts. This can be
achieved through two methods:

- Automatic Generation:
  Files are automatically regenerated when changes are detected by manually
  starting and running `dart run build_runner watch` in the background. This
  process continually monitors files, automatically regenerating them upon
  modification.
- Manual Generation:
  Regenerate files manually by executing either `dart run build_runner build`
  or `rps generate code`. This manual approach ensures that the auto-generated
  files are updated according to the modifications made.
- Visual Studio Code (VSCode) Integration:
  In Visual Studio Code the files changes monitoring process starts automatically
  in the background when the project is loaded. This allows for seamless
  regeneration of files as soon as modifications are detected.

## Linting

Use [dart fix](https://dart.dev/tools/dart-fix) to to analyze and fix the issues.
