# turbo-terraform
## Overview 
turbo-terraform leverages [Turbonomic's](https://turbonomic.com/) patented analysis engine to provide visibility and control across the entire stack in order to assure the performance, and leverage Terraform to apply the changes.

## Build
```console
make product
```
## Run
If using Terraform community version. tfpath is the argument for turbo-terraform to discover the existed terraform state files
```console
./_output/turbo-terraform -tfpath=<terraform_files_path>
```

If using Terraform enterprise version. tftoken is the argument for API token, org is the Terraform organization.
```console
./_output/turbo-terraform -org=<Name of the organization> -tftoken=<TF API token>
```

If using both, please provide all the arguments above.
```console
./_output/turbo-terraform -org=<Name of the organization> -tftoken=<TF API token> -tfpath=<terraform_files_path>
```