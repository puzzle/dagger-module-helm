#!/bin/sh

#################################################
# Example on how to call the Test method.
# Run the unit tests for the Helm Chart located inside the directory referenced by the directory parameter.
# Add the directory location with `"."` as `--args` parameter to tell helm unittest where to find the tests inside the passed directory.
# ARGUMENTS:
#   directory: directory that contains the Helm Chart
#   args: arguments for the helm test command
# RETURN:
#   The Helm unit test output as string.
#################################################
function helm_test() {
    dagger -m helm/ \
        call test \
            --directory ./helm/examples/testdata/mychart/ \
            --args "."
}


#################################################
# Example on how to call the Version method.
# Get and display the version of the Helm Chart located inside the directory referenced by the directory parameter.
# ARGUMENTS:
#   directory: directory that contains the Helm Chart
# RETURN:
#   The Helm Chart version as string.
#################################################
function helm_version() {
    dagger -m helm/ \
        call version \
            --directory ./helm/examples/testdata/mychart/
}
