# CI scripts

## Best Practices

- [Always start scripts with `set -euo`](#first-rule)
- [Each script should source it's dependencies](#second-rule)
- [Avoid default values for undefined variables](#third-rule)
- [Set and export variables on separate lines](#fourth-rule)

### Always start scripts with `set -euo` <a name="first-rule"></a>
This is the default behavior of majority of modern programming languages.

`set -e` exits on the first error, which is basically an unhandled exception.
`set -u` exits on unassigned/undefined variable instead of using empty string.
`set -o` exits on the first error with an actual exit code and not 0.

### Each script should source it's dependencies] <a name="second-rule"></a>
This is also the default behavior of most modern programming languages.
You can't use anything that's not in scope, so you have to import it. Relying
on the execution environment to have it imported already is also very brittle.

### Avoid default values for undefined variables <a name="third-rule"></a>
This silences `set -u` setting and makes debugging way harder and time consuming.

### Set and export variables on separate lines <a name="fourth-rule"></a>
Every line in bash can have only a single exit code.
```bash
export ONE_LINE=$(false) # exit code 0, because the second one is discarded

TWO_LINE=$(false) # exit code 1
export $TWO_LINE # exit code 0, but this would not be executed if `set -e` was provided
```
