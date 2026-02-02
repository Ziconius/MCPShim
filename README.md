# MCPShim
A shim between MCP client/servers which logs requests, and allows for tampering via HTTP proxy tools such as ZAPProxy, and BurpeSuite.

MCPShim can be used launched without the client if you want to communicate with the MCP server via HTTP without the client querying the server

## Enabling the Shim
To enable the shim add add the MCPShim to the configuration.

OpenCode example
```json
{
    "$schema": "https://opencode.ai/config.json",
    "mcp": {
        "shim": {
        "type": "local",
        "command": [
            "/path/to/shim",
            "npx",
            "@playwright/mcp@latest"
            ]
        }
    }
}
```

### standalone

To run the shim without a client run the binary via the commandline adding the MCP args i.e.

```sh
mcpshim npx @playwright/mcp@latest
```

## Configuration

The configuration file for MCPShim can be found in the users home directory under:
```
$HOME/.config/mcpshim/config.json
```

Below is an example file
```json
{
    "logfile": "/tmp/mcpshim/mcp_shim.log",
    "intercept": {
        "enabled": true,
        "address": "http://127.0.0.1:8080"
    }
}
```

TODO:
 - [x] Create shim config file
 - [x] Pass through MCP server `args` to launch the target MCP
 - [x] Create webproxy functionality
    - [ ] Merge request response into single HTTP request not seperated
 - [ ] Create docs on usage
 - [ ] Update logging
    - [ ] Implement more config fiel locations
    - [ ] Cleanup logging messages
 - [ ] Clean up waitgroup issues.
 - [ ] HTTP interface for fuzzing
 - [ ] Implement golangci-lint