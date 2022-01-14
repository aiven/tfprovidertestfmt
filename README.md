TFProviderTestFmt
=================

Features
============

Formats embedded HCL configurations, for example acceptance test manifests in terraform provider implementations. For an example PR please see: https://github.com/aiven/terraform-provider-aiven/pull/649/files

Limitations
============

Implementation Details:
* Traverses the AST of the go program and looks for literal nodes that are raw strings
* On encounters with valid HCL it formats the string properly and replaces the AST node

There is no guarantee that all diffs are valid with that method, if for example a manifest is put together from multiple strings or some other string gets parsed as valid HCL. In practice though thats unlikely enough to get some value out of this tool. It is recommended to inspect the diff before accepting it though. Eventually we will make it possible to add `ignore` comments to strings for better control on what to format. 

License
============
tfprovidertestfmt is licensed under the Apache license, version 2.0. Full license text is available in the [LICENSE](LICENSE) file.

Please note that the project explicitly does not require a CLA (Contributor License Agreement) from its contributors.

Contact
============
Bug reports and patches are very welcome, please post them as GitHub issues and pull requests at https://github.com/aiven/tfprovidertestfmt.
To report any possible vulnerabilities or other serious issues please see our [security](SECURITY.md) policy.
