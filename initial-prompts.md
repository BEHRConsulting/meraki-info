### Initial prompts
```
Description This project is a Golang application that collects Meraki network information.

The app should authenticate with Meraki cloud.

Use production authentication methods and libraries for Meraki, such as OAuth2, to ensure secure access to the Meraki account.

If the command route-tables is provided, output route tables. Routing information can be located on the security appliances, switches and switch stacks.

If the command licenses is provided, output license information. 

If the command down is provided, output all devices are are down. 

If the command alerting is provided, output all alerting devices.

If the command access is provided print a nice text output listing the organizations and networks available for the --apikey. Allow filtering by --org parameter

One of the commands is required, if not provided, display usage

In usage, show commands in alpha order.

If the option --org is provided use this as the organization. Can also be set with env variable MERAKI_ORG. Allow organization to be specified by name or id. This is not case sensitive

If the option --network is provided use this as the network. Can also be set with env variable MERAKI_NET. Allow network to be specified by name or id. This is not case sensitive. If --network is not provided, set --all

if the option --apikey is provided, use this as the api key. Can also be set with env variable MERAKI_APIKEY. Do not display a default for --apikey in usage, if set show "env MERAKI_APIKET is set".

If the option --output is provided, use this for the name of output file. If --output is "-" or not provided then send to stdout.

If the option --format is provided it can be text, xml, json, csv for the output file. Text is the default format.

If the option --all is provided generate a consolidated output. If the option --org is not specified, process all organizations. If option --net is not specified, process all networks. Consolidated output should include org name, org ID, network name, and network ID as required.

If the option --loglevel is provided, the app should set the logging level accordingly. The default logging level should be "error", but it can be set to "debug" or "info" based on the user's preference.

Create unit tests.

There should be no panics.

The app should handle errors gracefully, providing clear messages if something goes wrong, such as authentication failures, network issues, or file system errors. Error output should be sent to stderr.

The app should be efficient in terms of API calls to Dropbox, minimizing the number of requests made to avoid hitting rate limits.

The code should be well-structured and modular, making it easy to maintain and extend in the future.

The app should include comments and documentation to explain the functionality and usage. 

In usage, list commands in alpha order.

Sanitize all examples

<!-- need to do -->
Re-Generate a pretty REQUIREMENTS.md

```