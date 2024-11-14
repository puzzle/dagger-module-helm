#!/bin/sh

#################################################
# Example on how to call the Version method.
# Get and display the version of the Helm Chart located inside the directory referenced by the directory parameter.
#################################################
function helm_version() {
    dagger -m helm/ \
        call version \
            --directory ./helm/examples/testdata/mychart/ # directory that contains the Helm Chart
}
