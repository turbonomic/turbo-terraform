# turbo-terraform
## Overview 
turbo-terraform leverages [Turbonomic's](https://turbonomic.com/) patented analysis engine to provide visibility and control across the entire stack in order to assure the performance, and leverage Terraform to apply the changes.

## Build
```console
make product
```
## Run
tfpath is the argument for turbo-terraform to discover the existed terraform state files
```console
./_output/turbo-terraform -tfpath=/Users/Enlin/terraform-app
```
