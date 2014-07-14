Packages cannot have circular imports, therefore we have chosen this hierarchy for packages. In this hierarchy, packages may only import packes in lines above them.

[erasure] [logger] [network] [siacrypto] [siaencoding] [siafiles]
[quorum]
[quorum/script]
[participant]
[client] [server]
[client-cli]
[server-cli]
