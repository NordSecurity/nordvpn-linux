# Automated testing policy

There are 2 types of tests in this project:
1. Unit tests
    * Defined in `*_test.go` files along the source code.
1. QA (E2E) tests
    * Defined in `test/qa` directory.

The following information or guidelines may not apply to the existing code written before introduction of this document.

## Unit tests

Unit tests are the preffered way of testing. Generally all new code should be covered with unit tests. Well-grounded exceptions are allowed, but the code should still be written in a way so that as much as possible of it could be unit tested and the rest should be covered by QA tests.

### Test categories

Test categories are used to separate tests that interact with the environment from the ones that don't. Available categories are defined and documented in `test/category/category.go`.

All unit tests must start with `category.Set(t, category.XXX)` statement.  

### Table driven tests

https://go.dev/blog/subtests

Table driven tests are the preferred way to test new functions as it's a convenient way to test all execution branches. On the other hand simple separate tests are okay too if that seems more convenient, but consider table driven ones if there's a lot of repeatability between tests.

### Coverage

The unit test coverage of newly written code should be >50%. That's just the mandatory bare minimum though, because generally it is expected that all of the code is covered. 

Possible exceptions:
* Code that integrates with the OS and external libraries (keep such code minimal and mock such integrations in other tests)

### Mocking

Because the project uses many external dependencies, the usage of mocks in tests is wide spread. 

* The mock of interface A used for testing can be defined in one of two places:
  * In `*_test.go` file in a package that depends on the interface.
  * In `/test/mock/a.go` file (or `/test/mock/a/a.go` if there are circular dependency issues) if the mock is exported.
* Generally if the interface is big then it's better to export the mock so that it could be reused.

### Other guidelines

* A test reproducing a bug must be introduced prior to fixing the bug if possible (possible exceptions: bug reproduces only on untested distribution, etc.)

## QA tests

QA tests cover base app functionality and integrations. As they run slower than unit tests they shouldn't be used to cover the testing cases that can be covered by unit tests.

It is recommended to run these tests in Docker container, because they might change operating system environment unexpectedly. No persistent changes should be made though, so nothing that a reboot wouldn't fix.

### Coverage

QA tests should generally cover app's use cases without necessarily covering all of the edge cases.

It's okay to introduce new QA tests on bug fixes as well if the bug is easier to reproduce that way.
