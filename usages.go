package main

const (
	litmusBanner = `
██╗     ██╗████████╗███╗   ███╗██╗   ██╗███████╗
██║     ██║╚══██╔══╝████╗ ████║██║   ██║██╔════╝
██║     ██║   ██║   ██╔████╔██║██║   ██║███████╗
██║     ██║   ██║   ██║╚██╔╝██║██║   ██║╚════██║
███████╗██║   ██║   ██║ ╚═╝ ██║╚██████╔╝███████║
╚══════╝╚═╝   ╚═╝   ╚═╝     ╚═╝ ╚═════╝ ╚══════╝
════════════════════════════════════════════════
`
	longHelp = `Litmus helps you perform HTTP tests by providing a simple, TOML based DSL
for you to write endpoint tests with`

	cFlagUsage = "path to configuration folder"
	eFlagUsage = `environment variables: example baseurl=httpbin.org"`
	nFlagUsage = `name of specific test to run`
)
