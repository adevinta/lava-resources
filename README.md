# Lava Resources

Remote resources used by [Lava][lava].

Lava is an open source vulnerability scanner that makes it easy to run
security checks in your local and CI/CD environments.

## Managing Resources

This process is automated based on the following conventions:

  - Resources are grouped in directories.
  - All the resources in the same directory are versioned together.
  - Releases are created based on tags with the format `dir/semver`.
    Where `dir` is the directory containing the resources and `semver`
    is a semantic version.

To add new resources to the repository, follow these steps:

 1. If a new resource group is required, create a directory in the
    root of the repository.
    Otherwise, use an existing directory.
    Then, add the new resources inside.
 2. Create a pull request with the new changes.
 3. Once the pull request is approved and merged, push a tag with the
    format `dir/semver`.
    Where `dir` is the directory where the new resources are located
    and `semver` is a semantic version.

Resources can be updated and deleted following the same process.
For more details, run `go doc`.

## Contributing

**This project is in an early stage, we are not accepting external
contributions yet.**

To contribute, please read the [contribution
guidelines][contributing].


[lava]: https://github.com/adevinta/lava
[contributing]: /CONTRIBUTING.md
