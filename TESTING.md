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

### Running QA tests
In order to run QA tests, run `test:qaDocker` or `test:qaDockerFast`. It requires two arguments:
* name of the test category;
* pattern for test functions to run.
For test category names, check out [test/qa](test/qa) directory. Name of the test suite would be
the name of the desired test file with *test* omitted. For test name, simply provide the desired
function name from that file. Test names are selected by pattern matching. Since all test function
names start with `test_`, simply use `test` as an argument to run every test in the suite.

For example, to run every test from the `test_fileshare.py` category:

`mage test:qaDocker fileshare test`

And to run a single test:

`mage test:qaDocker fileshare test_accept`

It is possible to run multiple categories with one command by adding all the categories as the first argument in a string seperated by spaces:

`mage test:qaDocker "fileshare meshnet" test`

To run tests without rebuilding everything each time use `test:qaDockerFast` instead of
`test:qaDocker`.

### Test credentials for QA tests
`NA_TESTS_CREDENTIALS` environment variable is used to configure test credentials. It is a JSON
object containing a key(string):value(credentials) map.
For running the tests using Vagrant for the snap build the following environment variable are required:
* `SNAP_TEST_BOX_NAME` - represents the configuration name used by vagrant from [Vagrantfile](ci/snap/vagrant/Vagrantfile).
* `SNAP_TEST_DESTROY_VM_ON_EXIT` - [optional] a boolean value that controls if vagrant should destroy the created virtual machine after running the tests. By default it is `false`.
When testing the snap application using Vagrant `SNAP_TEST_BOX_NAME` environment variable must be defined

#### key
QA tests use a list of the following keys:
* `default` - default account used for most of tests. In meshnet and fileshare tests is used as an
account for "this" device.
* `qa-peer` - account used for fileshare and meshnet tests for "another peer". This account will be
invited by the `default` account to its' meshnet. Check out [qa peer README](ci/docker/qa-peer/README.md)
for more details.
* `valid` - valid account used in login tests. It has to be different than default account.
* `expired` - expired account is an account whose subscription has expired. Used in login tests.

Any of the used keys can be extended with `NA_CREDENTIALS_KEY` environment variable which will be
appended to the used key.
E. g. when `NA_CREDENTIALS_KEY` is set to `my_key`, `default_my_key` and `qa-peer-my_key` will be
used instead of `default` and `qa-peer` in the executed tests.

#### credentials
Credentials structure consists of the following fields:
* `token *` - token to be used to log in during the QA test execution. It can be acquired in the
[nordvpn account dashboard](https://my.nordaccount.com/dashboard/nordvpn/).
* `email` - email of the account to run tests with. In meshnet tests it is used to send and accept
invitations. In other cases it is used for logging.
* `password` - password of the account used in the login tests.

#### Note about meshnet tests
In case of meshnet tests(`test_meshnet_*.py`), two NordVPN accounts might be required(`qa-peer` and
`default`). Meshnet functionality does not require a subscription, so a secondary free account can
be used in this case.

## Running unit tests
You can run unit tests for a single package as you would run uts for any other go project, for
example:
```
cd nordvpn-app/meshnet
go test
```
We also have a mage targets for running tests for all of the packages:
* `test:cgoDocker` runs cgo tests.
* `test:go` runs regular unit tests.


### Coverage

QA tests should generally cover app's use cases without necessarily covering all of the edge cases.

It's okay to introduce new QA tests on bug fixes as well if the bug is easier to reproduce that way.
