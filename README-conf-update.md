# Purpose
Manipulate configuration files through ENV variables, so that the image is immutable and the container can be fully configured through ENV.

## HowTo
* Use `telegraf` to generate a default config file
* use `conf_update` with variables to update/edit the config and write back to the config file
* optionally remove all comments and whitespace
* Diff original + outcome
```
telegraf --sample-config > test.conf
export CONF_UPDATE=test.conf
export CONF_PREFIX=TLG
export CONF_STRIP_COMMENTS=false
export CONF_STRIP_EMPTYLINES=false
export TLG_GLOBAL='global_tags|deeper|another_key="value"'
export TLG_CPU='[inputs.cpu]|percpu=false' 
export TLG_SOME_NEW_KEY='[inputs.exec]|subsection|key="foo"'
./conf_update.go
telegraf --sample-config > orig.conf
diff -u orig.conf test.conf
```

## Issues to solve
* [x] `ENV` variables can not contain ".", keys and sections can
  * [x] try `PREFIX_ANYTHING=section|sub.section|keyname="value"`
  * docker can handle it: `docker run --rm -it -e VAR='foo|bar.name|key="ssdf"' busybox sh -c 'echo $VAR'`
  * also works with `--env-file`, only difference: no single quotes needed