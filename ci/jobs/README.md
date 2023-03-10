# CI Jobs

This directory contains GitLab job definitions so that `.gitlab-ci.yml` in the
project root would only contain includes and stage ordering.

Each file in this directory is named after a stage in the pipeline or starts
with a dot and is used in job definitions instead of including directly.
File name should be a verb and job name should start with an object so that
fully namespaced job name would read like a short sentence.

Namespacing is implemented by prefixing jobs with file name they're defined in.
Prefix and the job name is separated by a forward slash.
Job names are written in kebab case.

Code is formatted using [yamlfmt](https://github.com/google/yamlfmt).

## Best Practices

- [Use `rules` over `only`](#first-rule)
- [Use `!reference` or `matrix` over `extends`](#second-rule)
- [Always set `dependencies` attribute](#third-rule)
- [Never set `before_script` attribute](#fourth-rule)
- [Avoid DRY, unless it fixes GitLab's API](#fifth-rule)
- [Prefer REF version of GitLab's environment variables](#sixth-rule)
- [Always set `script` to a single item](#seventh-rule)
- [Always disable remote included jobs](#eighth-rule)

### Use `rules` over `only` <a name="first-rule"></a>
`rules` is a more flexible and newer version of `only`, which effective deprecates it.
Also, `rules` and `only` cannot be used in the same definition anyway, so there is no
reason to keep using `only` anymore.

### Use `!reference` or `matrix` over `extends` <a name="second-rule"></a>
This is basically composition over inheritance. `!reference` allows more granularity,
since specific fields can be picked, while `extends` just takes everything and forces
to override the unneeded parts. Also, `extends` does not work with attributes whose
values are lists, since it only inherits the last value of the list. There is no good
reason to use `extends` anymore. `parallel: matrix` can be used to implement generic
jobs, which are parameterized based on a list of input variables. Variables defined
in `parallel: matrix` block cannot be used in `dependencies` section, which means that
those jobs, which produce artifacts must have unique path for artifacts, since when
using parallel job in `dependencies` all of the artifacts will be downloaded, which
means that if the path is the same, the last artifact in the list will override the rest.

### Always set `dependencies` attribute <a name="third-rule"></a>
`dependencies` control what artifacts from the previous job could be used. By default,
it uses everything that git considers unstaged, even things that weren't specified in
the `artifacts` section of the job, which means that it introduces implicit ordering
dependencies. Therefore `dependencies` must always be set, even if the value is `[]`,
which means that the job doesn't have any dependencies.

### Never set `before_script` attribute <a name="fourth-rule"></a>
`before_script` accepts a script file or inline script to be executed before the
job. The pitfall is that any environment variables exported by the script file
are not visible to the job afterwards, which is primary usecase most of the time.

### Avoid DRY, unless it fixes GitLab's API <a name="fifth-rule"></a>
GitLab has many undocumented or poorly documented edge cases when it comes to it's
features. Excessive use of DRY forces the user to rely on those features without
understanding the consequences and if the added complexity really pays off in the end.
Job expansion rules when using something like `extends` is just one of the examples.
Even if the code is verbose or repetitive and violates DRY, a lot of times it is still
better than having to sit and debug for hours until painful realization that the
feature has a really nasty and non obvious edge case, or the name does not match what
it actually does. There is one case where DRY actually helps, fixing GitLab's API
by giving human readable names to poorly named combinations of built-in environment
variables used in conditional statements.

### Prefer REF version of GitLab's environment variables] <a="sixth-rule"></a>
GitLab's [predefined](https://docs.gitlab.com/ee/ci/variables/predefined_variables.html)
environment variables are problematic when trying to extract branch/commit information.
They force the user to handle regular and merge request pipelines in a different way,
even if all the user wants is a branch name. For example:
```
CI_COMMIT_BRANCH	12.6	0.5	The commit branch name. Available in branch pipelines, including pipelines for the default branch. Not available in merge request pipelines or tag pipelines.
```
and
```
CI_MERGE_REQUEST_SOURCE_BRANCH_NAME	11.6	all	The source branch name of the merge request.
```
Using REF version removes this distinction:
```
CI_COMMIT_REF_NAME	9.0	all	The branch or tag name for which project is built.
```

### Always set `script` to a single item <a name="seventh-rule"></a>
Ideally that single item should be path to a script file. If that is not possible,
that single item should be multiline string, for example:
```yaml
script:
  - >
    set -euox;
    doSomething;
    exit 0;
```
The only caveat is that each line has to be terminated by a semicolon.
Without multiline string it would be impossible to use bash's strict mode.

### Always disable remote included jobs <a name="eighth-rule"></a>
Included jobs are enabled by default. In order to use `!reference`, the job must
be included. Always disable included jobs that come from external repos and
`!reference` needed attributes instead.
