# Bazel

Bazel is an OSS build system developed by Google.

# Buildifier

There exists a couple of developer tools for working with Google's `bazel` buildsystem. Buildifier is one of them for formatting bazel BUILD, BUILD.bazel and BULK files with a standard convention.


## Setup

* Install Bazel: https://docs.bazel.build/versions/4.0.0/install.html
* Install Go: https://golang.org/doc/install

### Get the code:

* Download the tar uploaded and untar it. Or,
* Checkout the repo(https://gitlab.eng.vmware.com/rtabassum/mirrors_github_bazel-buildtools) which is forked from VMware internal gitlab mirrors repo:

	$ git clone git@gitlab.eng.vmware.com:rtabassum/mirrors_github_bazel-buildtools.git
	$ cd mirrors_github_bazel-buildtools

	Checkout the branch `osborathon-magic-table`. Its pulled from the latest tag `4.0.1`
	$ git checkout osborathon-magic-table


### Build the tool:

	$ bazel build //buildifier

The binary should be built at this location: `bazel-bin/buildifier/buildifier_/buildifier`


### Usage:

Use buildifier to create standardized formatting for BUILD and .bazel src files:

    $ bazel-bin/buildifier/buildifier_/buildifier path/to/file

You can also add the built binary to the PATH and directly invoke:

	$ buildifier path/to/file

You can also process multiple files at once:

    $ buildifier path/to/file1 path/to/file2

You can make buildifier automatically find all Starlark files (i.e. BUILD, WORKSPACE, .bzl)
in a directory recursively:

    $ buildifier -r path/to/dir

To utilize the new tags being added for formating tabular data add tag in your src file above the tabular data.
`# buildifier: table`.

Buildifier will properly format the tabular data when you run the Buildifier on your src file.

	$ bazel-bin/buildifier/buildifier_/buildifier path/to/file

Also, to sort the table in your source file by a specific column, use tag "# buildifier: table sort <column_index>"



## Run unittests

You can run all the default unittests:

$ bazel test --test_output=all //build:go_default_test