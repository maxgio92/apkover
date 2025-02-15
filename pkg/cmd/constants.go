package cmd

const (
	usage         = "apkover"
	melangeRunner = "bubblewrap"
	description   = `
  ___ ______ _   __                    
 / _ \| ___ | | / /                    
/ /_\ | |_/ | |/ /  _____   _____ _ __ 
|  _  |  __/|    \ / _ \ \ / / _ | '__|
| | | | |   | |\  | (_) \ V |  __| |   
\_| |_\_|   \_| \_/\___/ \_/ \___|_|   
                                       
APKover is a tool to extract the test coverage of a package as the percentage
of statements of the packaged software that are covered by the Melange tests.

It leverages the language specific instrumentation features to generate integration tests coverage,
by re-building the package to enable the software measure the coverage data at runtime.
Finally, it reports the overall package test coverage as a percentage.

A minimum threshold can be set to make APKover exit 1 when a package does not meet the coverage criteria.
`
)
