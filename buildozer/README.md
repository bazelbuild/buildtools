# Buildozer

Buildozer is a command line tool to refactor multiple [Bazel](https://github.com/bazelbuild/bazel) BUILD files using standard commands. 

### Dependencies

1. Protobuf go runtime: to download  
`go get -u github.com/golang/protobuf/{proto,protoc-gen-go}`


### Installation

1. Change directory to the buildifier/buildozer

```bash
gopath=$(go env GOPATH)
cd $gopath/src/github.com/bazelbuild/buildifier/buildozer
```

2. Install
```bash
go install
```

### Usage
`buildozer [OPTIONS] ['command args' | -f FILE ] label-list`

Here, `label-list` is a comma separated list of Bazel labels, for example `//path/to/pkg1:rule1, //path/to/pkg2:rule2`. Buildozer reads commands from FILE ( - for stdin (format:|-separated command line arguments to buildozer, excluding flags))

OPTIONS can be one of the following:

`--stdout` : write changed BUILD file to stdout  
`--buildifier` : path to buildifier binary  
`-P cores` : number of cores to use for concurrent actions  
`--numio N` : number of concurrent actions  
`-k` : apply all commands, even if there are failures  
`--eol-comments` : when adding a new comment, put it on the same line if possible  
`--root_dir` : If present, use this folder rather than $PWD to find the root directory  
`--quiet` : suppress informational messages  
`--edit-variables` : For attributes that simply assign a variable (e.g. hdrs = LIB_HDRS), edit the build variable instead of appending to the attribute  
`--output_proto` : output serialized devtools.buildozer.Output protos instead of human-readable strings  
`--shorten_labels` : convert added labels to short form, e.g. //foo:bar => :bar  
`--delete_with_comments` : If a list attribute should be deleted even if there is a comment attached to it  

Buildozer supports the following commands(`'command args'`):
- `add <attr> <value(s)>`
- `new_load <path> <symbol(s)>`
- `del_subinclude <label>`
- `replace_subinclude <label> <bzl-path> <symbol(s)>`
- `comment <attr>? <value>? <comment>`
- `print_comment <attr>? <value>?`
- `delete`
- `fix <fix(es)>?`
- `move <old_attr> <new_attr> <value(s)>`
- `new <rule_kind> <rule_name> [(before|after) <relative_rule_name>]`
- `print <attribute(s)>`
- `remove <attr> <value(s)>`
- `rename <old_attr> <new_attr>`
- `replace <attr> <old_value> <new_value>`
- `set <attr> <value(s)>`
- `set_if_absent <attr> <value(s)>`
- `copy <attr> <from_rule>`
- `copy_no_overwrite <attr> <from_rule>`

Here, `<attr>` represents an attribute (being `add`ed/`rename`d/`delete`d etc.), for eg: `srcs`, `<values(s)>`  represents values of the attribute and so on. A '?' indicates that the preceding argument is optional.

The fix command without a fix specified applied all eligible fixes.  
Use `//path/to/pkg:__pkg__` as label for file level changes like `new_load` and `new_rule`.  
A transformation can be applied to all rules of a particular kind by using `%rule_kind` at the end of the label(see examples below).


### Examples

```bash
# Edit //pkg:rule and //pkg:rule2, and add a dependency on //base
buildozer 'add deps //base' //pkg:rule //pkg:rule2

# A load for a skylark file in //pkg
buildozer 'new_load /tools/build_rules/build_test build_test' //pkg:__pkg__

# Change the default_visibility to public for the package //pkg
buildozer 'set default_visibility //visibility:public' //pkg:__pkg__

# Change all gwt_module targets to java_library in the package //pkg
buildozer 'set kind java_library' //pkg:%gwt_module

# Replace the dependency on pkg_v1 with a dependency on pkg_v2
buildozer 'replace deps //pkg_v1 //pkg_v2' //pkg:rule

# Add a cc_binary rule named new_bin before the rule named tests
buildozer 'new cc_binary new_bin before tests' //:__pkg__

# Add an attribute new_attr with value "def_val" to all cc_binary rules
# Note that special characters will automatically be escaped in the string
buildozer 'add new_attr def_val' //:%cc_binary

```

### Source Structure

`buildozer/main.go` : Entry point for the buildozer binary  
`edit/buildozer.go` : Implementation of functions for the buildozer commands  
`edit/edit.go`: Library functions to perform various operations on ASTs. These functions are called by the impl functions in buildozer.go  
`edit/fix.go`:  Functions for various fixes for the `buildozer 'fix <fix(es)>'` command, like cleaning unused loads, changing labels to canonical notation, etc.  
`edit/types.go`: Type information for attributes  

