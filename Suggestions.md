# Suggestions 

## Major

Currently test plugin building fails without any information on why. To ease up debugging validation of the provided test
file could really help.

## Minor 

The app does not support machines that got multiple Go versions installed. Before running the command to build the plugin
the version of Go should be requested.
Possible solution:
 ```go
cmd := exec.Command("bash", "-c", "cd "+serviceDir+" && "+runtime.Version()+" build -buildmode=plugin")
```  

`service.name` field as a reference for the test folder seems misleading to me and has nothing to do with the other fields 
in service. I think it would be better to reserve the name field for stuff like database names and let the folder containing
the tests have it's own field like `test_package`.

## Fixes

- example-configuration.yml and readme do not mention  bosh.ca field although it is implemented
- example-configuration.yml github field has wrong shifting.

- pulling the repo should always overwrite old files.


