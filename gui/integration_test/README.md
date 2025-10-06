# Integration tests

## How to work with tests

The main goal is to make those tests easy and pleasant to write.

### Main classes

There are two types of classes to work with:

1. `AppCtl`

It should be used for all activities which are not part of validation in the
test. For example, imagine you want to test that the user information on the
account page displays correct email, when you are connected to obfuscated server
(abstract scenario).
The steps you need to make in GUI are:

1. Open the app.
2. Wait to display the login page.
3. Login.
4. Go to settings.
5. Set technology to OpenVPN.
6. Enable obfuscated servers.
7. Go to the account page.
8. Validate that user information contains all you expect.

In all steps 1-7 the test can fail, but not because the user info does not
contain what you expect, but because you didn't even get to the account page.

That's bad - we don't want to test steps that are not critical for this test.
Moreover, if you have 10 tests which require the same steps, you have 10 tests
which can fail at the step which is not a target of your validation and 10
tests which execute the same code taking time.

`AppCtl` can be used to set specific application state before performing
the test validation:

- if the test doesn't need to check the logging in process, use `AppCtl` to
  perform "fake" login
- if you need to fake subscription expiration, use `AppCtl`
- if the tests needs specific settings (and you are not testing the act of
  changing the settings itself), use `AppCtl`

> [!NOTE]
> If the goal of your test is to verify that by clicking on the GUI does what
> you expect - that's a separate story and needs **separate** test. See next
> section for details.

1. `XyzScreenHandle`

This set of classes expose API to move around and click elements in the GUI.
Those APIs execute real actions and should be used when you are testing if those
actions do what you expect them to do.

If you need to operate on an account screen, get `AccountScreenHandle`, if you need
main screen, get `MainScreenHandle` etc.

So to sum up:

- if your test checks that with `some-conditions`, the app displays `something`,
  then use `AppCtl` to set `some-conditions`
- if your test checks if user can set `some-conditions` in the app, use
  `XyzScreenHandle` set of classes to navigate around app and set those
  conditions
