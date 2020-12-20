# Purpose
Manipulate configuration files through ENV variables, so that the image is immutable and the container can be fully configured through ENV.

## HowTo
```
telegraf --sample-config > test.conf
PFX_GLOBAL='global_tags|deeper|another_key="value"' PFX_CPU='inputs.cpu|percpus=false' PFX_SOME_NEW_KEY='newsection|subsection|key="value"' ./conf_update.go
```

## TODO
* [ ] create missing (sub-)sections
* [ ] make configurable (`prefix`, `removeComments`, `removeEmptyLines`)

## Issues to solve
* [x] `ENV` variables can not contain ".", keys and sections can
  * [x] try `PREFIX_ANYTHING=section|sub.section|keyname="value"`
  * docker can handle it: `docker run --rm -it -e VAR='foo|bar.name|key="ssdf"' busybox sh -c 'echo $VAR'`
  * also works with `--env-file`, only difference: no single quotes needed