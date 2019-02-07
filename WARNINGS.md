# Buildifier warnings

Warning categories supported by buildifier's linter:

  * [attr-cfg](#attr-cfg)
  * [attr-license](#attr-license)
  * [attr-non-empty](#attr-non-empty)
  * [attr-output-default](#attr-output-default)
  * [attr-single-file](#attr-single-file)
  * [confusing-name](#confusing-name)
  * [constant-glob](#constant-glob)
  * [ctx-actions](#ctx-actions)
  * [ctx-args](#ctx-args)
  * [depset-iteration](#depset-iteration)
  * [depset-union](#depset-union)
  * [dict-concatenation](#dict-concatenation)
  * [duplicated-name](#duplicated-name)
  * [filetype](#filetype)
  * [function-docstring](#function-docstring)
  * [git-repository](#git-repository)
  * [http-archive](#http-archive)
  * [integer-division](#integer-division)
  * [load](#load)
  * [load-on-top](#load-on-top)
  * [module-docstring](#module-docstring)
  * [name-conventions](#name-conventions)
  * [native-build](#native-build)
  * [native-package](#native-package)
  * [no-effect](#no-effect)
  * [out-of-order-load](#out-of-order-load)
  * [output-group](#output-group)
  * [package-name](#package-name)
  * [package-on-top](#package-on-top)
  * [positional-args](#positional-args)
  * [redefined-variable](#redefined-variable)
  * [repository-name](#repository-name)
  * [return-value](#return-value)
  * [rule-impl-return](#rule-impl-return)
  * [same-origin-load](#same-origin-load)
  * [string-iteration](#string-iteration)
  * [uninitialized](#uninitialized)
  * [unreachable](#unreachable)
  * [unsorted-dict-items](#unsorted-dict-items)
  * [unused-variable](#unused-variable)

--------------------------------------------------------------------------------

## <a name="attr-cfg"></a>`cfg = "data"` for attr definitions has no effect

  * Category_name: `attr-cfg`
  * Flag in Bazel: [`--incompatible_disallow_data_transition`](https://docs.bazel.build/versions/master/skylark/backward-compatibility.html#disallow-cfg-data)
  * Automatic fix: yes

The [Configuration](https://docs.bazel.build/versions/master/skylark/rules.html#configurations)
`cfg = "data"` is deprecated and has no effect. Consider removing it.

--------------------------------------------------------------------------------

## <a name="attr-license"></a>`attr.license()` is deprecated and shouldn't be used

  * Category_name: `attr-license`
  * Flag in Bazel: `--incompatible_no_attr_license`
  * Automatic fix: no

The `attr.license()` method is almost never used and being deprecated.

--------------------------------------------------------------------------------

## <a name="attr-non-empty"></a>`non_empty` attribute for attr definitions are deprecated

  * Category_name: `attr-non-empty`
  * Flag in Bazel: `--incompatible_disable_deprecated_attr_params`
  * Automatic fix: yes

The `non_empty` [attribute](https://docs.bazel.build/versions/master/skylark/lib/attr.html)
for attr definitions is deprecated, please use `allow_empty` with an opposite value instead.

--------------------------------------------------------------------------------

## <a name="attr-output-default"></a>The `default` parameter for `attr.output()`is deprecated

  * Category_name: `attr-output-default`
  * Flag in Bazel: [`--incompatible_no_output_attr_default`](https://docs.bazel.build/versions/master/skylark/backward-compatibility.html#disable-default-parameter-of-output-attributes)
  * Automatic fix: no

The `default` parameter of `attr.output()` is bug-prone, as two targets of the same rule would be
unable to exist in the same package under default behavior. Use Starlark macros to specify defaults
for these attributes instead.

--------------------------------------------------------------------------------

## <a name="attr-single-file"></a>`single_file` is deprecated

  * Category_name: `attr-single-file`
  * Flag in Bazel: `--incompatible_disable_deprecated_attr_params`
  * Automatic fix: yes

The `single_file` [attribute](https://docs.bazel.build/versions/master/skylark/lib/attr.html)
is deprecated, please use `allow_single_file` instead.

--------------------------------------------------------------------------------

## <a name="confusing-name"></a>Never use `l`, `I`, or `O` as names

  * Category_name: `confusing-name`
  * Automatic fix: no

The names `l`, `I`, or `O` can be easily confused with `I`, `l`, or `0` correspondingly.

--------------------------------------------------------------------------------

## <a name="constant-glob"></a>Glob pattern has no wildcard ('*')

  * Category name: `constant-glob`
  * Automatic fix: no

[Glob function](https://docs.bazel.build/versions/master/be/functions.html#glob)
is used to get a list of files from the depot. The patterns (the first argument)
typically include a wildcard (* character). A pattern without a wildcard is
often useless and sometimes harmful.

To fix the warning, move the string out of the glob:

```
- glob(["*.cc", "test.cpp"])
+ glob(["*.cc"]) + ["test.cpp"]
```

**There’s one important difference**: before the change, Bazel would silently
ignore test.cpp if file is missing; after the change, Bazel will throw an error
if file is missing.

If `test.cpp` doesn’t exist, the fix becomes:

```
- glob(["*.cc", "test.cpp"])
+ glob(["*.cc"])
```

which improves maintenance and readability.

If no pattern has a wildcard, just remove the glob. It will also improve build
performance (glob can be relatively slow):

```
- glob(["test.cpp"])
+ ["test.cpp"]
```

### How to disable this warning

You can disable this warning by adding `# buildozer: disable=constant-glob` on
the line or at the beginning of a rule.

--------------------------------------------------------------------------------

## <a name="ctx-actions"></a>`ctx.{action_name}` is deprecated

  * Category_name: `ctx-actions`
  * Flag in Bazel: [`--incompatible_new_actions_api`](https://docs.bazel.build/versions/master/skylark/backward-compatibility.html#new-actions-api)
  * Automatic fix: yes

The following [actions](https://docs.bazel.build/versions/master/skylark/lib/actions.html)
are deprecated, please use the new API:

  * [`ctx.new_file`](https://docs.bazel.build/versions/master/skylark/lib/ctx.html#new_file) → [`ctx.actions.declare_file`](https://docs.bazel.build/versions/master/skylark/lib/actions.html#declare_file)
  * `ctx.experimental_new_directory` → [`ctx.actions.declare_directory`](https://docs.bazel.build/versions/master/skylark/lib/actions.html#declare_directory)
  * [`ctx.file_action`](https://docs.bazel.build/versions/master/skylark/lib/ctx.html#file_action) → [`ctx.actions.write`](https://docs.bazel.build/versions/master/skylark/lib/actions.html#write)
  * [`ctx.action(command = "...")`](https://docs.bazel.build/versions/master/skylark/lib/ctx.html#action) → [`ctx.actions.run_shell`](https://docs.bazel.build/versions/master/skylark/lib/actions.html#run_shell)
  * [`ctx.action(executable = "...")`](https://docs.bazel.build/versions/master/skylark/lib/ctx.html#action) → [`ctx.actions.run`](https://docs.bazel.build/versions/master/skylark/lib/actions.html#run)
  * [`ctx.empty_action`](https://docs.bazel.build/versions/master/skylark/lib/ctx.html#empty_action) → [`ctx.actions.do_nothing`](https://docs.bazel.build/versions/master/skylark/lib/actions.html#do_nothing)
  * [`ctx.template_action`](https://docs.bazel.build/versions/master/skylark/lib/ctx.html#template_action) → [`ctx.actions.expand_template`](https://docs.bazel.build/versions/master/skylark/lib/actions.html#expand_template)

--------------------------------------------------------------------------------

## <a name="ctx-args"></a>`ctx.actions.args().add()` for multiple arguments is deprecated

  * Category_name: `ctx-args`
  * Flag in Bazel: [`--incompatible_disallow_old_style_args_add`](https://docs.bazel.build/versions/master/skylark/backward-compatibility.html#new-args-api)
  * Automatic fix: yes

It's deprecated to use the [`add`](https://docs.bazel.build/versions/master/skylark/lib/Args.html#add)
method of `ctx.actions.args()` to add a list (or a depset) of variables. Please use either
[`add_all`](https://docs.bazel.build/versions/master/skylark/lib/Args.html#add_all) or
[`add_joined`](https://docs.bazel.build/versions/master/skylark/lib/Args.html#add_joined),
depending on the desired behavior.

--------------------------------------------------------------------------------

## <a name="depset-iteration"></a>Depset iteration is deprecated

  * Category_name: `depset-iteration`
  * Flag in Bazel: [`--incompatible_depset_is_not_iterable`](https://docs.bazel.build/versions/master/skylark/backward-compatibility.html#depset-is-no-longer-iterable)
  * Automatic fix: yes

Depsets are complex structures, iterations over them and lookups require flattening them to
a list which may be a heavy operation. To make it more obvious it's now required to call
the `.to_list()` method on them in order to be able to iterate their items:

    deps = depset()
    [x.path for x in deps]  # deprecated
    [x.path for x in deps.to_list()]  # recommended

--------------------------------------------------------------------------------

## <a name="depset-union"></a>Depsets should be joined using the depset constructor

  * Category_name: `depset-union`
  * Flag in Bazel: [`--incompatible_depset_union`](https://docs.bazel.build/versions/master/skylark/backward-compatibility.html#depset-union)
  * Automatic fix: no

The following ways to merge two depsets are deprecated:

    depset1 + depset2
    depset1 | depset2
    depset1.union(depset2)

Please use the [depset](https://docs.bazel.build/versions/master/skylark/lib/depset.html) constructor
instead:

    depset(transitive = [depset1, depset2])

--------------------------------------------------------------------------------

## <a name="dict-concatenation"></a>Dictionary concatenation is deprecated

  * Category_name: `dict-concatenation`
  * Flag in Bazel: [`--incompatible_disallow_dict_plus`](https://docs.bazel.build/versions/master/skylark/backward-compatibility.html#dictionary-concatenation)
  * Automatic fix: no

The `+` operator to concatenate dicts is deprecated. The operator used to create a new dict and
copy the data to it. There are several ways to avoid it, for example, instead of `d = d1 + d2 + d3`
you can use one of the following:

  * Use [Skylib](https://github.com/bazelbuild/bazel-skylib):

    load("@bazel_skylib//lib:dicts.bzl", "dicts")

    d = dicts.add(d1, d2, d3)

  * The same if you don't want to use Skylib:

    d = dict(d1.items() + d2.items() + d3.items())

  * The same in several steps:

    d = dict(d1)  # If you don't want `d1` to be mutated
    d.update(d2)
    d.update(d3)

--------------------------------------------------------------------------------

## <a name="duplicated-name"></a>A rule with name `foo` was already found on line

  * Category name: `duplicated-name`
  * Automatic fix: no

### Background

Each label in Bazel has a unique name, and Bazel doesn’t allow two rules to have
the same name. With macros, this may be accepted by Bazel (if each macro
generates different rules):

```
my_first_macro(name = "foo")
my_other_macro(name = "foo")
```

Although the build may work, this code can be very confusing. It can confuse
users reading a BUILD file (if they look for the rule “foo”, they may read see
only one of the macros). It will also confuse tools that edit BUILD files.

### How to fix it

Just change the name attribute of one rule/macro.

### How to disable this warning

You can disable this warning by adding `# buildozer: disable=duplicated-name` on
the line or at the beginning of a rule.

--------------------------------------------------------------------------------

## <a name="filetype"></a>The `FileType` function is deprecated

  * Category_name: `filetype`
  * Flag in Bazel: [`--incompatible_disallow_filetype`](https://docs.bazel.build/versions/master/skylark/backward-compatibility.html#filetype-is-deprecated)
  * Automatic fix: no

The function `FileType` is [deprecated](https://docs.bazel.build/versions/master/skylark/backward-compatibility.html#filetype-is-deprecated).
Instead of using it as an argument to the [`rule` function](https://docs.bazel.build/versions/master/skylark/lib/globals.html#rule)
just use a list of strings.

--------------------------------------------------------------------------------

## <a name="function-docstring"></a>Function docstring

  * Category_name: `function-docstring`
  * Automatic fix: no

Public functions should have docstrings describing functions and their signatures.
A docstring is a string statement which should be the first statement of the file
(it may follow comment lines). Docstrings are expected to be formatted in the
following way:

    """One-line summary: must be followed and may be preceded by a blank line.
    
    Optional additional description like this.
    
    If it's a function docstring and the function has more than one argument, the docstring has
    to document these parameters as follows:

    Args:
      parameter1: description of the first parameter. Each parameter line
        should be indented by one, preferably two, spaces (as here).
      parameter2: description of the second
        parameter that spans two lines. Each additional line should have a
        hanging indentation of at least one, preferably two, additional spaces (as here).
      another_parameter (unused, mutable): a parameter may be followed
        by additional attributes in parentheses

    Returns:
      Description of the return value.
      Should be indented by at least one, preferably two spaces (as here)
      Can span multiple lines.
    """

--------------------------------------------------------------------------------

## <a name="git-repository"></a>Function `git_repository` is not global anymore

  * Category_name: `git-repository`
  * Flag in Bazel: [`--incompatible_remove_native_git_repository`](https://docs.bazel.build/versions/master/skylark/backward-compatibility.html#remove-native-git-repository)
  * Automatic fix: yes

Native `git_repository` and `new_git_repository` functions are [being removed](https://docs.bazel.build/versions/master/skylark/backward-compatibility.html#remove-native-git-repository).
Please use the Starklark versions instead:

    load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository", "new_git_repository")

--------------------------------------------------------------------------------

## <a name="http-archive"></a>Function `http_archive` is not global anymore

  * Category_name: `http-archive`
  * Flag in Bazel: [`--incompatible_remove_native_http_archive`](https://docs.bazel.build/versions/master/skylark/backward-compatibility.html#remove-native-http-archive)
  * Automatic fix: yes

Native `http_archive` function are [being removed](https://docs.bazel.build/versions/master/skylark/backward-compatibility.html#remove-native-http-archive).
Please use the Starklark versions instead:

    load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

--------------------------------------------------------------------------------

## <a name="integer-division"></a>The `/` operator for integer division is deprecated

  * Category_name: `integer-division`
  * Flag in Bazel: [`--incompatible_disallow_slash_operator`](https://docs.bazel.build/versions/master/skylark/backward-compatibility.html#integer-division-operator-is)
  * Automatic fix: yes

The `/` operator is deprecated in favor of `//`, please use the latter for
integer division:

`
a = b // c
d //= e
`

--------------------------------------------------------------------------------

## <a name="load"></a>Loaded symbol is unused

  * Category_name: `load`
  * Automatic fix: yes

### Background

[load](https://docs.bazel.build/versions/master/skylark/concepts.html#loading-an-extension)
is used to import definitions in a BUILD file. If the definition is not used in
the file, the load can be safely removed. If a symbol is loaded two times, you
will get a warning on the second occurrence.

### How to fix it

Delete the line. When load is used to import multiple symbols, you can remove
the unused symbols from the list. To fix your BUILD files automatically, try
this command:

```
buildozer 'fix unusedLoads' path/to/BUILD
```

If you want to keep the load, you can disable the warning by adding a comment
`# @unused`.

### How to disable this warning

You can disable this warning by adding `# buildozer: disable=load` on the line
or at the beginning of a rule.

--------------------------------------------------------------------------------

## <a name="load-on-top"></a>Load statements should be at the top of the file.

  * Category_name: `load-on-top`
  * Flag in Bazel: [`--incompatible_bzl_disallow_load_after_statement`](https://docs.bazel.build/versions/master/skylark/backward-compatibility.html#load-must-appear-at-top-of-file)
  * Automatic fix: yes

Load statements should be first statements (with the exception of `WORKSPACE` files),
they can follow only comments and docstrings.

--------------------------------------------------------------------------------

## <a name="module-docstring"></a>The file has no module docstring.

  * Category_name: `module-docstring`
  * Automatic fix: no

`.bzl` files should have docstrings on top of them. A docstring is a string statement
which should be the first statement of the file (it may follow comment lines). 

--------------------------------------------------------------------------------

## <a name="name-conventions"></a>Name conventions

  * Category_name: `name-conventions`
  * Automatic fix: yes

By convention, all variables should be lower_snake_case, constant should be
UPPER_SNAKE_CASE, and providers should be UpperCamelCase ending with `Info`.

--------------------------------------------------------------------------------

## <a name="native-build"></a>The `native` module shouldn't be used in BUILD files

  * Category_name: `native-build`
  * Automatic fix: yes

There's no need in using `native.` in BUILD files, its members are available as global symbols
there.

--------------------------------------------------------------------------------

## <a name="native-package"></a>`native.package()` shouldn't be used in .bzl files

  * Category_name: `native-package`
  * Automatic fix: no

It's discouraged and will be disallowed to use `native.package()` in .bzl files. It can silently
modify the semantics of a BUILD file and makes it hard to maintain.

--------------------------------------------------------------------------------

## <a name="no-effect"></a>Expression result is not used

  * Category_name: `no-effect`
  * Automatic fix: no

The statement has no effect. Consider removing it or storing its result in a
variable.

--------------------------------------------------------------------------------

## <a name="out-of-order-load"></a>Load statements should be ordered by their labels.

  * Category_name: `out-of-order-load`
  * Automatic fix: yes

Load statements should be ordered by their first argument - extension file label.
This makes it easier to developers to locate loads of interest and reduces chances
for conflicts when performing large-scale automated refactoring.

When applying automated fixes, it's highly recommended to also use
[`load-on-top`](#load-on-top) fixes, since otherwise the relative order
of a symbol load and its usage can change resulting in runtime error.

--------------------------------------------------------------------------------

## <a name="unsorted-dict-items"></a>Dictionary items should be ordered by their keys.

  * Category_name: `unsorted-dict-items`
  * Automatic fix: yes

Dictionary items should be sorted lexicagraphically by their keys. This makes
it easier to find the item of interest and reduces chances of conflicts when
performing large-scale automated refactoring.

The order is affected by `NamePriority` dictionary passed using `-tables` or
`-add_tables` flags.

If you want to preserve the original dictionary items order, you can disable
the warning by adding a comment `# @unsorted-dict-items` to the dictionary
expression or any of its enclosing expressins (binary, if etc). For example,

    # @unsorted-dict-items
    d = {
      "b": "bvalue",
      "a": "avalue",
    }

will not be reported as an issue because the assignment operation that uses
the dictionary with unsorted items has a comment disabling this warning.

--------------------------------------------------------------------------------

## <a name="output-group"></a>`ctx.attr.dep.output_group` is deprecated

  * Category_name: `output-group`
  * Flag in Bazel: [`--incompatible_no_target_output_group`](https://docs.bazel.build/versions/master/skylark/backward-compatibility.html#disable-output-group-field-on-target)
  * Automatic fix: yes

The `output_group` field of a target is [deprecated](https://docs.bazel.build/versions/master/skylark/backward-compatibility.html#disable-output-group-field-on-target)
in favor of the [`OutputGroupInfo` provider](https://docs.bazel.build/versions/master/skylark/lib/OutputGroupInfo.html).

--------------------------------------------------------------------------------

## <a name="package-name"></a>Global variable `PACKAGE_NAME` is deprecated

  * Category_name: `package-name`
  * Flag in Bazel: [`--incompatible_package_name_is_a_function`](https://docs.bazel.build/versions/master/skylark/backward-compatibility.html#package-name-is-a-function)
  * Automatic fix: yes

The global variable `PACKAGE_NAME` is deprecated, please use
[`native.package_name()`](https://docs.bazel.build/versions/master/skylark/lib/native.html#package_name)
instead.

--------------------------------------------------------------------------------

## <a name="package-on-top"></a>Package declaration should be at the top of the file

  * Category_name: `package-on-top`
  * Automatic fix: no

Here is a typical structure of a BUILD file:

*   `load()` statements
*   `package()`
*   calls to rules, macros

Instantiating a rule and setting the package defaults later can be very
confusing, and has been a source of bugs (tools and humans sometimes believe
package applies to everything in a BUILD file). This might become an error in
the future (but it requires large-scale changes in google3).

### What can be used before package()?

The linter allows the following to be before `package()`:

*   comments
*   `load()`
*   variable declarations
*   `package_group()`
*   `licenses()`

### How to disable this warning

You can disable this warning by adding `# buildozer: disable=package-on-top` on
the line or at the beginning of a rule.

--------------------------------------------------------------------------------

## <a name="positional-args"></a>Keyword arguments should be used over positional arguments

  * Category_name: `positional-args`
  * Automatic fix: no

All top level calls (except for some built-ins) should use keyword args over
positional arguments. Positional arguments can cause subtle errors if the order
is switched or if an argument is removed. Keyword args also greatly improve
readability.

```
- my_macro("foo", "bar")
+ my_macro(name = "foo", env = "bar")
```

The linter allows the following functions to be called with positional
arguments:

*   `load()`
*   `vardef()`
*   `export_files()`
*   `licenses()`
*   `print()`

### How to disable this warning

You can disable this warning by adding `# buildozer: disable=positional-args` on
the line or at the beginning of a rule.

--------------------------------------------------------------------------------

## <a name="redefined-variable"></a>Variable has already been defined

  * Category_name: `redefined-variable`
  * Automatic fix: no

### Background

In .bzl files, redefining a global variable is already forbidden. This helps
both humans and tools reason about the code. For consistency, we want to bring
this restriction also to BUILD files.

### How to fix it

Rename one of the variables.

Note that the content of lists and dictionaries can still be modified. We will
forbid reassignment, but not every side-effect.

### How to disable this warning

You can disable this warning by adding `# buildozer: disable=unused-variable` on
the line or at the beginning of a rule.

--------------------------------------------------------------------------------

## <a name="repository-name"></a>Global variable `REPOSITORY_NAME` is deprecated

  * Category_name: `repository-name`
  * Flag in Bazel: [`--incompatible_package_name_is_a_function`](https://docs.bazel.build/versions/master/skylark/backward-compatibility.html#package-name-is-a-function)
  * Automatic fix: yes

The global variable `REPOSITORY_NAME` is deprecated, please use
[`native.repository_name()`](https://docs.bazel.build/versions/master/skylark/lib/native.html#repository_name)
instead.

--------------------------------------------------------------------------------

## <a name="return-value"></a>Some but not all execution paths of a function return a value

  * Category_name: `return-value`
  * Automatic fix: no

Some but not all execution paths of a function return a value. Either there's
an explicit empty `return` statement, or an implcit return in the end of a
function. If it is intentional, make it explicit using `return None`. If you
know certain parts of the code cannot be reached, add the statement
`fail("unreachable")` to them.

--------------------------------------------------------------------------------

## <a name="rule-impl-return"></a>Avoid using the legacy provider syntax

  * Category_name: `rule-impl-return`
  * Automatic fix: no

Returning structs from rule implementation functions is
<a href="https://docs.bazel.build/versions/master/skylark/rules.html#migrating-from-legacy-providers">deprecated</a>,
consider using
<a href="https://docs.bazel.build/versions/master/skylark/rules.html#providers">providers</a>
or lists of providers instead.

--------------------------------------------------------------------------------

## <a name="same-origin-load"></a>Same label is used for multiple loads

  * Category_name: `same-origin-load`
  * Automatic fix: yes

### Background

[load](https://docs.bazel.build/versions/master/skylark/concepts.html#loading-an-extension)
is used to import definitions in a BUILD file. If the same label is used for loading
symbols more the ones, all such loads can be merged into a single one.

### How to fix it

Merge all loads into a single one. For example,

```
load(":f.bzl", "s1")
load(":f.bzl", "s2")
```

can be written more compactly as

```
load(":f.bzl", "s1", "s2")
```

--------------------------------------------------------------------------------

## <a name="string-iteration"></a>String iteration is deprecated

  * Category_name: `string-iteration`
  * Flag in Bazel: [`--incompatible_string_is_not_iterable`](https://docs.bazel.build/versions/master/skylark/backward-compatibility.html#string-is-no-longer-iterable)
  * Automatic fix: no

Iteration over strings often leads to confusion with iteration over a sequence of strings,
therefore strings won't be recognized as sequences of 1-element strings (like in Python).
Use string indexing and `len` instead:

    my_string = "hello world"
    for i in range(len(my_string)):
        char = my_string[i]
        # do something with char

--------------------------------------------------------------------------------

## <a name="uninitialized"></a>Variable may not have been initialized

  * Category_name: `uninitialized`
  * Automatic fix: no

The local value can be not initialized at the time of execution. It may happen if it's
initialized in one of the if-else clauses but not in all of them, or in a for-loop which
can potentially be empty.

--------------------------------------------------------------------------------

## <a name="unreachable"></a>The statement is unreachable

  * Category_name: `unreachable`
  * Automatic fix: no

The statement is unreachable because it follows a `return`, `break`, `continue`,
or `fail()` statement.

--------------------------------------------------------------------------------

## <a name="unused-variable"></a>Variable is unused

  * Category_name: `unused-variable`
  * Automatic fix: no

This happens when a variable is set but not used in the file, e.g.

```
x = [1, 2]
```

The line can often be safely removed.

If you want to keep the variable, you can disable the warning by adding a
comment `# @unused`.

```
x = [1, 2] # @unused
```

### How to disable this warning

You can disable this warning by adding `# buildozer: disable=unused-variable` on
the line or at the beginning of a rule.
